// Package store persists check history in MongoDB.
package store

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const opTimeout = 5 * time.Second

type CheckStore struct {
	client   *mongo.Client
	col      *mongo.Collection
	monitors *mongo.Collection
	alerts   *mongo.Collection
}

// Alert is a persisted alert-history record (the consumed alert.raised event).
type Alert struct {
	URL        string    `json:"url" bson:"url"`
	Status     string    `json:"status" bson:"status"`
	StatusCode int       `json:"status_code" bson:"status_code"`
	Type       string    `json:"type" bson:"type"`
	Timestamp  time.Time `json:"timestamp" bson:"timestamp"`
}

// Monitor is a tracked URL. _id is the URL, so adds are idempotent.
type Monitor struct {
	URL       string    `json:"url" bson:"_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Stats struct {
	URL           string  `json:"url"`
	Total         int64   `json:"total_checks"`
	Up            int64   `json:"successful"`
	Down          int64   `json:"failed"`
	UptimePercent float64 `json:"uptime_percentage"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

func NewCheckStore(ctx context.Context, uri, db string) (*CheckStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	db2 := client.Database(db)
	col := db2.Collection("checks")
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "url", Value: 1}, {Key: "timestamp", Value: -1}},
	})
	alerts := db2.Collection("alerts")
	_, _ = alerts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "url", Value: 1}, {Key: "timestamp", Value: -1}},
	})
	// Retention: TTL indexes cap unbounded growth. HISTORY_TTL_DAYS controls the
	// checks collection (default 30, 0 disables); alerts are kept 90 days.
	if days := envInt("HISTORY_TTL_DAYS", 30); days > 0 {
		_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(days * 86400)),
		})
	}
	_, _ = alerts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(90 * 86400),
	})
	return &CheckStore{client: client, col: col, monitors: db2.Collection("monitors"), alerts: alerts}, nil
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

// SaveAlert appends an alert to the history collection.
func (s *CheckStore) SaveAlert(ctx context.Context, a Alert) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	_, err := s.alerts.InsertOne(ctx, a)
	return err
}

// Alerts returns recent alerts, newest first, optionally filtered by url.
func (s *CheckStore) Alerts(ctx context.Context, url string, limit int64) ([]Alert, error) {
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	filter := bson.M{}
	if url != "" {
		filter["url"] = url
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cur, err := s.alerts.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []Alert{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CheckStore) AddMonitor(ctx context.Context, url string) (Monitor, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	now := time.Now().UTC()
	_, err := s.monitors.UpdateOne(ctx,
		bson.M{"_id": url},
		bson.M{"$setOnInsert": bson.M{"created_at": now}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return Monitor{}, err
	}
	var m Monitor
	if err := s.monitors.FindOne(ctx, bson.M{"_id": url}).Decode(&m); err != nil {
		return Monitor{URL: url, CreatedAt: now}, nil
	}
	return m, nil
}

func (s *CheckStore) ListMonitors(ctx context.Context) ([]Monitor, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	cur, err := s.monitors.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	out := []Monitor{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CheckStore) DeleteMonitor(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	_, err := s.monitors.DeleteOne(ctx, bson.M{"_id": url})
	return err
}

func (s *CheckStore) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	return s.client.Ping(ctx, nil)
}

func (s *CheckStore) Save(ctx context.Context, r models.HealthResult) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	// Deterministic _id makes writes idempotent: a JetStream redelivery of the
	// same result (at-least-once) upserts the same document instead of adding a
	// duplicate history row.
	id := fmt.Sprintf("%s|%d", r.URL, r.Timestamp.UnixNano())
	_, err := s.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$setOnInsert": r},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *CheckStore) History(ctx context.Context, url string, limit int64) ([]models.HealthResult, error) {
	filter := bson.M{}
	if url != "" {
		filter["url"] = url
	}
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cur, err := s.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []models.HealthResult{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CheckStore) Stats(ctx context.Context, url string) (Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	match := bson.M{}
	if url != "" {
		match["url"] = url
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "up", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{bson.D{{Key: "$eq", Value: bson.A{"$status", "UP"}}}, 1, 0}}}}}},
			{Key: "avg", Value: bson.D{{Key: "$avg", Value: "$responsetime"}}},
		}}},
	}
	cur, err := s.col.Aggregate(ctx, pipeline)
	if err != nil {
		return Stats{}, err
	}
	defer cur.Close(ctx)

	var rows []struct {
		Total int64   `bson:"total"`
		Up    int64   `bson:"up"`
		Avg   float64 `bson:"avg"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return Stats{}, err
	}

	st := Stats{URL: url}
	if len(rows) > 0 {
		st.Total = rows[0].Total
		st.Up = rows[0].Up
		st.Down = rows[0].Total - rows[0].Up
		st.AvgLatencyMs = rows[0].Avg / 1e6
		if st.Total > 0 {
			st.UptimePercent = float64(st.Up) / float64(st.Total) * 100
		}
	}
	return st, nil
}

func (s *CheckStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
