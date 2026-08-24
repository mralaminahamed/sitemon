package storage

import (
	"path/filepath"
	"testing"
	"time"
)

// TestCheckRoundtrip guards the timestamp fix: a saved check must come back
// with a non-zero timestamp matching what was written (previously the fixed
// parse layout silently produced the zero time).
func TestCheckRoundtrip(t *testing.T) {
	db, err := NewDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)
	rec := CheckRecord{
		URL:          "https://example.com",
		Status:       "UP",
		StatusCode:   200,
		ResponseTime: 42,
		Timestamp:    now,
	}
	if err := db.SaveCheck(rec); err != nil {
		t.Fatalf("SaveCheck: %v", err)
	}

	checks, err := db.GetChecks(10, nil, nil)
	if err != nil {
		t.Fatalf("GetChecks: %v", err)
	}
	if len(checks) != 1 {
		t.Fatalf("got %d checks, want 1", len(checks))
	}
	got := checks[0]
	if got.Timestamp.IsZero() {
		t.Error("timestamp round-tripped to zero time")
	}
	if !got.Timestamp.Equal(now) {
		t.Errorf("timestamp = %v, want %v", got.Timestamp, now)
	}
	if got.URL != rec.URL || got.StatusCode != 200 {
		t.Errorf("record mismatch: %+v", got)
	}
}
