// Command ai analyzes check history and produces incident summaries.
//
// It answers the ai.analyze RPC: pull history from Mongo, compute anomaly
// metrics, and (when ANTHROPIC_API_KEY is set) add a Claude-written summary.
package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/apps/ai/internal/analyzer"
	"github.com/mralaminahamed/sitemon/apps/ai/internal/llm"
	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/store"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info")})
	ctx := context.Background()

	checks, err := store.NewCheckStore(ctx, mustEnv("MONGO_URI"), envOr("MONGO_DB", "sitemon"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("ai: cannot connect to Mongo")
	}
	defer checks.Close(ctx)

	summarizer := llm.New()

	b, err := bus.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("ai: cannot connect to NATS")
	}
	defer b.Close()

	stop, err := b.Reply(bus.SubjectAIAnalyze, "ai", func(data []byte) (any, error) {
		var req bus.RunAnalyzeRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return nil, err
		}
		limit := int64(req.Limit)
		if limit <= 0 {
			limit = 100
		}
		results, err := checks.History(context.Background(), req.URL, limit)
		if err != nil {
			return nil, err
		}
		a := analyzer.Analyze(req.URL, results)

		sctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if summary, err := summarizer.Summarize(sctx, a); err != nil {
			logger.Log.Warn().Err(err).Msg("ai: summarize failed")
		} else {
			a.Summary = summary
		}
		return a, nil
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("ai: subscribe analyze")
	}
	defer stop()

	go func() {
		addr := health.AddrFromEnv(":8090")
		logger.Log.Info().Str("addr", addr).Msg("ai health listening")
		_ = health.Serve("ai", addr, health.Check{Name: "nats", Ping: b.Ping}, health.Check{Name: "mongo", Ping: checks.Ping})
	}()

	logger.Log.Info().Msg("ai ready (ai.analyze)")
	waitForSignal()
	logger.Log.Info().Msg("ai shutting down")
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

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		logger.Log.Fatal().Str("key", key).Msg("ai: required env var not set")
	}
	return v
}
