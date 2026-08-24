package dto

import (
	"testing"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

func TestFromHealthResult(t *testing.T) {
	w := FromHealthResult(models.HealthResult{ResponseTime: 42 * time.Millisecond})
	if w.ResponseTimeMs != 42 {
		t.Errorf("ResponseTimeMs = %d, want 42", w.ResponseTimeMs)
	}
}

func TestFromRequestStats(t *testing.T) {
	w := FromRequestStats(models.RequestStats{
		AvgLatency: 150 * time.Millisecond,
		P99Latency: 900 * time.Millisecond,
		Duration:   2 * time.Second,
	})
	if w.AvgLatencyMs != 150 || w.P99LatencyMs != 900 || w.DurationMs != 2000 {
		t.Errorf("got avg=%d p99=%d dur=%d, want 150/900/2000", w.AvgLatencyMs, w.P99LatencyMs, w.DurationMs)
	}
}
