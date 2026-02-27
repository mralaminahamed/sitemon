package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/monitor"
)

var checkCmd = &cobra.Command{
	Use:   "check [url...]",
	Short: "Check health of URLs",
	Long: `Check the health status of one or more URLs.
Returns status code, response time, and health status (UP/DOWN/WARNING).
Use --watch for continuous monitoring at specified intervals.
Supports multiple URLs for batch checking.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		urls := urlFlags
		if len(args) > 0 {
			urls = append(urls, args...)
		}

		if len(urls) == 0 {
			return fmt.Errorf("at least one URL is required")
		}

		for i, u := range urls {
			if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
				urls[i] = "https://" + u
			}
		}

		client := http.NewClient(timeout)
		healthChecker := monitor.NewHealthChecker(client)

		if watchFlag {
			return healthChecker.Watch(urls[0], interval, outputFile)
		}

		results := healthChecker.CheckMultiple(urls)

		if outputFile != "" {
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %w", err)
			}
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
			fmt.Printf("Results saved to %s\n", outputFile)
		}

		if jsonOutput {
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %w", err)
			}
			fmt.Println(string(data))
		} else {
			for _, r := range results {
				statusIcon := "✓"
				if r.Status != "UP" {
					statusIcon = "✗"
				}
				fmt.Printf("[%s] %s %s - %d (%s) in %v\n",
					r.Timestamp.Format("15:04:05"),
					statusIcon,
					r.URL,
					r.StatusCode,
					r.Status,
					r.ResponseTime,
				)
				if r.Error != "" {
					fmt.Printf("       Error: %s\n", r.Error)
				}
			}
		}

		healthChecker.PrintSummary()

		return nil
	},
}

var (
	urlFlags     []string
	watchFlag    bool
	jsonOutput   bool
	outputFile   string
	interval     time.Duration
	timeout      time.Duration
)

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringSliceVarP(&urlFlags, "url", "u", []string{}, "URL(s) to check (can be specified multiple times)")
	checkCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch mode for continuous monitoring")
	checkCmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "interval for watch mode")
	checkCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "request timeout")
	checkCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")
}
