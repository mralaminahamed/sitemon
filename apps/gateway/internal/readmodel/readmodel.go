// Package readmodel is the gateway's query side: latest status (Redis or
// in-memory) plus check history/stats (Mongo).
package readmodel

import (
	"context"

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
