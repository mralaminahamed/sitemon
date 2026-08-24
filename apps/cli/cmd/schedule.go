package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/http"
	"github.com/mralaminahamed/sitemon/packages/shared/monitor"
	"github.com/mralaminahamed/sitemon/packages/shared/notify"
	"github.com/mralaminahamed/sitemon/packages/shared/scheduler"
	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Run scheduled health checks",
	Long: `Run health checks on a cron schedule.
Examples:
  sitemon schedule --cron "*/5 * * * *" --url https://example.com
  sitemon schedule --cron "0 * * * *" --url https://example.com --webhook https://hooks.slack.com/...
  sitemon schedule --cron "*/5 * * * *" --url https://codexpert.io --bypass-cloudflare`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cronExpr == "" {
			return fmt.Errorf("cron expression is required (--cron)")
		}

		if len(scheduleURLs) == 0 {
			return fmt.Errorf("at least one URL is required (--url)")
		}

		sched, err := scheduler.ParseCron(cronExpr)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}

		client := http.NewClient(scheduleTimeout)
		if scheduleBypassCF {
			client = http.NewClient(scheduleTimeout, http.WithCloudflareBypass())
		}
		healthChecker := monitor.NewHealthChecker(client)

		var alerter *notify.AlertManager
		if scheduleWebhook != "" {
			alerter = notify.NewAlertManager(&notify.ThresholdConfig{
				WebhookURL: scheduleWebhook,
			})
		}

		fmt.Printf("Sitemon Scheduler\n")
		fmt.Printf("=================\n")
		fmt.Printf("Schedule: %s (%s)\n", cronExpr, scheduler.HumanReadable(cronExpr))
		fmt.Printf("URLs: %v\n", scheduleURLs)
		fmt.Printf("Timeout: %v\n", scheduleTimeout)
		if scheduleWebhook != "" {
			fmt.Printf("Webhook: %s\n", scheduleWebhook)
		}
		fmt.Printf("\nPress Ctrl+C to stop\n\n")

		stopChan := make(chan bool)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			close(stopChan)
		}()

		nextRun := sched.Next()
		fmt.Printf("Next run at: %s\n\n", nextRun.Format(time.RFC1123))

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				fmt.Println("\nScheduler stopped.")
				return nil
			case <-ticker.C:
				now := time.Now()
				if now.After(nextRun) || now.Equal(nextRun) {
					fmt.Printf("[%s] Running health checks...\n", now.Format("15:04:05"))

					for _, url := range scheduleURLs {
						result, err := healthChecker.Check(url)
						if err != nil {
							fmt.Printf("  ✗ %s - ERROR: %v\n", url, err)
							alerter.Evaluate(url, "DOWN", 0)
						} else {
							statusIcon := "✓"
							if result.Status != "UP" {
								statusIcon = "✗"
							}
							fmt.Printf("  %s %s - %d (%s) in %v\n",
								statusIcon, url, result.StatusCode, result.Status, result.ResponseTime)
							// Evaluate every run (incl. UP) so recovery fires;
							// dedup lives inside Evaluate.
							alerter.Evaluate(url, result.Status, result.StatusCode)
						}
					}

					nextRun = sched.Next()
					fmt.Printf("\nNext run at: %s\n\n", nextRun.Format(time.RFC1123))
				}
			}
		}
	},
}

var (
	cronExpr         string
	scheduleURLs     []string
	scheduleTimeout  time.Duration
	scheduleWebhook  string
	scheduleBypassCF bool
)

func init() {
	rootCmd.AddCommand(scheduleCmd)

	scheduleCmd.Flags().StringVar(&cronExpr, "cron", "", "cron expression (e.g., */5 * * * *)")
	scheduleCmd.Flags().StringSliceVarP(&scheduleURLs, "url", "u", []string{}, "URL(s) to check")
	scheduleCmd.Flags().DurationVar(&scheduleTimeout, "timeout", 10*time.Second, "request timeout")
	scheduleCmd.Flags().StringVar(&scheduleWebhook, "webhook", "", "webhook URL for alerts")
	scheduleCmd.Flags().BoolVar(&scheduleBypassCF, "bypass-cloudflare", false, "use browser headers to bypass Cloudflare bot detection")
}
