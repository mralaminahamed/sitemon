package scheduler

import (
	"testing"
	"time"
)

func TestParseField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		min, max int
		want     []int
	}{
		{"wildcard", "*", 0, 5, []int{0, 1, 2, 3, 4, 5}},
		{"single", "5", 0, 59, []int{5}},
		{"list", "1,3,5", 0, 59, []int{1, 3, 5}},
		{"range", "1-3", 0, 59, []int{1, 2, 3}},
		{"step", "*/15", 0, 59, []int{0, 15, 30, 45}},
		{"range step", "0-20/10", 0, 59, []int{0, 10, 20}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseField(tt.field, tt.min, tt.max)
			if err != nil {
				t.Fatalf("parseField(%q) error: %v", tt.field, err)
			}
			if !equalInts(got, tt.want) {
				t.Errorf("parseField(%q) = %v, want %v", tt.field, got, tt.want)
			}
		})
	}
}

func TestCalculateNextRun(t *testing.T) {
	// From 10:02, "*/5 * * * *" should fire next at 10:05.
	from := time.Date(2026, 1, 1, 10, 2, 30, 0, time.UTC)
	next := calculateNextRun("*/5 * * * *", from)
	if next.Minute() != 5 || next.Hour() != 10 {
		t.Errorf("next run = %v, want minute 5 hour 10", next)
	}
	if !next.After(from) {
		t.Errorf("next run %v is not after %v", next, from)
	}
}

func TestParseCronRejectsShort(t *testing.T) {
	if _, err := ParseCron("* * *"); err == nil {
		t.Error("expected error for too-few fields")
	}
}

func TestHumanReadable(t *testing.T) {
	if got := HumanReadable("*/5 * * * *"); got == "" {
		t.Error("HumanReadable returned empty")
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
