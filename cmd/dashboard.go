package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/tui"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Run interactive TUI dashboard",
	Long: `Run an interactive terminal dashboard for real-time health monitoring.
Shows live status of all monitored URLs with automatic refresh.

Examples:
  portman dashboard -u https://example.com
  portman dashboard -u https://example.com -u https://google.com --interval 5s
  portman dashboard -u https://example.com --timeout 30s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urls := dashboardURLs
		if len(args) > 0 {
			urls = args
		}

		if len(urls) == 0 {
			return fmt.Errorf("at least one URL is required (--url)")
		}

		for i, u := range urls {
			if len(u) > 4 && u[:4] != "http" {
				urls[i] = "https://" + u
			}
		}

		dash := tui.NewDashboard(urls, dashboardInterval, dashboardTimeout)
		return dash.Start()
	},
}

var (
	dashboardURLs       []string
	dashboardInterval   time.Duration
	dashboardTimeout    time.Duration
)

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.Flags().StringSliceVarP(&dashboardURLs, "url", "u", []string{}, "URL(s) to monitor (supports multiple)")
	dashboardCmd.Flags().DurationVar(&dashboardInterval, "interval", 5*time.Second, "refresh interval (e.g., 5s, 1m)")
	dashboardCmd.Flags().DurationVar(&dashboardTimeout, "timeout", 10*time.Second, "request timeout")
}
