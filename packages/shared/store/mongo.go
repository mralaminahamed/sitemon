// Package store persists check history in MongoDB.
package store

import (
	"context"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CheckStore struct {
	client *mongo.Client
	col    *mongo.Collection
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
	col := client.Database(db).Collection("checks")
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "url", Value: 1}, {Key: "timestamp", Value: -1}},
	})
	return &CheckStore{client: client, col: col}, nil
}

func (s *CheckStore) Save(ctx context.Context, r models.HealthResult) error {
	_, err := s.col.InsertOne(ctx, r)
	return err
}

func (s *CheckStore) History(ctx context.Context, url string, limit int64) ([]models.HealthResult, error) {
	filter := bson.M{}
	if url != "" {
		filter["url"] = url
	}
	if limit <= 0 {
		limit = 100
	}
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
