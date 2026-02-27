package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/monitor"
)

var checkCmd = &cobra.Command{
	Use:   "check [url]",
	Short: "Check health of a URL",
	Long: `Check the health status of one or more URLs.
Returns status code, response time, and health status (UP/DOWN/WARNING).
Use --watch for continuous monitoring at specified intervals.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := urlFlag
		if len(args) > 0 {
			url = args[0]
		}
		if url == "" {
			return fmt.Errorf("URL is required")
		}

		client := http.NewClient(timeout)
		healthChecker := monitor.NewHealthChecker(client)

		if watchFlag {
			return healthChecker.Watch(url, interval)
		}

		result, err := healthChecker.Check(url)
		if err != nil {
			return err
		}

		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Status Code: %d\n", result.StatusCode)
		fmt.Printf("Response Time: %v\n", result.ResponseTime)
		fmt.Printf("Checked at: %v\n", result.Timestamp)
		return nil
	},
}

var (
	urlFlag     string
	watchFlag   bool
	interval    time.Duration
	timeout     time.Duration
)

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().StringVarP(&urlFlag, "url", "u", "", "URL to check")
	checkCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "watch mode for continuous monitoring")
	checkCmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "interval for watch mode")
	checkCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "request timeout")
}
