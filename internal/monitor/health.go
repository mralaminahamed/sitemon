package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/logger"
)

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

type HealthChecker struct {
	client  *http.Client
	results []HealthResult
}

func NewHealthChecker(client *http.Client) *HealthChecker {
	return &HealthChecker{
		client:  client,
		results: make([]HealthResult, 0),
	}
}

func (h *HealthChecker) Check(url string) (*HealthResult, error) {
	start := time.Now()
	result := h.client.SendGet(url)
	elapsed := time.Since(start)

	healthResult := HealthResult{
		URL:          url,
		ResponseTime: elapsed,
		Timestamp:    time.Now(),
	}

	if result.Error != nil {
		healthResult.Status = "DOWN"
		healthResult.StatusCode = 0
		healthResult.Error = result.Error.Error()
		logger.Log.Error().Err(result.Error).Str("url", url).Msg("health check failed")
	} else {
		healthResult.StatusCode = result.StatusCode
		switch {
		case result.StatusCode >= 500:
			healthResult.Status = "DOWN"
		case result.StatusCode >= 400:
			healthResult.Status = "WARNING"
		case result.StatusCode >= 300:
			healthResult.Status = "REDIRECT"
		default:
			healthResult.Status = "UP"
		}

		logger.Log.Debug().
			Str("url", url).
			Int("status_code", result.StatusCode).
			Dur("latency", elapsed).
			Msg("health check completed")
	}

	h.results = append(h.results, healthResult)
	return &healthResult, nil
}

func (h *HealthChecker) CheckMultiple(urls []string) []HealthResult {
	results := make([]HealthResult, 0, len(urls))
	for _, url := range urls {
		result, _ := h.Check(url)
		results = append(results, *result)
	}
	return results
}

func (h *HealthChecker) Watch(url string, interval time.Duration, output string) error {
	fmt.Printf("Watching %s every %v (Ctrl+C to stop)\n", url, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		result, err := h.Check(url)
		if err != nil {
			fmt.Printf("[%s] %s - ERROR: %v\n", 
				result.Timestamp.Format("15:04:05"), url, err)
		} else {
			statusIcon := "✓"
			if result.Status != "UP" {
				statusIcon = "✗"
			}
			fmt.Printf("[%s] %s %s - %d (%s) in %v\n",
				result.Timestamp.Format("15:04:05"),
				statusIcon,
				url,
				result.StatusCode,
				result.Status,
				result.ResponseTime,
			)
		}

		if output != "" {
			h.ExportResults(output)
		}

		<-ticker.C
	}
}

func (h *HealthChecker) GetStats() *HealthStats {
	if len(h.results) == 0 {
		return &HealthStats{}
	}

	var totalLatency int64
	minLatency := h.results[0].ResponseTime
	maxLatency := h.results[0].ResponseTime
	successful := 0
	failed := 0

	for _, r := range h.results {
		totalLatency += r.ResponseTime.Milliseconds()
		if r.ResponseTime < minLatency {
			minLatency = r.ResponseTime
		}
		if r.ResponseTime > maxLatency {
			maxLatency = r.ResponseTime
		}
		if r.Status == "UP" {
			successful++
		} else {
			failed++
		}
	}

	avgLatencyMs := totalLatency / int64(len(h.results))

	return &HealthStats{
		TotalChecks:     len(h.results),
		Successful:      successful,
		Failed:          failed,
		UptimePercent:   float64(successful) / float64(len(h.results)) * 100,
		AvgResponseTime: time.Duration(avgLatencyMs) * time.Millisecond,
		MinResponseTime: minLatency,
		MaxResponseTime: maxLatency,
	}
}

func (h *HealthChecker) GetResults() []HealthResult {
	return h.results
}

func (h *HealthChecker) GetLatest() *HealthResult {
	if len(h.results) == 0 {
		return nil
	}
	return &h.results[len(h.results)-1]
}

func (h *HealthChecker) ExportResults(filename string) error {
	data, err := json.MarshalIndent(h.results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (h *HealthChecker) ExportStats(filename string) error {
	stats := h.GetStats()
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (h *HealthChecker) PrintSummary() {
	stats := h.GetStats()
	fmt.Println("\n--- Summary ---")
	fmt.Printf("Total Checks: %d\n", stats.TotalChecks)
	fmt.Printf("Successful: %d\n", stats.Successful)
	fmt.Printf("Failed: %d\n", stats.Failed)
	fmt.Printf("Uptime: %.2f%%\n", stats.UptimePercent)
	if stats.TotalChecks > 0 {
		fmt.Printf("Avg Response Time: %v\n", stats.AvgResponseTime)
		fmt.Printf("Min Response Time: %v\n", stats.MinResponseTime)
		fmt.Printf("Max Response Time: %v\n", stats.MaxResponseTime)
	}
}

type LatencyStats struct {
	Min    time.Duration
	Max    time.Duration
	Avg    time.Duration
	P50    time.Duration
	P90    time.Duration
	P95    time.Duration
	P99    time.Duration
}

func CalculateLatencyStats(latencies []time.Duration) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var total time.Duration
	for _, l := range sorted {
		total += l
	}

	getPercentile := func(p float64) time.Duration {
		index := int(float64(len(sorted)) * p)
		if index >= len(sorted) {
			index = len(sorted) - 1
		}
		return sorted[index]
	}

	return LatencyStats{
		Min:  sorted[0],
		Max:  sorted[len(sorted)-1],
		Avg:  total / time.Duration(len(sorted)),
		P50:  getPercentile(0.50),
		P90:  getPercentile(0.90),
		P95:  getPercentile(0.95),
		P99:  getPercentile(0.99),
	}
}
