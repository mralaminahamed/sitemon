// Package readmodel is the gateway's query side: latest status (Redis or
// in-memory) plus check history/stats (Mongo).
package readmodel

import (
	"context"
	"errors"

	"github.com/mralaminahamed/sitemon/apps/gateway/internal/statuscache"
	"github.com/mralaminahamed/sitemon/packages/shared/cache"
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
		_ = rm.redis.StatusPut(ctx, res)
	} else {
		rm.mem.Put(res)
	}
	if rm.store != nil {
		_ = rm.store.Save(ctx, res)
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
	return rm.store.DeleteMonitor(ctx, url)
}
