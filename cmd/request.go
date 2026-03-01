package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mralaminahamed/sitemon/internal/http"
	"github.com/mralaminahamed/sitemon/internal/logger"
	"github.com/mralaminahamed/sitemon/internal/models"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

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

		opts := []http.ClientOption{
			http.WithTimeout(requestTimeout),
		}

		if bypassCloudflare {
			opts = append(opts, http.WithCloudflareBypass())
		}

		for _, h := range requestHeaders {
			opts = append(opts, http.WithHeader(h[0], h[1]))
		}

		if requestProxy != "" {
			opts = append(opts, http.WithProxy(requestProxy))
		}

		if followRedirects {
			opts = append(opts, http.WithFollowRedirects(true))
		}

		if requestCookie != "" {
			opts = append(opts, http.WithCookie(requestCookie))
		}

		client := http.NewFastClient(requestTimeout, opts...)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g := errgroup.Group{}
		g.SetLimit(workers)

		var successCount int64
		var failureCount int64
		var totalLatency int64
		var totalResponseSize int64
		var minResponseSize int64 = -1
		var maxResponseSize int64
		latencies := make([]time.Duration, 0, totalRequests)
		var latenciesMu sync.Mutex
		statusCodes := make(map[int]int)
		var statusCodesMu sync.Mutex

		var progress *progressbar.ProgressBar
		if !quietMode {
			progress = progressbar.Default(int64(totalRequests))
			progress.Describe(fmt.Sprintf("Sending requests to %s", url))
		}

		startTime := time.Now()
		var sentCount int64

		sendRequest := func() http.RequestResult {
			switch requestMethod {
			case "HEAD":
				return client.SendHead(url)
			case "POST":
				return client.SendPost(url, requestBody)
			case "PUT":
				return client.SendPut(url, requestBody)
			case "PATCH":
				return client.SendPatch(url, requestBody)
			case "DELETE":
				return client.SendDelete(url)
			case "OPTIONS":
				return client.SendOptions(url)
			default:
				return client.SendGet(url)
			}
		}

		var rampUpDuration time.Duration
		if rampUp > 0 {
			rampUpDuration = rampUp
		}

		isDurationMode := requestDuration > 0
		actualTotal := int64(0)
		rampUpEndTime := startTime.Add(rampUpDuration)

		runTest := func() {
			maxIterations := totalRequests
			if isDurationMode {
				maxIterations = 999999999
			}

			for i := 0; i < maxIterations; i++ {
				if isDurationMode && time.Since(startTime) >= requestDuration {
					break
				}

				if rampUp > 0 && time.Now().Before(rampUpEndTime) {
					elapsedRamp := time.Since(startTime)
					currentRPS := int(float64(rps) * float64(elapsedRamp) / float64(rampUpDuration))
					if currentRPS < 1 {
						currentRPS = 1
					}
					limiter := rate.NewLimiter(rate.Limit(currentRPS), currentRPS)
					if err := limiter.Wait(ctx); err != nil {
						break
					}
				} else {
					limiter := rate.NewLimiter(rate.Limit(rps), rps)
					if err := limiter.Wait(ctx); err != nil {
						break
					}
				}

				g.Go(func() error {
					result := sendRequest()
					latency := result.ResponseTime

					atomic.AddInt64(&sentCount, 1)
					atomic.AddInt64(&totalLatency, latency.Milliseconds())
					atomic.AddInt64(&totalResponseSize, result.ResponseSize)

					if minResponseSize == -1 || result.ResponseSize < minResponseSize {
						atomic.StoreInt64(&minResponseSize, result.ResponseSize)
					}
					if result.ResponseSize > maxResponseSize {
						atomic.StoreInt64(&maxResponseSize, result.ResponseSize)
					}

					latenciesMu.Lock()
					latencies = append(latencies, latency)
					latenciesMu.Unlock()

					statusCodesMu.Lock()
					statusCodes[result.StatusCode]++
					statusCodesMu.Unlock()

					if progress != nil {
						progress.Add(1)
					}

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
		}

		runTest()

		if err := g.Wait(); err != nil && err != context.Canceled {
			logger.Log.Error().Err(err).Msg("request group failed")
		}

		elapsed := time.Since(startTime)
		actualTotal = atomic.LoadInt64(&sentCount)

		if actualTotal == 0 {
			return fmt.Errorf("no requests completed")
		}

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
		avgResponseSize := atomic.LoadInt64(&totalResponseSize) / actualTotal
		minRespSize := atomic.LoadInt64(&minResponseSize)
		maxRespSize := atomic.LoadInt64(&maxResponseSize)

		stats := models.RequestStats{
			TotalRequests:   int(actualTotal),
			Successful:      int(atomic.LoadInt64(&successCount)),
			Failed:          int(atomic.LoadInt64(&failureCount)),
			Duration:        elapsed,
			RequestsPerSec:  float64(actualTotal) / elapsed.Seconds(),
			SuccessRate:     float64(atomic.LoadInt64(&successCount)) / float64(actualTotal) * 100,
			AvgLatency:      time.Duration(avgLatencyMs) * time.Millisecond,
			MinLatency:      latencies[0],
			MaxLatency:      latencies[len(latencies)-1],
			P50Latency:      calcPercentile(0.50),
			P90Latency:      calcPercentile(0.90),
			P95Latency:      calcPercentile(0.95),
			P99Latency:      calcPercentile(0.99),
			TargetURL:       url,
			Method:          requestMethod,
			Workers:         workers,
			RPS:             rps,
			ResponseSizeAvg: avgResponseSize,
			ResponseSizeMin: minRespSize,
			ResponseSizeMax: maxRespSize,
			StatusCodes:     statusCodes,
			RampUpTime:      rampUpDuration,
			DurationMode:    isDurationMode,
		}

		if jsonOutput {
			data, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal stats: %w", err)
			}
			fmt.Println(string(data))
		} else if requestCsvOutput {
			writeCSV(stats)
		} else if requestPrometheus {
			writePrometheus(stats)
		} else if requestJUnit {
			writeJUnitXML(stats)
		} else {
			printStats(stats)
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

func printStats(stats models.RequestStats) {
	fmt.Printf("\n========== Request Results ==========\n")
	fmt.Printf("Target URL:     %s\n", stats.TargetURL)
	fmt.Printf("Method:        %s\n", stats.Method)
	fmt.Printf("Workers:       %d\n", stats.Workers)
	fmt.Printf("RPS Limit:     %d\n", stats.RPS)
	if stats.RampUpTime > 0 {
		fmt.Printf("Ramp-up:       %v\n", stats.RampUpTime)
	}
	fmt.Printf("--------------------------------------\n")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful:     %d (%.2f%%)\n", stats.Successful, stats.SuccessRate)
	fmt.Printf("Failed:         %d\n", stats.Failed)
	fmt.Printf("Duration:       %v\n", stats.Duration)
	fmt.Printf("Requests/sec:   %.2f\n", stats.RequestsPerSec)
	fmt.Printf("--------------------------------------\n")
	fmt.Printf("Latency Stats:\n")
	fmt.Printf("  Min:    %v\n", stats.MinLatency)
	fmt.Printf("  Avg:    %v\n", stats.AvgLatency)
	fmt.Printf("  Max:    %v\n", stats.MaxLatency)
	fmt.Printf("  P50:    %v\n", stats.P50Latency)
	fmt.Printf("  P90:    %v\n", stats.P90Latency)
	fmt.Printf("  P95:    %v\n", stats.P95Latency)
	fmt.Printf("  P99:    %v\n", stats.P99Latency)
	fmt.Printf("--------------------------------------\n")
	fmt.Printf("Response Size:\n")
	fmt.Printf("  Avg:    %d bytes\n", stats.ResponseSizeAvg)
	fmt.Printf("  Min:    %d bytes\n", stats.ResponseSizeMin)
	fmt.Printf("  Max:    %d bytes\n", stats.ResponseSizeMax)
	fmt.Printf("--------------------------------------\n")
	fmt.Printf("Status Codes:\n")
	codes := []int{}
	for code := range stats.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)
	for _, code := range codes {
		fmt.Printf("  %d: %d\n", code, stats.StatusCodes[code])
	}
	fmt.Printf("======================================\n")
}

func writeCSV(stats models.RequestStats) {
	filename := "sitemon_results.csv"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CSV: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Requests", strconv.Itoa(stats.TotalRequests)})
	writer.Write([]string{"Successful", strconv.Itoa(stats.Successful)})
	writer.Write([]string{"Failed", strconv.Itoa(stats.Failed)})
	writer.Write([]string{"Success Rate", fmt.Sprintf("%.2f", stats.SuccessRate) + "%"})
	writer.Write([]string{"Duration", stats.Duration.String()})
	writer.Write([]string{"Requests/sec", fmt.Sprintf("%.2f", stats.RequestsPerSec)})
	writer.Write([]string{"Avg Latency", stats.AvgLatency.String()})
	writer.Write([]string{"Min Latency", stats.MinLatency.String()})
	writer.Write([]string{"Max Latency", stats.MaxLatency.String()})
	writer.Write([]string{"P50 Latency", stats.P50Latency.String()})
	writer.Write([]string{"P90 Latency", stats.P90Latency.String()})
	writer.Write([]string{"P95 Latency", stats.P95Latency.String()})
	writer.Write([]string{"P99 Latency", stats.P99Latency.String()})
	writer.Write([]string{"Avg Response Size", strconv.FormatInt(stats.ResponseSizeAvg, 10)})

	fmt.Printf("CSV saved to %s\n", filename)
}

func writePrometheus(stats models.RequestStats) {
	fmt.Printf("# HELP sitemon_total_requests Total number of requests\n")
	fmt.Printf("# TYPE sitemon_total_requests counter\n")
	fmt.Printf("sitemon_total_requests %d\n", stats.TotalRequests)
	fmt.Printf("# HELP sitemon_successful Successful requests\n")
	fmt.Printf("# TYPE sitemon_successful counter\n")
	fmt.Printf("sitemon_successful %d\n", stats.Successful)
	fmt.Printf("# HELP sitemon_failed Failed requests\n")
	fmt.Printf("# TYPE sitemon_failed counter\n")
	fmt.Printf("sitemon_failed %d\n", stats.Failed)
	fmt.Printf("# HELP sitemon_requests_per_second Requests per second\n")
	fmt.Printf("# TYPE sitemon_requests_per_second gauge\n")
	fmt.Printf("sitemon_requests_per_second %.2f\n", stats.RequestsPerSec)
	fmt.Printf("# HELP sitemon_avg_latency_ms Average latency in milliseconds\n")
	fmt.Printf("# TYPE sitemon_avg_latency_ms gauge\n")
	fmt.Printf("sitemon_avg_latency_ms %.2f\n", float64(stats.AvgLatency.Milliseconds()))
	fmt.Printf("# HELP sitemon_p99_latency_ms P99 latency in milliseconds\n")
	fmt.Printf("# TYPE sitemon_p99_latency_ms gauge\n")
	fmt.Printf("sitemon_p99_latency_ms %.2f\n", float64(stats.P99Latency.Milliseconds()))
}

func writeJUnitXML(stats models.RequestStats) {
	filename := "sitemon_results.xml"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JUnit XML: %v\n", err)
		return
	}
	defer file.Close()

	failures := stats.Failed
	errors := 0
	if stats.SuccessRate < 95 {
		errors = stats.Failed
	}

	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="sitemon" tests="%d" failures="%d" errors="%d" time="%.3f">
  <testcase name="load-test" classname="sitemon" time="%.3f">
    %s
  </testcase>
  <system-out>
    Total: %d, Successful: %d, Failed: %d, RPS: %.2f, Avg Latency: %v, P99: %v
  </system-out>
</testsuite>
`, stats.TotalRequests, failures, errors, stats.Duration.Seconds(), stats.Duration.Seconds(),
		getFailureXML(failures),
		stats.TotalRequests, stats.Successful, stats.Failed, stats.RequestsPerSec, stats.AvgLatency, stats.P99Latency)

	fmt.Printf("JUnit XML saved to %s\n", filename)
}

func getFailureXML(failures int) string {
	if failures == 0 {
		return ""
	}
	return fmt.Sprintf(`<failure message="%d requests failed">%d requests failed</failure>`, failures, failures)
}

var (
	requestURL        string
	workers           int
	rps               int
	totalRequests     int
	requestDuration   time.Duration
	rampUp            time.Duration
	requestTimeout    time.Duration
	requestMethod     string
	requestBody       interface{}
	requestHeaders    [][]string
	requestCookie     string
	requestProxy      string
	requestJsonOutput bool
	requestCsvOutput  bool
	requestPrometheus bool
	requestJUnit      bool
	followRedirects   bool
	statsOutput       string
	quietMode         bool
)

func init() {
	rootCmd.AddCommand(requestCmd)

	requestCmd.Flags().StringVarP(&requestURL, "url", "u", "", "URL to request")
	requestCmd.Flags().IntVarP(&workers, "workers", "w", 10, "number of concurrent workers")
	requestCmd.Flags().IntVarP(&rps, "rps", "r", 100, "requests per second limit")
	requestCmd.Flags().IntVarP(&totalRequests, "count", "n", 1000, "total number of requests")
	requestCmd.Flags().DurationVar(&requestDuration, "duration", 0, "run for duration (e.g., 30s, 5m)")
	requestCmd.Flags().DurationVar(&rampUp, "ramp-up", 0, "ramp-up period (e.g., 10s)")
	requestCmd.Flags().DurationVar(&requestTimeout, "timeout", 10*time.Second, "request timeout")
	requestCmd.Flags().StringVarP(&requestMethod, "method", "m", "GET", "HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS)")
	requestCmd.Flags().StringVar(&requestBodyContent, "body", "", "request body for POST/PUT/PATCH (JSON string)")
	requestCmd.Flags().StringArrayVarP(&headerSlice, "header", "H", []string{}, "custom header (key:value)")
	requestCmd.Flags().StringVar(&requestCookie, "cookie", "", "cookie string (name=value)")
	requestCmd.Flags().StringVar(&requestProxy, "proxy", "", "HTTP proxy URL")
	requestCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	requestCmd.Flags().BoolVar(&requestCsvOutput, "csv", false, "output in CSV format")
	requestCmd.Flags().BoolVar(&requestPrometheus, "prometheus", false, "output in Prometheus format")
	requestCmd.Flags().BoolVar(&requestJUnit, "junit", false, "output in JUnit XML format")
	requestCmd.Flags().BoolVar(&followRedirects, "follow-redirects", false, "follow HTTP redirects")
	requestCmd.Flags().StringVarP(&statsOutput, "stats", "s", "", "save stats to file")
	requestCmd.Flags().BoolVar(&quietMode, "quiet", false, "suppress progress bar")
	requestCmd.Flags().BoolVar(&bypassCloudflare, "bypass-cloudflare", false, "use browser headers to bypass Cloudflare bot detection")
}

var requestBodyContent string
var headerSlice []string
