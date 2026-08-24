// Command notifier consumes check.result, detects UP<->down transitions, and
// sends webhook alerts (Slack/Discord/Telegram/generic). It also republishes
// an alert.raised audit event.
//
// Transition state is in-memory in Phase 2 (single instance); Phase 3 moves it
// to Redis so replicas share it. Config: WEBHOOK_URL (optional).
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/cache"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/notify"
)

// classify maps a status transition to an alert type. prev == "" (first
// sighting) yields no alert.
func classify(prev, status string) string {
	switch {
	case prev == "":
		return ""
	case prev == "UP" && status != "UP":
		return "down"
	case prev != "UP" && status == "UP":
		return "recovery"
	default:
		return ""
	}
}

type tracker struct {
	mu   sync.Mutex
	last map[string]string
}

func (t *tracker) decide(url, status string) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	prev := t.last[url]
	t.last[url] = status
	return classify(prev, status)
}

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info")})

	var notifier *notify.Notifier
	if hook := os.Getenv("WEBHOOK_URL"); hook != "" {
		notifier = notify.NewNotifier(hook)
	} else {
		logger.Log.Warn().Msg("notifier: WEBHOOK_URL unset — alerts will be logged only")
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
			logger.Log.Warn().Err(err).Msg("notifier: Redis unavailable, using in-memory dedup")
		} else {
			redis = r
			defer redis.Close()
		}
	}
	tr := &tracker{last: map[string]string{}}

	decide := func(ctx context.Context, url, status string) string {
		if redis != nil {
			prev, err := redis.Transition(ctx, url, status)
			if err != nil {
				logger.Log.Error().Err(err).Msg("redis transition")
				return ""
			}
			return classify(prev, status)
		}
		return tr.decide(url, status)
	}

	stop, err := b.Consume(context.Background(), "notifier", bus.SubjectCheckResult, func(data []byte) error {
		var result models.HealthResult
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
		alertType := decide(context.Background(), result.URL, result.Status)
		if alertType == "" {
			return nil
		}

		logger.Log.Info().Str("url", result.URL).Str("status", result.Status).
			Str("type", alertType).Msg("alert transition")

		if notifier != nil {
			if err := notifier.SendAlert(result.URL, result.Status, result.StatusCode, alertType); err != nil {
				logger.Log.Error().Err(err).Msg("webhook send failed")
			}
		}

		pctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = b.Publish(pctx, bus.SubjectAlertRaised, bus.AlertEvent{
			URL:        result.URL,
			Status:     result.Status,
			StatusCode: result.StatusCode,
			Type:       alertType,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		})
		return nil
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("notifier: consume results")
	}
	defer stop()

	go func() {
		addr := health.AddrFromEnv(":8083")
		logger.Log.Info().Str("addr", addr).Msg("notifier health listening")
		_ = health.Serve("notifier", addr)
	}()

	logger.Log.Info().Msg("notifier ready (consuming check.result)")
	waitForSignal()
	logger.Log.Info().Msg("notifier shutting down")
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
