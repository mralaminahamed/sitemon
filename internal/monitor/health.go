package monitor

import (
	"fmt"
	"time"

	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/logger"
)

type HealthResult struct {
	URL          string        `json:"url"`
	Status       string        `json:"status"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	Timestamp    time.Time     `json:"timestamp"`
}

type HealthChecker struct {
	client *http.Client
}

func NewHealthChecker(client *http.Client) *HealthChecker {
	return &HealthChecker{client: client}
}

func (h *HealthChecker) Check(url string) (*HealthResult, error) {
	start := time.Now()
	resp, err := h.client.Get(url)
	elapsed := time.Since(start)

	if err != nil {
		logger.Log.Error().Err(err).Str("url", url).Msg("health check failed")
		return &HealthResult{
			URL:          url,
			Status:       "DOWN",
			StatusCode:   0,
			ResponseTime: elapsed,
			Timestamp:    time.Now(),
		}, err
	}

	status := "UP"
	if resp.StatusCode() >= 500 {
		status = "DOWN"
	} else if resp.StatusCode() >= 400 {
		status = "WARNING"
	}

	logger.Log.Debug().
		Str("url", url).
		Int("status_code", resp.StatusCode()).
		Dur("latency", elapsed).
		Msg("health check completed")

	return &HealthResult{
		URL:          url,
		Status:       status,
		StatusCode:   resp.StatusCode(),
		ResponseTime: elapsed,
		Timestamp:    time.Now(),
	}, nil
}

func (h *HealthChecker) Watch(url string, interval time.Duration) error {
	fmt.Printf("Watching %s every %v (Ctrl+C to stop)\n", url, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		result, err := h.Check(url)
		if err != nil {
			fmt.Printf("[%s] %s - ERROR: %v\n", result.Timestamp.Format("15:04:05"), url, err)
		} else {
			fmt.Printf("[%s] %s - %d (%s) in %v\n",
				result.Timestamp.Format("15:04:05"),
				url,
				result.StatusCode,
				result.Status,
				result.ResponseTime,
			)
		}

		<-ticker.C
	}
}
