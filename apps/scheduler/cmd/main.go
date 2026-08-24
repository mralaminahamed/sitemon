// Command scheduler emits check.job events on a cron schedule.
//
// Config via env:
//
//	SCHEDULE_CRON   cron expression (default "*/5 * * * *")
//	SCHEDULE_URLS   comma-separated URLs (required)
//	CHECK_TIMEOUT_MS, BYPASS_CLOUDFLARE
package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/scheduler"
)

func main() {
	logger.InitLogger(logger.LoggerOptions{Level: envOr("LOG_LEVEL", "info")})

	cronExpr := envOr("SCHEDULE_CRON", "*/5 * * * *")
	urls := splitURLs(os.Getenv("SCHEDULE_URLS"))
	if len(urls) == 0 {
		logger.Log.Fatal().Msg("scheduler: SCHEDULE_URLS is required")
	}
	timeoutMs, _ := strconv.Atoi(os.Getenv("CHECK_TIMEOUT_MS"))
	bypassCF := os.Getenv("BYPASS_CLOUDFLARE") == "true"

	sched, err := scheduler.ParseCron(cronExpr)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("scheduler: invalid cron")
	}

	b, err := bus.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("scheduler: cannot connect to NATS")
	}
	defer b.Close()
	if err := b.EnsureStream(context.Background()); err != nil {
		logger.Log.Fatal().Err(err).Msg("scheduler: ensure stream")
	}

	go func() {
		addr := health.AddrFromEnv(":8082")
		logger.Log.Info().Str("addr", addr).Msg("scheduler health listening")
		_ = health.Serve("scheduler", addr, health.Check{Name: "nats", Ping: b.Ping})
	}()

	logger.Log.Info().Str("cron", cronExpr).Strs("urls", urls).
		Str("human", scheduler.HumanReadable(cronExpr)).Msg("scheduler started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	nextRun := sched.Next()
	logger.Log.Info().Time("next", nextRun).Msg("first run scheduled")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			logger.Log.Info().Msg("scheduler shutting down")
			return
		case now := <-ticker.C:
			if now.Before(nextRun) {
				continue
			}
			dispatch(b, urls, timeoutMs, bypassCF)
			nextRun = sched.Next()
			logger.Log.Info().Time("next", nextRun).Msg("dispatched; next run scheduled")
		}
	}
}

func dispatch(b *bus.Bus, urls []string, timeoutMs int, bypassCF bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, u := range urls {
		job := bus.CheckJob{URL: u, TimeoutMs: timeoutMs, BypassCloudflare: bypassCF}
		if err := b.Publish(ctx, bus.SubjectCheckJob, job); err != nil {
			logger.Log.Error().Err(err).Str("url", u).Msg("publish check.job")
			continue
		}
		logger.Log.Debug().Str("url", u).Msg("check.job published")
	}
}

func splitURLs(csv string) []string {
	var out []string
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
			p = "https://" + p
		}
		out = append(out, p)
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
