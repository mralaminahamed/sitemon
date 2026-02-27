package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/schollz/progressbar/v3"
	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

var requestCmd = &cobra.Command{
	Use:   "request [url]",
	Short: "Send high-volume HTTP requests (GET, HEAD, POST, DELETE)",
	Long: `Send high-volume HTTP requests for load testing.
Uses fast client with connection keep-alive for better performance.
HEAD is fastest as it doesn't download response body.`,
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
		ctx := context.Background()

		limiter := rate.NewLimiter(rate.Limit(rps), rps)

		g := errgroup.Group{}
		g.SetLimit(workers)

		var successCount int64
		var failureCount int64
		var totalLatency int64

		bar := progressbar.Default(int64(totalRequests))

		startTime := time.Now()

		var sentCount int64

		sendRequest := func() (int, error) {
			switch requestMethod {
			case "HEAD":
				return client.SendHead(url)
			case "POST":
				return client.SendPost(url, nil)
			case "DELETE":
				return client.SendDelete(url)
			default:
				return client.SendGet(url)
			}
		}

		for i := 0; i < totalRequests; i++ {
			if err := limiter.Wait(ctx); err != nil {
				break
			}

			g.Go(func() error {
			 reqStart := time.Now()
			 statusCode, err := sendRequest()
			 latency := time.Since(reqStart)

			 atomic.AddInt64(&sentCount, 1)
			 atomic.AddInt64(&totalLatency, latency.Milliseconds())

				bar.Add(1)

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					logger.Log.Debug().Err(err).Msg("request failed")
				} else if statusCode >= 400 {
					atomic.AddInt64(&failureCount, 1)
					logger.Log.Debug().Int("status", statusCode).Msg("request failed")
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

		fmt.Printf("\n--- Results ---\n")
		fmt.Printf("Total Requests: %d\n", actualTotal)
		fmt.Printf("Successful: %d\n", successCount)
		fmt.Printf("Failed: %d\n", failureCount)
		fmt.Printf("Duration: %v\n", elapsed)
		if actualTotal > 0 {
			fmt.Printf("Requests/sec: %.2f\n", float64(actualTotal)/elapsed.Seconds())
		}
		if successCount > 0 {
			fmt.Printf("Avg Latency: %.2fms\n", float64(totalLatency)/float64(successCount))
		}

		return nil
	},
}

var (
	requestURL      string
	workers         int
	rps             int
	totalRequests   int
	requestTimeout  time.Duration
	requestMethod   string
)

func init() {
	rootCmd.AddCommand(requestCmd)

	requestCmd.Flags().StringVarP(&requestURL, "url", "u", "", "URL to request")
	requestCmd.Flags().IntVarP(&workers, "workers", "w", 10, "number of concurrent workers")
	requestCmd.Flags().IntVarP(&rps, "rps", "r", 100, "requests per second limit")
	requestCmd.Flags().IntVarP(&totalRequests, "count", "n", 1000, "total number of requests")
	requestCmd.Flags().DurationVar(&requestTimeout, "timeout", 10*time.Second, "request timeout")
	requestCmd.Flags().StringVarP(&requestMethod, "method", "m", "GET", "HTTP method (GET, HEAD, POST, DELETE)")
}
