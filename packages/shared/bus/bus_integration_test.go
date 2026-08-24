//go:build integration

package bus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startNATS(t *testing.T) string {
	ctx := context.Background()
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "nats:2-alpine",
			Cmd:          []string{"-js", "-m", "8222"},
			ExposedPorts: []string{"4222/tcp"},
			WaitingFor:   wait.ForListeningPort("4222/tcp").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() { _ = c.Terminate(ctx) })
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "4222")
	return "nats://" + host + ":" + port.Port()
}

func TestEventRoundtrip(t *testing.T) {
	b, err := Connect(startNATS(t))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	ctx := context.Background()
	if err := b.EnsureStream(ctx); err != nil {
		t.Fatal(err)
	}

	got := make(chan CheckJob, 1)
	stop, err := b.Consume(ctx, "test-consumer", SubjectCheckJob, func(data []byte) error {
		var j CheckJob
		_ = json.Unmarshal(data, &j)
		got <- j
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	if err := b.Publish(ctx, SubjectCheckJob, CheckJob{URL: "https://x.com"}); err != nil {
		t.Fatal(err)
	}
	select {
	case j := <-got:
		if j.URL != "https://x.com" {
			t.Errorf("got %+v", j)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestRequestReplyErrorSurfaces(t *testing.T) {
	b, err := Connect(startNATS(t))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	stop, err := b.Reply("test.rpc", "q", func([]byte) (any, error) {
		return nil, errors.New("boom")
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	var out map[string]any
	if err := b.Request("test.rpc", map[string]string{}, &out, 5*time.Second); err == nil {
		t.Fatal("expected remote error to surface, got nil")
	}
}
