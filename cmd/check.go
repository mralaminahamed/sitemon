package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/internal/history"
	"github.com/mralaminahamed/sitemon/internal/http"
	"github.com/mralaminahamed/sitemon/internal/monitor"
	"github.com/mralaminahamed/sitemon/internal/ssl"
	"github.com/mralaminahamed/sitemon/internal/validation"
	"github.com/mralaminahamed/sitemon/internal/webhook"
	"github.com/spf13/cobra"
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

		client := http.NewClient(checkTimeout)
		if bypassCloudflare {
			client = http.NewClient(checkTimeout, http.WithCloudflareBypass())
		}
		healthChecker := monitor.NewHealthChecker(client)

		if sslCheck {
			return runSSLCheck(urls, client)
		}

		if watchFlag {
			return runWatchMode(urls, client, healthChecker)
		}

		results := healthChecker.CheckMultiple(urls)

		if len(checkContains) > 0 || len(checkNotContains) > 0 {
			validator := validation.NewContentValidator(checkContains, checkNotContains, "")
			for i, r := range results {
				respClient := http.NewClient(checkTimeout)
				resp, err := respClient.Get(r.URL)
				if err == nil {
					valid, reason, _ := validator.Validate(resp)
					if !valid {
						results[i].Status = "WARNING"
						results[i].Error = fmt.Sprintf("validation failed: %s", reason)
					}
				}
			}
		}

		if saveToHistory {
			db, err := history.NewDatabase(checkDBPath)
			if err == nil {
				defer db.Close()
				for _, r := range results {
					db.SaveCheck(history.CheckRecord{
						URL:          r.URL,
						Status:       r.Status,
						StatusCode:   r.StatusCode,
						ResponseTime: r.ResponseTime.Milliseconds(),
						Timestamp:    r.Timestamp,
						Error:        r.Error,
					})
				}
			}
		}

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

func runSSLCheck(urls []string, client *http.Client) error {
	for _, url := range urls {
		info, err := ssl.CheckCertificate(url, checkTimeout)
		if err != nil {
			fmt.Printf("✗ %s - Error: %v\n", url, err)
			continue
		}

		daysColor := "✓"
		if info.DaysRemaining < 30 {
			daysColor = "✗"
		} else if info.DaysRemaining < 7 {
			daysColor = "⚠"
		}

		fmt.Printf("[SSL Check] %s\n", url)
		fmt.Printf("  Issuer: %s\n", info.Issuer)
		fmt.Printf("  Valid From: %s\n", info.ValidFrom.Format("2006-01-02"))
		fmt.Printf("  Valid Until: %s\n", info.ValidUntil.Format("2006-01-02"))
		fmt.Printf("  Days Remaining: %s %d days\n", daysColor, info.DaysRemaining)
		fmt.Printf("  Protocol: %s\n", info.Protocol)
		fmt.Println()
	}
	return nil
}

func runWatchMode(urls []string, client *http.Client, healthChecker *monitor.HealthChecker) error {
	var notifier *webhook.Notifier
	if checkWebhook != "" {
		notifier = webhook.NewNotifier(checkWebhook)
	}

	stopChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		close(stopChan)
	}()

	fmt.Printf("Watching %d URL(s) every %v (Ctrl+C to stop)\n", len(urls), checkInterval)
	if checkWebhook != "" {
		fmt.Printf("Alerts enabled: %s\n", checkWebhook)
	}
	if maxLatency > 0 {
		fmt.Printf("Max latency threshold: %v\n", maxLatency)
	}
	fmt.Println()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("\nStopped.")
			return nil
		case <-ticker.C:
			for _, url := range urls {
				result, err := healthChecker.Check(url)
				if err != nil {
					fmt.Printf("[%s] ✗ %s - ERROR: %v\n",
						result.Timestamp.Format("15:04:05"), url, err)
					if notifier != nil {
						notifier.SendAlert(url, "DOWN", 0, "down")
					}
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

					if notifier != nil && result.Status != "UP" {
						notifier.SendAlert(url, result.Status, result.StatusCode, "down")
					}
				}
			}

			if saveToHistory {
				db, err := history.NewDatabase(checkDBPath)
				if err == nil {
					defer db.Close()
					for _, result := range healthChecker.GetResults() {
						db.SaveCheck(history.CheckRecord{
							URL:          result.URL,
							Status:       result.Status,
							StatusCode:   result.StatusCode,
							ResponseTime: result.ResponseTime.Milliseconds(),
							Timestamp:    result.Timestamp,
							Error:        result.Error,
						})
					}
				}
			}
		}
	}
}

var (
	urlFlags         []string
	watchFlag        bool
	jsonOutput       bool
	outputFile       string
	checkInterval    time.Duration
	checkTimeout     time.Duration
	checkWebhook     string
	maxLatency       time.Duration
	checkContains    []string
	checkNotContains []string
	sslCheck         bool
	saveToHistory    bool
	checkDBPath      string
	bypassCloudflare bool
)

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringSliceVarP(&urlFlags, "url", "u", []string{}, "URL(s) to check (can be specified multiple times)")
	checkCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch mode for continuous monitoring")
	checkCmd.Flags().DurationVar(&checkInterval, "interval", 30*time.Second, "interval for watch mode")
	checkCmd.Flags().DurationVar(&checkTimeout, "timeout", 10*time.Second, "request timeout")
	checkCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")
	checkCmd.Flags().StringVar(&checkWebhook, "webhook", "", "webhook URL for alerts (Slack, Discord, Telegram)")
	checkCmd.Flags().DurationVar(&maxLatency, "max-latency", 0, "maximum allowed latency before alert")
	checkCmd.Flags().StringSliceVar(&checkContains, "contains", nil, "response must contain these strings")
	checkCmd.Flags().StringSliceVar(&checkNotContains, "not-contains", nil, "response must NOT contain these strings")
	checkCmd.Flags().BoolVar(&sslCheck, "check-ssl", false, "check SSL certificate")
	checkCmd.Flags().BoolVar(&saveToHistory, "save", false, "save results to history database")
	checkCmd.Flags().StringVar(&checkDBPath, "db", "", "path to history database")
	checkCmd.Flags().BoolVar(&bypassCloudflare, "bypass-cloudflare", false, "use browser headers to bypass Cloudflare bot detection")
}
