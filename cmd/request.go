package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mralaminahamed/sitemon/internal/http"
	"github.com/mralaminahamed/sitemon/internal/logger"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

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

var requestCmd = &cobra.Command{
	Use:   "request [url]",
	Short: "Send high-volume HTTP requests (GET, HEAD, POST, DELETE, PUT, PATCH)",
	Long: `Send high-volume HTTP requests for load testing.
Uses fast client with connection keep-alive for better performance.
HEAD is fastest as it doesn't download response body.

Reports detailed latency statistics including p50, p90, p95, and p99 percentiles.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := requestURL
		if len(args) > 0 {
			url = args[0]
		}
		if url == "" {
			return fmt.Errorf("URL is required")
		}

		client := http.NewFastClient(requestTimeout)
		if bypassCloudflare {
			client = http.NewFastClient(requestTimeout, http.WithCloudflareBypass())
		}
		ctx := context.Background()

		limiter := rate.NewLimiter(rate.Limit(rps), rps)

		g := errgroup.Group{}
		g.SetLimit(workers)

		var successCount int64
		var failureCount int64
		var totalLatency int64
		latencies := make([]time.Duration, 0, totalRequests)
		var latenciesMu sync.Mutex

		bar := progressbar.Default(int64(totalRequests))
		bar.Describe(fmt.Sprintf("Sending requests to %s", url))

		startTime := time.Now()

		var sentCount int64

		sendRequest := func() http.RequestResult {
			switch requestMethod {
			case "HEAD":
				return client.SendHead(url)
			case "POST":
				return client.SendPost(url, nil)
			case "PUT":
				return client.SendPut(url, nil)
			case "PATCH":
				return client.SendPatch(url, nil)
			case "DELETE":
				return client.SendDelete(url)
			case "OPTIONS":
				return client.SendOptions(url)
			default:
				return client.SendGet(url)
			}
		}

		for i := 0; i < totalRequests; i++ {
			if err := limiter.Wait(ctx); err != nil {
				break
			}

			g.Go(func() error {
				result := sendRequest()
				latency := result.ResponseTime

				atomic.AddInt64(&sentCount, 1)
				atomic.AddInt64(&totalLatency, latency.Milliseconds())

				latenciesMu.Lock()
				latencies = append(latencies, latency)
				latenciesMu.Unlock()

				bar.Add(1)

				if result.Error != nil {
					atomic.AddInt64(&failureCount, 1)
					logger.Log.Debug().Err(result.Error).Msg("request failed")
				} else if result.StatusCode >= 400 {
					atomic.AddInt64(&failureCount, 1)
					logger.Log.Debug().Int("status", result.StatusCode).Msg("request failed")
				} else {
					atomic.AddInt64(&successCount, 1)
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			logger.Log.Error().Err(err).Msg("request group failed")
		}

		elapsed := time.Since(startTime)
		actualTotal := atomic.LoadInt64(&sentCount)

		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		calcPercentile := func(p float64) time.Duration {
			if len(latencies) == 0 {
				return 0
			}
			index := int(float64(len(latencies)) * p)
			if index >= len(latencies) {
				index = len(latencies) - 1
			}
			if index < 0 {
				index = 0
			}
			return latencies[index]
		}

		avgLatencyMs := atomic.LoadInt64(&totalLatency) / actualTotal

		stats := RequestStats{
			TotalRequests:  int(actualTotal),
			Successful:     int(atomic.LoadInt64(&successCount)),
			Failed:         int(atomic.LoadInt64(&failureCount)),
			Duration:       elapsed,
			RequestsPerSec: float64(actualTotal) / elapsed.Seconds(),
			SuccessRate:    float64(atomic.LoadInt64(&successCount)) / float64(actualTotal) * 100,
			AvgLatency:     time.Duration(avgLatencyMs) * time.Millisecond,
			MinLatency:     latencies[0],
			MaxLatency:     latencies[len(latencies)-1],
			P50Latency:     calcPercentile(0.50),
			P90Latency:     calcPercentile(0.90),
			P95Latency:     calcPercentile(0.95),
			P99Latency:     calcPercentile(0.99),
			TargetURL:      url,
			Method:         requestMethod,
			Workers:        workers,
			RPS:            rps,
		}

		if jsonOutput {
			data, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal stats: %w", err)
			}
			fmt.Println(string(data))
		} else {
			fmt.Printf("\n========== Request Results ==========\n")
			fmt.Printf("Target URL:     %s\n", url)
			fmt.Printf("Method:        %s\n", requestMethod)
			fmt.Printf("Workers:       %d\n", workers)
			fmt.Printf("RPS Limit:     %d\n", rps)
			fmt.Printf("--------------------------------------\n")
			fmt.Printf("Total Requests: %d\n", actualTotal)
			fmt.Printf("Successful:     %d (%.2f%%)\n", stats.Successful, stats.SuccessRate)
			fmt.Printf("Failed:         %d\n", stats.Failed)
			fmt.Printf("Duration:       %v\n", elapsed)
			fmt.Printf("Requests/sec:   %.2f\n", stats.RequestsPerSec)
			fmt.Printf("--------------------------------------\n")
			fmt.Printf("Latency Stats (ms):\n")
			fmt.Printf("  Min:    %v\n", stats.MinLatency)
			fmt.Printf("  Avg:    %v\n", stats.AvgLatency)
			fmt.Printf("  Max:    %v\n", stats.MaxLatency)
			fmt.Printf("  P50:    %v\n", stats.P50Latency)
			fmt.Printf("  P90:    %v\n", stats.P90Latency)
			fmt.Printf("  P95:    %v\n", stats.P95Latency)
			fmt.Printf("  P99:    %v\n", stats.P99Latency)
			fmt.Printf("======================================\n")
		}

		if statsOutput != "" {
			data, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal stats: %w", err)
			}
			if err := os.WriteFile(statsOutput, data, 0644); err != nil {
				return fmt.Errorf("failed to write stats: %w", err)
			}
			fmt.Printf("\nStats saved to %s\n", statsOutput)
		}

		return nil
	},
}

var (
	requestURL        string
	workers           int
	rps               int
	totalRequests     int
	requestTimeout    time.Duration
	requestMethod     string
	requestJsonOutput bool
	statsOutput       string
)

func init() {
	rootCmd.AddCommand(requestCmd)

	requestCmd.Flags().StringVarP(&requestURL, "url", "u", "", "URL to request")
	requestCmd.Flags().IntVarP(&workers, "workers", "w", 10, "number of concurrent workers")
	requestCmd.Flags().IntVarP(&rps, "rps", "r", 100, "requests per second limit")
	requestCmd.Flags().IntVarP(&totalRequests, "count", "n", 1000, "total number of requests")
	requestCmd.Flags().DurationVar(&requestTimeout, "timeout", 10*time.Second, "request timeout")
	requestCmd.Flags().StringVarP(&requestMethod, "method", "m", "GET", "HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS)")
	requestCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	requestCmd.Flags().StringVarP(&statsOutput, "stats", "s", "", "save stats to file")
	requestCmd.Flags().BoolVar(&bypassCloudflare, "bypass-cloudflare", false, "use browser headers to bypass Cloudflare bot detection")
}
