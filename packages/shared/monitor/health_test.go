package monitor

import (
	"testing"
	"time"
)

func TestCalculateLatencyStats(t *testing.T) {
	var lat []time.Duration
	for i := 1; i <= 100; i++ {
		lat = append(lat, time.Duration(i)*time.Millisecond)
	}

	s := CalculateLatencyStats(lat)

	if s.Min != 1*time.Millisecond {
		t.Errorf("Min = %v, want 1ms", s.Min)
	}
	if s.Max != 100*time.Millisecond {
		t.Errorf("Max = %v, want 100ms", s.Max)
	}
	// Percentiles must be monotonic.
	if !(s.P50 <= s.P90 && s.P90 <= s.P95 && s.P95 <= s.P99) {
		t.Errorf("percentiles not monotonic: p50=%v p90=%v p95=%v p99=%v",
			s.P50, s.P90, s.P95, s.P99)
	}
	if s.Avg < s.Min || s.Avg > s.Max {
		t.Errorf("Avg %v outside [%v,%v]", s.Avg, s.Min, s.Max)
	}
}

func TestCalculateLatencyStatsEmpty(t *testing.T) {
	s := CalculateLatencyStats(nil)
	if s.Min != 0 || s.Max != 0 || s.P99 != 0 {
		t.Errorf("empty input should yield zero stats, got %+v", s)
	}
}
