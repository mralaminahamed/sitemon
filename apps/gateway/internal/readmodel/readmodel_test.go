package readmodel

import (
	"context"
	"testing"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

func TestInMemoryFallback(t *testing.T) {
	rm := New(nil, nil) // no redis, no mongo
	ctx := context.Background()

	rm.PutResult(ctx, models.HealthResult{URL: "https://a.com", Status: "UP"})
	rm.PutResult(ctx, models.HealthResult{URL: "https://b.com", Status: "DOWN"})
	rm.PutResult(ctx, models.HealthResult{URL: "https://a.com", Status: "WARNING"})

	status, err := rm.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 2 {
		t.Fatalf("got %d urls, want 2", len(status))
	}
	// sorted by URL; a.com holds the latest (WARNING)
	if status[0].URL != "https://a.com" || status[0].Status != "WARNING" {
		t.Errorf("a.com = %+v, want latest WARNING", status[0])
	}

	if rm.HasHistory() {
		t.Error("HasHistory should be false without a store")
	}
	hist, _ := rm.History(ctx, "https://a.com", 10)
	if len(hist) != 0 {
		t.Errorf("history should be empty without a store, got %d", len(hist))
	}
	st, _ := rm.Stats(ctx, "https://a.com")
	if st.URL != "https://a.com" {
		t.Errorf("stats url = %q", st.URL)
	}
}
