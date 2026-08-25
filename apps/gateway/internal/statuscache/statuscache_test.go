package statuscache

import (
	"testing"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

func TestPutSnapshotDelete(t *testing.T) {
	c := New()
	c.Put(models.HealthResult{URL: "https://a.com", Status: "UP"})
	c.Put(models.HealthResult{URL: "https://b.com", Status: "DOWN"})
	if len(c.Snapshot()) != 2 {
		t.Fatalf("want 2 entries, got %d", len(c.Snapshot()))
	}
	c.Delete("https://a.com")
	snap := c.Snapshot()
	if len(snap) != 1 || snap[0].URL != "https://b.com" {
		t.Fatalf("delete failed, snapshot=%+v", snap)
	}
	c.Delete("https://missing.com") // no-op, must not panic
}
