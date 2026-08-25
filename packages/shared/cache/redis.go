// Package cache is the Redis-backed shared state: latest status per URL and
// notifier transition tracking.
package cache

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/redis/go-redis/v9"
)

const (
	statusKey    = "sitemon:status"
	lastStatusNS = "sitemon:laststatus:"
	opTimeout    = 3 * time.Second
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

func (r *Redis) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return r.c.Ping(ctx).Err()
}

func (r *Redis) StatusPut(ctx context.Context, res models.HealthResult) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return r.c.HSet(ctx, statusKey, res.URL, data).Err()
}

func (r *Redis) StatusSnapshot(ctx context.Context) ([]models.HealthResult, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
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

// AlertState is the notifier's per-URL transition state, shared across replicas
// via Redis. Fails counts consecutive non-UP results; Alerted records whether a
// "down" alert has already fired for the current outage.
type AlertState struct {
	Fails   int  `json:"fails"`
	Alerted bool `json:"alerted"`
}

// LoadAlertState returns the stored state for url (zero value if none).
func (r *Redis) LoadAlertState(ctx context.Context, url string) (AlertState, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	raw, err := r.c.Get(ctx, lastStatusNS+url).Result()
	if err == redis.Nil {
		return AlertState{}, nil
	}
	if err != nil {
		return AlertState{}, err
	}
	var s AlertState
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return AlertState{}, nil // treat corrupt state as fresh
	}
	return s, nil
}

// SaveAlertState persists the notifier state for url.
func (r *Redis) SaveAlertState(ctx context.Context, url string, s AlertState) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return r.c.Set(ctx, lastStatusNS+url, data, 0).Err()
}

// StatusDelete removes a URL's cached latest status and notifier state, so a
// deleted monitor stops appearing in the status snapshot.
func (r *Redis) StatusDelete(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if err := r.c.HDel(ctx, statusKey, url).Err(); err != nil {
		return err
	}
	return r.c.Del(ctx, lastStatusNS+url).Err()
}

func (r *Redis) Close() error { return r.c.Close() }
