// Package readmodel is the gateway's query side: latest status (Redis or
// in-memory) plus check history/stats (Mongo).
package readmodel

import (
	"context"
	"errors"

	"github.com/mralaminahamed/sitemon/apps/gateway/internal/statuscache"
	"github.com/mralaminahamed/sitemon/packages/shared/cache"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/store"
)

type ReadModel struct {
	redis *cache.Redis
	mem   *statuscache.Cache
	store *store.CheckStore
}

func New(r *cache.Redis, st *store.CheckStore) *ReadModel {
	return &ReadModel{redis: r, mem: statuscache.New(), store: st}
}

func (rm *ReadModel) PutResult(ctx context.Context, res models.HealthResult) {
	if rm.redis != nil {
		if err := rm.redis.StatusPut(ctx, res); err != nil {
			logger.Log.Error().Err(err).Str("url", res.URL).Msg("readmodel: status cache write")
		}
	} else {
		rm.mem.Put(res)
	}
	if rm.store != nil {
		if err := rm.store.Save(ctx, res); err != nil {
			logger.Log.Error().Err(err).Str("url", res.URL).Msg("readmodel: history write")
		}
	}
}

func (rm *ReadModel) Status(ctx context.Context) ([]models.HealthResult, error) {
	if rm.redis != nil {
		return rm.redis.StatusSnapshot(ctx)
	}
	return rm.mem.Snapshot(), nil
}

func (rm *ReadModel) HasHistory() bool { return rm.store != nil }

func (rm *ReadModel) History(ctx context.Context, url string, limit int64) ([]models.HealthResult, error) {
	if rm.store == nil {
		return []models.HealthResult{}, nil
	}
	return rm.store.History(ctx, url, limit)
}

func (rm *ReadModel) Stats(ctx context.Context, url string) (store.Stats, error) {
	if rm.store == nil {
		return store.Stats{URL: url}, nil
	}
	return rm.store.Stats(ctx, url)
}

func (rm *ReadModel) Alerts(ctx context.Context, url string, limit int64) ([]store.Alert, error) {
	if rm.store == nil {
		return []store.Alert{}, nil
	}
	return rm.store.Alerts(ctx, url, limit)
}

// ErrNoStore is returned by monitor writes when Mongo is not configured.
var ErrNoStore = errors.New("monitors require MongoDB (MONGO_URI)")

func (rm *ReadModel) ListMonitors(ctx context.Context) ([]store.Monitor, error) {
	if rm.store == nil {
		return []store.Monitor{}, nil
	}
	return rm.store.ListMonitors(ctx)
}

func (rm *ReadModel) AddMonitor(ctx context.Context, url string) (store.Monitor, error) {
	if rm.store == nil {
		return store.Monitor{}, ErrNoStore
	}
	return rm.store.AddMonitor(ctx, url)
}

func (rm *ReadModel) DeleteMonitor(ctx context.Context, url string) error {
	if rm.store == nil {
		return ErrNoStore
	}
	if err := rm.store.DeleteMonitor(ctx, url); err != nil {
		return err
	}
	// Clear the read model too, or the deleted URL lingers in GET /api/status.
	if rm.redis != nil {
		if err := rm.redis.StatusDelete(ctx, url); err != nil {
			logger.Log.Error().Err(err).Str("url", url).Msg("readmodel: clear status on delete")
		}
	} else {
		rm.mem.Delete(url)
	}
	return nil
}
