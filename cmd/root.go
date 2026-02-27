package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/config"
	"github.com/mralaminahamed/portman/internal/logger"
)

var rootCmd = &cobra.Command{
	Use:   "portman",
	Short: "Portfolio HTTP Request Manager & Health Check Tool",
	Long: `Portman is a CLI tool for HTTP request management and portfolio site health monitoring.

Features:
  - Health checks for URLs with status monitoring
  - Continuous watch mode for ongoing monitoring
  - High-volume request testing with rate limiting
  - JSON report generation

Usage:
  portman check --url https://example.com
  portman check --url https://example.com --watch --interval 30s
  portman request --url https://example.com --workers 100 --rps 500
  portman report --output report.json

For more information, visit: https://github.com/mralaminahamed/portman`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Log.Error().Err(err).Msg("command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := config.InitConfig(); err != nil {
			logger.Log.Error().Err(err).Msg("failed to initialize config")
		}
	})

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path (default is config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
}
