// Command gateway is the sitemon REST/WebSocket entrypoint.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/mralaminahamed/sitemon/apps/gateway/internal/dto"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/readmodel"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/server"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/ws"
	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/cache"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/store"
)

func main() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logger.InitLogger(logger.LoggerOptions{Level: level})
	ctx := context.Background()

	var redis *cache.Redis
	if url := os.Getenv("REDIS_URL"); url != "" {
		if r, err := cache.NewRedis(ctx, url); err != nil {
			logger.Log.Warn().Err(err).Msg("gateway: Redis unavailable, using in-memory status")
		} else {
			redis = r
			defer redis.Close()
		}
	}

	var checks *store.CheckStore
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		if st, err := store.NewCheckStore(ctx, uri, envOr("MONGO_DB", "sitemon")); err != nil {
			logger.Log.Warn().Err(err).Msg("gateway: Mongo unavailable, history disabled")
		} else {
			checks = st
			defer checks.Close(ctx)
		}
	}

	rm := readmodel.New(redis, checks)

	hub := ws.NewHub(func() []byte {
		results, _ := rm.Status(context.Background())
		b, _ := json.Marshal(map[string]any{"type": "snapshot", "results": dto.FromHealthResults(results)})
		return b
	})

	var b *bus.Bus
	if url := os.Getenv("NATS_URL"); url != "" {
		if conn, err := bus.Connect(url); err != nil {
			logger.Log.Warn().Err(err).Msg("gateway: NATS unavailable, running in-process")
		} else {
			b = conn
			defer b.Close()
			subscribeResults(b, rm, hub)
			logger.Log.Info().Msg("gateway: connected to NATS (distributed mode)")
		}
	}

	svc := service.New(b)

	// Readiness reflects only the deps that are actually configured — the
	// gateway runs standalone with none of them.
	var ready []health.Check
	if redis != nil {
		ready = append(ready, health.Check{Name: "redis", Ping: redis.Ping})
	}
	if checks != nil {
		ready = append(ready, health.Check{Name: "mongo", Ping: checks.Ping})
	}
	if b != nil {
		ready = append(ready, health.Check{Name: "nats", Ping: b.Ping})
	}

	addr := health.AddrFromEnv(":8080")
	if err := server.Run(svc, rm, hub, addr, ready...); err != nil {
		logger.Log.Fatal().Err(err).Msg("gateway exited with error")
	}
}

func subscribeResults(b *bus.Bus, rm *readmodel.ReadModel, hub *ws.Hub) {
	if err := b.EnsureStream(context.Background()); err != nil {
		logger.Log.Warn().Err(err).Msg("gateway: ensure stream")
		return
	}
	_, err := b.Consume(context.Background(), "gateway-status", bus.SubjectCheckResult, func(data []byte) error {
		var r models.HealthResult
		if err := json.Unmarshal(data, &r); err != nil {
			return errors.Join(bus.ErrDrop, err)
		}
		rm.PutResult(context.Background(), r)
		if msg, err := json.Marshal(map[string]any{"type": "result", "result": dto.FromHealthResult(r)}); err == nil {
			hub.Broadcast(msg)
		}
		return nil
	})
	if err != nil {
		logger.Log.Warn().Err(err).Msg("gateway: subscribe check.result")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
