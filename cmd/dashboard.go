package cmd

import (
	"fmt"
	"time"

	"github.com/mralaminahamed/sitemon/internal/tui"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Run interactive TUI dashboard",
	Long: `Run an interactive terminal dashboard for real-time health monitoring.
Shows live status of all monitored URLs with automatic refresh.

Examples:
  sitemon dashboard -u https://example.com
  sitemon dashboard -u https://example.com -u https://google.com --interval 5s
  sitemon dashboard -u https://example.com --timeout 30s
  sitemon dashboard -u https://codexpert.io --bypass-cloudflare`,
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

		dash := tui.NewDashboard(urls, dashboardInterval, dashboardTimeout, dashboardBypassCF)
		return dash.Start()
	},
}

var (
	dashboardURLs     []string
	dashboardInterval time.Duration
	dashboardTimeout  time.Duration
	dashboardBypassCF bool
)

func init() {
	rootCmd.AddCommand(dashboardCmd)

	dashboardCmd.Flags().StringSliceVarP(&dashboardURLs, "url", "u", []string{}, "URL(s) to monitor (supports multiple)")
	dashboardCmd.Flags().DurationVar(&dashboardInterval, "interval", 5*time.Second, "refresh interval (e.g., 5s, 1m)")
	dashboardCmd.Flags().DurationVar(&dashboardTimeout, "timeout", 10*time.Second, "request timeout")
	dashboardCmd.Flags().BoolVar(&dashboardBypassCF, "bypass-cloudflare", false, "use browser headers to bypass Cloudflare bot detection")
}
