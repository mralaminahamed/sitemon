// Package statuscache holds the latest health result per URL. It is the
// gateway's read model, fed by check.result events off the bus, and served by
// GET /api/status. In Phase 3 this moves to Redis so all gateway replicas share
// it; the interface stays the same.
package statuscache

import (
	"sort"
	"sync"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

type Cache struct {
	mu     sync.RWMutex
	latest map[string]models.HealthResult
}

func New() *Cache {
	return &Cache{latest: make(map[string]models.HealthResult)}
}

// Put records the latest result for its URL.
func (c *Cache) Put(r models.HealthResult) {
	c.mu.Lock()
	c.latest[r.URL] = r
	c.mu.Unlock()
}

// Snapshot returns all latest results, sorted by URL.
func (c *Cache) Snapshot() []models.HealthResult {
	c.mu.RLock()
	out := make([]models.HealthResult, 0, len(c.latest))
	for _, r := range c.latest {
		out = append(out, r)
	}
	c.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out
}
