package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/config"
	"github.com/mralaminahamed/portman/internal/logger"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "portman",
	Short: "Portfolio HTTP Request Manager & Health Check Tool",
	Long: `Portman is a CLI tool for HTTP request management and portfolio site health monitoring.

Features:
  - Health checks for URLs with status monitoring
  - Continuous watch mode for ongoing monitoring
  - High-volume request testing with rate limiting
  - Detailed latency statistics (p50, p90, p95, p99)
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

		cfg := config.GetConfig()
		logger.InitLogger(logger.LoggerOptions{
			Level:    cfg.LogLevel,
			LogFile:  cfg.LogFile,
		})
	})

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path (default is config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-file", "", "log file path")
	rootCmd.PersistentFlags().Bool("version", false, "show version information")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Portman version %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  date: %s\n", date)
			fmt.Printf("  built by: %s\n", builtBy)
		},
	})
}
