package models

import "time"

type RequestStats struct {
	TotalRequests  int           `json:"total_requests"`
	Successful     int           `json:"successful"`
	Failed         int           `json:"failed"`
	Duration       time.Duration `json:"duration"`
	RequestsPerSec float64       `json:"requests_per_sec"`
	SuccessRate    float64       `json:"success_rate"`
	AvgLatency     time.Duration `json:"avg_latency_ms"`
	MinLatency     time.Duration `json:"min_latency_ms"`
	MaxLatency     time.Duration `json:"max_latency_ms"`
	P50Latency     time.Duration `json:"p50_latency_ms"`
	P90Latency     time.Duration `json:"p90_latency_ms"`
	P95Latency     time.Duration `json:"p95_latency_ms"`
	P99Latency     time.Duration `json:"p99_latency_ms"`
	TargetURL      string        `json:"target_url"`
	Method         string        `json:"method"`
	Workers        int           `json:"workers"`
	RPS            int           `json:"rps_limit"`
}
