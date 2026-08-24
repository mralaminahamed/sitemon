package models

import "time"

type RequestStats struct {
	TotalRequests   int           `json:"total_requests"`
	Successful      int           `json:"successful"`
	Failed          int           `json:"failed"`
	Duration        time.Duration `json:"duration"`
	RequestsPerSec  float64       `json:"requests_per_sec"`
	SuccessRate     float64       `json:"success_rate"`
	AvgLatency      time.Duration `json:"avg_latency_ms"`
	MinLatency      time.Duration `json:"min_latency_ms"`
	MaxLatency      time.Duration `json:"max_latency_ms"`
	P50Latency      time.Duration `json:"p50_latency_ms"`
	P90Latency      time.Duration `json:"p90_latency_ms"`
	P95Latency      time.Duration `json:"p95_latency_ms"`
	P99Latency      time.Duration `json:"p99_latency_ms"`
	TargetURL       string        `json:"target_url"`
	Method          string        `json:"method"`
	Workers         int           `json:"workers"`
	RPS             int           `json:"rps_limit"`
	ResponseSizeAvg int64         `json:"response_size_avg_bytes"`
	ResponseSizeMin int64         `json:"response_size_min_bytes"`
	ResponseSizeMax int64         `json:"response_size_max_bytes"`
	StatusCodes     map[int]int   `json:"status_codes"`
	Errors          []string      `json:"errors,omitempty"`
	RampUpTime      time.Duration `json:"ramp_up_time,omitempty"`
	DurationMode    bool          `json:"duration_mode,omitempty"`
}

type PrometheusMetrics struct {
	TotalRequests  float64 `prometheus:"total_requests"`
	Successful     float64 `prometheus:"successful"`
	Failed         float64 `prometheus:"failed"`
	RequestsPerSec float64 `prometheus:"requests_per_sec"`
	SuccessRate    float64 `prometheus:"success_rate"`
	AvgLatencyMs   float64 `prometheus:"avg_latency_ms"`
	MinLatencyMs   float64 `prometheus:"min_latency_ms"`
	MaxLatencyMs   float64 `prometheus:"max_latency_ms"`
	P50LatencyMs   float64 `prometheus:"p50_latency_ms"`
	P90LatencyMs   float64 `prometheus:"p90_latency_ms"`
	P95LatencyMs   float64 `prometheus:"p95_latency_ms"`
	P99LatencyMs   float64 `prometheus:"p99_latency_ms"`
}
