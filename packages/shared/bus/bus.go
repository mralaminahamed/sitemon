// Package bus is the sitemon event/RPC layer over NATS.
//
//   - JetStream (durable, load-balanced) carries events: check.job,
//     check.result, alert.raised.
//   - Core NATS request-reply carries synchronous RPC: checker.run,
//     checker.loadtest.
//
// Services depend on this package rather than nats.go directly, so the wire
// format and stream config live in one place.
package bus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ErrDrop marks a message as non-retryable so Consume terminates it instead of
// redelivering (e.g. malformed payloads). Wrap it: errors.Join(bus.ErrDrop, err).
var ErrDrop = errors.New("drop message")

// replyEnvelope wraps request-reply payloads so remote errors surface to the
// caller instead of being silently unmarshaled into a zero value.
type replyEnvelope struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

// Subjects.
const (
	StreamName = "SITEMON"

	SubjectCheckJob    = "check.job"    // scheduler -> checker
	SubjectCheckResult = "check.result" // checker   -> gateway, notifier
	SubjectAlertRaised = "alert.raised" // notifier  -> (audit / gateway)

	SubjectCheckerRun      = "checker.run"      // gateway -> checker (req-reply)
	SubjectCheckerLoadTest = "checker.loadtest" // gateway -> checker (req-reply)
	SubjectAIAnalyze       = "ai.analyze"       // gateway -> ai (req-reply)

	// QueueCheckers load-balances work across checker replicas.
	QueueCheckers = "checkers"
)

// Bus wraps a NATS connection plus its JetStream context.
type Bus struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// Connect dials NATS and initializes JetStream.
func Connect(url string) (*Bus, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream init: %w", err)
	}
	return &Bus{nc: nc, js: js}, nil
}

// EnsureStream creates or updates the SITEMON stream covering all event
// subjects. Idempotent — safe to call from every service at startup.
func (b *Bus) EnsureStream(ctx context.Context) error {
	_, err := b.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      StreamName,
		Subjects:  []string{"check.>", "alert.>"},
		Retention: jetstream.LimitsPolicy,
		MaxAge:    24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("ensure stream: %w", err)
	}
	return nil
}

// Publish JSON-encodes v and publishes it to subject via JetStream.
func (b *Bus) Publish(ctx context.Context, subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := b.js.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}

// Consume starts a durable, explicit-ack JetStream consumer filtered to
// filterSubject. Handlers sharing a durable name load-balance. It returns a
// stop func. The handler's error only controls ack: nil acks, non-nil naks
// for redelivery.
func (b *Bus) Consume(ctx context.Context, durable, filterSubject string, handler func([]byte) error) (func(), error) {
	cons, err := b.js.CreateOrUpdateConsumer(ctx, StreamName, jetstream.ConsumerConfig{
		Durable:       durable,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		BackOff:       []time.Duration{time.Second, 5 * time.Second, 15 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer %s: %w", durable, err)
	}
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		if err := handler(msg.Data()); err != nil {
			// Non-retryable (bad data) is terminated so it can't hot-loop;
			// everything else is redelivered up to MaxDeliver with backoff.
			if errors.Is(err, ErrDrop) {
				_ = msg.Term()
			} else {
				_ = msg.Nak()
			}
			return
		}
		_ = msg.Ack()
	})
	if err != nil {
		return nil, fmt.Errorf("consume %s: %w", durable, err)
	}
	return cc.Stop, nil
}

// Request sends a core NATS request (JSON in, JSON out) and decodes the reply
// into out.
func (b *Bus) Request(subject string, in, out any, timeout time.Duration) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	msg, err := b.nc.Request(subject, data, timeout)
	if err != nil {
		return fmt.Errorf("request %s: %w", subject, err)
	}
	var env replyEnvelope
	if err := json.Unmarshal(msg.Data, &env); err != nil {
		return fmt.Errorf("decode reply %s: %w", subject, err)
	}
	if env.Error != "" {
		return fmt.Errorf("%s: %s", subject, env.Error)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// Reply registers a queue-subscribed responder. Handlers sharing queue
// load-balance requests. It returns an unsubscribe func.
func (b *Bus) Reply(subject, queue string, handler func([]byte) (any, error)) (func(), error) {
	respond := func(m *nats.Msg, env replyEnvelope) {
		data, _ := json.Marshal(env)
		_ = m.Respond(data)
	}
	sub, err := b.nc.QueueSubscribe(subject, queue, func(m *nats.Msg) {
		out, err := handler(m.Data)
		if err != nil {
			respond(m, replyEnvelope{Error: err.Error()})
			return
		}
		raw, err := json.Marshal(out)
		if err != nil {
			respond(m, replyEnvelope{Error: err.Error()})
			return
		}
		respond(m, replyEnvelope{Data: raw})
	})
	if err != nil {
		return nil, fmt.Errorf("reply %s: %w", subject, err)
	}
	return func() { _ = sub.Unsubscribe() }, nil
}

// Ping reports whether the NATS connection is healthy.
func (b *Bus) Ping(context.Context) error {
	if b.nc == nil || !b.nc.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}

// Close drains and closes the connection.
func (b *Bus) Close() {
	if b.nc != nil {
		_ = b.nc.Drain()
	}
}
