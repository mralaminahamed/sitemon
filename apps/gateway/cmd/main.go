// Command gateway is the sitemon REST/WebSocket entrypoint.
//
// Phase 2: it connects to NATS (best-effort). With a bus it dispatches checks
// and load tests to the checker service and maintains a live status cache from
// check.result events. Without a bus it runs the engines in-process, so it
// still works standalone. See ARCHITECTURE.md.
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/mralaminahamed/sitemon/apps/gateway/internal/server"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/statuscache"
	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	"github.com/mralaminahamed/sitemon/packages/shared/health"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

func main() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logger.InitLogger(logger.LoggerOptions{Level: level})

	cache := statuscache.New()

	// Connect to NATS if configured; degrade to in-process mode on failure so
	// the gateway always starts.
	var b *bus.Bus
	if url := os.Getenv("NATS_URL"); url != "" {
		conn, err := bus.Connect(url)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("gateway: NATS unavailable, running in-process")
		} else {
			b = conn
			defer b.Close()
			subscribeResults(b, cache)
			logger.Log.Info().Msg("gateway: connected to NATS (distributed mode)")
		}
	} else {
		logger.Log.Info().Msg("gateway: NATS_URL unset, running in-process")
	}

	svc := service.New(b)

	addr := health.AddrFromEnv(":8080")
	if err := server.Run(svc, cache, addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("gateway exited with error")
	}
}

// subscribeResults feeds check.result events into the status cache.
func subscribeResults(b *bus.Bus, cache *statuscache.Cache) {
	if err := b.EnsureStream(context.Background()); err != nil {
		logger.Log.Warn().Err(err).Msg("gateway: ensure stream")
		return
	}
	_, err := b.Consume(context.Background(), "gateway-status", bus.SubjectCheckResult, func(data []byte) error {
		var r models.HealthResult
		if err := json.Unmarshal(data, &r); err != nil {
			return err
		}
		cache.Put(r)
		return nil
	})
	if err != nil {
		logger.Log.Warn().Err(err).Msg("gateway: subscribe check.result")
	}
}
