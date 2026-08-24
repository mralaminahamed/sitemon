package models

import "time"

type HealthResult struct {
	URL          string        `json:"url"`
	Status       string        `json:"status"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time_ms"`
	Timestamp    time.Time     `json:"timestamp"`
	Error        string        `json:"error,omitempty"`
}

type HealthStats struct {
	TotalChecks     int           `json:"total_checks"`
	Successful      int           `json:"successful"`
	Failed          int           `json:"failed"`
	UptimePercent   float64       `json:"uptime_percentage"`
	AvgResponseTime time.Duration `json:"avg_response_time_ms"`
	MinResponseTime time.Duration `json:"min_response_time_ms"`
	MaxResponseTime time.Duration `json:"max_response_time_ms"`
}

type LatencyStats struct {
	Min time.Duration
	Max time.Duration
	Avg time.Duration
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
}
