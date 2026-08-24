package dto

import (
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

// Wire types present latency as real milliseconds. The domain models keep
// time.Duration (nanoseconds) for internal use — Redis JSON and Mongo BSON — so
// conversion happens only at the API boundary here.

type HealthResult struct {
	URL            string    `json:"url"`
	Status         string    `json:"status"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	Timestamp      time.Time `json:"timestamp"`
	Error          string    `json:"error,omitempty"`
}

func FromHealthResult(r models.HealthResult) HealthResult {
	return HealthResult{
		URL:            r.URL,
		Status:         r.Status,
		StatusCode:     r.StatusCode,
		ResponseTimeMs: r.ResponseTime.Milliseconds(),
		Timestamp:      r.Timestamp,
		Error:          r.Error,
	}
}

func FromHealthResults(rs []models.HealthResult) []HealthResult {
	out := make([]HealthResult, len(rs))
	for i, r := range rs {
		out[i] = FromHealthResult(r)
	}
	return out
}

type RequestStats struct {
	TotalRequests   int         `json:"total_requests"`
	Successful      int         `json:"successful"`
	Failed          int         `json:"failed"`
	DurationMs      int64       `json:"duration_ms"`
	RequestsPerSec  float64     `json:"requests_per_sec"`
	SuccessRate     float64     `json:"success_rate"`
	AvgLatencyMs    int64       `json:"avg_latency_ms"`
	MinLatencyMs    int64       `json:"min_latency_ms"`
	MaxLatencyMs    int64       `json:"max_latency_ms"`
	P50LatencyMs    int64       `json:"p50_latency_ms"`
	P90LatencyMs    int64       `json:"p90_latency_ms"`
	P95LatencyMs    int64       `json:"p95_latency_ms"`
	P99LatencyMs    int64       `json:"p99_latency_ms"`
	TargetURL       string      `json:"target_url"`
	Method          string      `json:"method"`
	Workers         int         `json:"workers"`
	RPS             int         `json:"rps_limit"`
	ResponseSizeAvg int64       `json:"response_size_avg_bytes"`
	ResponseSizeMin int64       `json:"response_size_min_bytes"`
	ResponseSizeMax int64       `json:"response_size_max_bytes"`
	StatusCodes     map[int]int `json:"status_codes"`
	RampUpTimeMs    int64       `json:"ramp_up_time_ms,omitempty"`
	DurationMode    bool        `json:"duration_mode,omitempty"`
}

func FromRequestStats(s models.RequestStats) RequestStats {
	return RequestStats{
		TotalRequests:   s.TotalRequests,
		Successful:      s.Successful,
		Failed:          s.Failed,
		DurationMs:      s.Duration.Milliseconds(),
		RequestsPerSec:  s.RequestsPerSec,
		SuccessRate:     s.SuccessRate,
		AvgLatencyMs:    s.AvgLatency.Milliseconds(),
		MinLatencyMs:    s.MinLatency.Milliseconds(),
		MaxLatencyMs:    s.MaxLatency.Milliseconds(),
		P50LatencyMs:    s.P50Latency.Milliseconds(),
		P90LatencyMs:    s.P90Latency.Milliseconds(),
		P95LatencyMs:    s.P95Latency.Milliseconds(),
		P99LatencyMs:    s.P99Latency.Milliseconds(),
		TargetURL:       s.TargetURL,
		Method:          s.Method,
		Workers:         s.Workers,
		RPS:             s.RPS,
		ResponseSizeAvg: s.ResponseSizeAvg,
		ResponseSizeMin: s.ResponseSizeMin,
		ResponseSizeMax: s.ResponseSizeMax,
		StatusCodes:     s.StatusCodes,
		RampUpTimeMs:    s.RampUpTime.Milliseconds(),
		DurationMode:    s.DurationMode,
	}
}
