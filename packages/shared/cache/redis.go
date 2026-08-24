// Package cache is the Redis-backed shared state: latest status per URL and
// notifier transition tracking.
package cache

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/redis/go-redis/v9"
)

const (
	statusKey    = "sitemon:status"
	lastStatusNS = "sitemon:laststatus:"
)

type Redis struct {
	c *redis.Client
}

func NewRedis(ctx context.Context, url string) (*Redis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	c := redis.NewClient(opt)
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{c: c}, nil
}

func (r *Redis) StatusPut(ctx context.Context, res models.HealthResult) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return r.c.HSet(ctx, statusKey, res.URL, data).Err()
}

func (r *Redis) StatusSnapshot(ctx context.Context) ([]models.HealthResult, error) {
	m, err := r.c.HGetAll(ctx, statusKey).Result()
	if err != nil {
		return nil, err
	}
	out := make([]models.HealthResult, 0, len(m))
	for _, v := range m {
		var res models.HealthResult
		if json.Unmarshal([]byte(v), &res) == nil {
			out = append(out, res)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out, nil
}

// Transition atomically stores the new status and returns the previous one
// ("" if none), letting the caller detect up<->down changes across replicas.
func (r *Redis) Transition(ctx context.Context, url, status string) (string, error) {
	prev, err := r.c.GetSet(ctx, lastStatusNS+url, status).Result()
	if err == redis.Nil {
		return "", nil
	}
	return prev, err
}

func (r *Redis) Close() error { return r.c.Close() }
