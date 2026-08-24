// Command checker runs health/load work on behalf of the platform.
//
// Phase 2: it both answers synchronous RPC (checker.run, checker.loadtest via
// core NATS request-reply) and consumes scheduled check.job events from
// JetStream, publishing check.result for each. See ARCHITECTURE.md.
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/apps/checker/internal/engine"
	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/loadtest"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info")})

	b, err := bus.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("checker: cannot connect to NATS")
	}
	defer b.Close()

	ctx := context.Background()
	if err := b.EnsureStream(ctx); err != nil {
		logger.Log.Fatal().Err(err).Msg("checker: ensure stream")
	}

	// --- Synchronous RPC: single check ---
	stopRun, err := b.Reply(bus.SubjectCheckerRun, bus.QueueCheckers, func(data []byte) (any, error) {
		var req bus.RunCheckRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return engine.Check(req.URL, time.Duration(req.TimeoutMs)*time.Millisecond, req.BypassCloudflare), nil
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("checker: subscribe run")
	}
	defer stopRun()

	// --- Synchronous RPC: load test ---
	stopLoad, err := b.Reply(bus.SubjectCheckerLoadTest, bus.QueueCheckers, func(data []byte) (any, error) {
		var req bus.RunLoadTestRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		return engine.LoadTest(context.Background(), loadtest.Options{
			URL:      req.URL,
			Method:   req.Method,
			Workers:  req.Workers,
			RPS:      req.RPS,
			Count:    req.Count,
			Duration: time.Duration(req.DurationMs) * time.Millisecond,
			Timeout:  time.Duration(req.TimeoutMs) * time.Millisecond,
			Headers:  req.Headers,
			BypassCF: req.BypassCloudflare,
		})
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("checker: subscribe loadtest")
	}
	defer stopLoad()

	// --- Scheduled jobs: check.job -> check.result ---
	stopJobs, err := b.Consume(ctx, "checker-jobs", bus.SubjectCheckJob, func(data []byte) error {
		var job bus.CheckJob
		if err := json.Unmarshal(data, &job); err != nil {
			return err // nak; malformed will redeliver but that's acceptable here
		}
		result := engine.Check(job.URL, time.Duration(job.TimeoutMs)*time.Millisecond, job.BypassCloudflare)

		pctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := b.Publish(pctx, bus.SubjectCheckResult, result); err != nil {
			logger.Log.Error().Err(err).Str("url", job.URL).Msg("publish check.result")
			return err
		}
		logger.Log.Debug().Str("url", job.URL).Str("status", result.Status).Msg("job checked")
		return nil
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("checker: consume jobs")
	}
	defer stopJobs()

	go func() {
		addr := health.AddrFromEnv(":8081")
		logger.Log.Info().Str("addr", addr).Msg("checker health listening")
		_ = health.Serve("checker", addr)
	}()

	logger.Log.Info().Msg("checker ready (rpc + job consumer)")
	waitForSignal()
	logger.Log.Info().Msg("checker shutting down")
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
