// Command notifier consumes check.result, detects UP<->down transitions, and
// sends webhook alerts (Slack/Discord/Telegram/generic). It persists an alert
// history record and republishes an alert.raised audit event.
//
// Delivery is reliable: transition state advances only after the webhook send
// succeeds, so a transient webhook failure Naks the message for redelivery
// instead of silently losing the alert. ALERT_FAILURE_THRESHOLD (default 2)
// requires that many consecutive non-UP results before a "down" fires, damping
// flaps. State is shared across replicas via Redis (in-memory fallback).
// Config: WEBHOOK_URL, TELEGRAM_CHAT_ID, REDIS_URL, MONGO_URI.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/cache"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/notify"
	"github.com/mralaminahamed/sitemon/packages/shared/store"
)

// alertStateStore is the per-URL transition state, backed by Redis or memory.
type alertStateStore interface {
	LoadAlertState(context.Context, string) (cache.AlertState, error)
	SaveAlertState(context.Context, string, cache.AlertState) error
}

// step advances the alert state for one check result and returns the next state
// plus the alert to fire ("", "down", or "recovery"). A "down" fires only once
// per outage, and only after threshold consecutive non-UP results.
func step(s cache.AlertState, status string, threshold int) (cache.AlertState, string) {
	if status == "UP" {
		if s.Alerted {
			return cache.AlertState{}, "recovery"
		}
		return cache.AlertState{}, ""
	}
	s.Fails++
	if !s.Alerted && s.Fails >= threshold {
		s.Alerted = true
		return s, "down"
	}
	return s, ""
}

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info")})

	threshold := 2
	if v, err := strconv.Atoi(os.Getenv("ALERT_FAILURE_THRESHOLD")); err == nil && v > 0 {
		threshold = v
	}

	var notifier *notify.Notifier
	if hook := os.Getenv("WEBHOOK_URL"); hook != "" {
		notifier = notify.NewNotifier(hook)
	} else {
		logger.Log.Warn().Msg("notifier: WEBHOOK_URL unset — alerts will be recorded only")
	}

	b, err := bus.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("notifier: cannot connect to NATS")
	}
	defer b.Close()
	if err := b.EnsureStream(context.Background()); err != nil {
		logger.Log.Fatal().Err(err).Msg("notifier: ensure stream")
	}

	var redis *cache.Redis
	if url := os.Getenv("REDIS_URL"); url != "" {
		if r, err := cache.NewRedis(context.Background(), url); err != nil {
			logger.Log.Warn().Err(err).Msg("notifier: Redis unavailable, using in-memory state")
		} else {
			redis = r
			defer redis.Close()
		}
	}
	var states alertStateStore = newMemState()
	if redis != nil {
		states = redis
	}

	var db *store.CheckStore
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		if st, err := store.NewCheckStore(context.Background(), uri, envOr("MONGO_DB", "sitemon")); err != nil {
			logger.Log.Warn().Err(err).Msg("notifier: Mongo unavailable, alert history disabled")
		} else {
			db = st
			defer db.Close(context.Background())
		}
	}

	var (
		wg    sync.WaitGroup
		locks keyedMutex
	)

	handle := func(data []byte) error {
		var result models.HealthResult
		if err := json.Unmarshal(data, &result); err != nil {
			return errors.Join(bus.ErrDrop, err)
		}
		unlock := locks.lock(result.URL)
		defer unlock()

		ctx := context.Background()
		prev, err := states.LoadAlertState(ctx, result.URL)
		if err != nil {
			return err // Nak; do not advance state
		}
		next, alertType := step(prev, result.Status, threshold)

		// Send first: on failure return the error so the message is redelivered
		// and re-sent, instead of advancing state and losing the alert.
		if alertType != "" && notifier != nil {
			if err := notifier.SendAlert(result.URL, result.Status, result.StatusCode, alertType); err != nil {
				logger.Log.Error().Err(err).Str("url", result.URL).Msg("webhook send failed; will retry")
				return err
			}
		}
		if err := states.SaveAlertState(ctx, result.URL, next); err != nil {
			return err
		}
		if alertType == "" {
			return nil
		}

		logger.Log.Info().Str("url", result.URL).Str("status", result.Status).
			Str("type", alertType).Msg("alert fired")

		// History + audit event are best-effort: the webhook already went out,
		// so do not resend on a persistence hiccup.
		now := time.Now().UTC()
		if db != nil {
			if err := db.SaveAlert(ctx, store.Alert{
				URL: result.URL, Status: result.Status, StatusCode: result.StatusCode,
				Type: alertType, Timestamp: now,
			}); err != nil {
				logger.Log.Error().Err(err).Msg("persist alert history")
			}
		}
		pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = b.Publish(pctx, bus.SubjectAlertRaised, bus.AlertEvent{
			URL: result.URL, Status: result.Status, StatusCode: result.StatusCode,
			Type: alertType, Timestamp: now.Format(time.RFC3339),
		})
		return nil
	}

	stop, err := b.Consume(context.Background(), "notifier", bus.SubjectCheckResult, func(data []byte) error {
		wg.Add(1)
		defer wg.Done()
		return handle(data)
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("notifier: consume results")
	}

	checks := []health.Check{{Name: "nats", Ping: b.Ping}}
	if redis != nil {
		checks = append(checks, health.Check{Name: "redis", Ping: redis.Ping})
	}
	go func() {
		addr := health.AddrFromEnv(":8083")
		logger.Log.Info().Str("addr", addr).Msg("notifier health listening")
		_ = health.Serve("notifier", addr, checks...)
	}()

	logger.Log.Info().Int("threshold", threshold).Msg("notifier ready (consuming check.result)")
	waitForSignal()
	logger.Log.Info().Msg("notifier shutting down")
	stop()
	drain(&wg, 10*time.Second)
}

// drain waits for in-flight handlers to finish, up to timeout.
func drain(wg *sync.WaitGroup, timeout time.Duration) {
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(timeout):
		logger.Log.Warn().Msg("notifier: drain timed out with handlers in flight")
	}
}

// keyedMutex serializes processing per URL so concurrent results for the same
// monitor can't interleave their read-modify-write of the alert state.
type keyedMutex struct {
	mu sync.Mutex
	m  map[string]*sync.Mutex
}

func (k *keyedMutex) lock(key string) func() {
	k.mu.Lock()
	if k.m == nil {
		k.m = map[string]*sync.Mutex{}
	}
	mu, ok := k.m[key]
	if !ok {
		mu = &sync.Mutex{}
		k.m[key] = mu
	}
	k.mu.Unlock()
	mu.Lock()
	return mu.Unlock
}

type memState struct {
	mu sync.Mutex
	m  map[string]cache.AlertState
}

func newMemState() *memState { return &memState{m: map[string]cache.AlertState{}} }

func (s *memState) LoadAlertState(_ context.Context, url string) (cache.AlertState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[url], nil
}

func (s *memState) SaveAlertState(_ context.Context, url string, st cache.AlertState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[url] = st
	return nil
}

func waitForSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
