//go:build integration

package store

import (
	"context"
	"testing"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startMongo(t *testing.T) string {
	ctx := context.Background()
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "mongo:7",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start mongo: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "27017")
	return "mongodb://" + host + ":" + port.Port()
}

func TestSaveIdempotentAndQuery(t *testing.T) {
	ctx := context.Background()
	st, err := NewCheckStore(ctx, startMongo(t), "sitemon")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close(ctx)

	r := models.HealthResult{
		URL: "https://example.com", Status: "UP", StatusCode: 200,
		ResponseTime: 40 * time.Millisecond, Timestamp: time.Now().UTC(),
	}
	for i := 0; i < 3; i++ { // simulate redelivery
		if err := st.Save(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	h, err := st.History(ctx, "https://example.com", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 1 {
		t.Fatalf("history len = %d, want 1 (idempotent)", len(h))
	}

	r2 := r
	r2.Status = "DOWN"
	r2.Timestamp = r.Timestamp.Add(time.Second)
	if err := st.Save(ctx, r2); err != nil {
		t.Fatal(err)
	}

	stats, err := st.Stats(ctx, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 2 || stats.Up != 1 || stats.Down != 1 {
		t.Errorf("stats = %+v, want total 2 up 1 down 1", stats)
	}
}
