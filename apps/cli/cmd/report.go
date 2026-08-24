package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate health or load test report",
	Long: `Generate a comprehensive JSON report for health checks or load tests.
Report includes detailed statistics, check history, and performance metrics.
Can combine multiple check results and request statistics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		report := models.ReportData{
			GeneratedAt: time.Now().Format(time.RFC3339),
			ReportType:  reportType,
		}

		if checkResultsFile != "" {
			data, err := os.ReadFile(checkResultsFile)
			if err != nil {
				return fmt.Errorf("failed to read check results: %w", err)
			}

			var checks []models.HealthResult
			if err := json.Unmarshal(data, &checks); err != nil {
				return fmt.Errorf("failed to parse check results: %w", err)
			}

			report.Checks = checks

			successful := 0
			failed := 0
			for _, c := range checks {
				if c.Status == "UP" {
					successful++
				} else {
					failed++
				}
			}

			report.Summary.TotalChecks = len(checks)
			report.Summary.Successful = successful
			report.Summary.Failed = failed
			if len(checks) > 0 {
				report.Summary.UptimePercent = float64(successful) / float64(len(checks)) * 100
			}
		}

		if requestStatsFile != "" {
			data, err := os.ReadFile(requestStatsFile)
			if err != nil {
				return fmt.Errorf("failed to read request stats: %w", err)
			}

			var stats models.RequestStats
			if err := json.Unmarshal(data, &stats); err != nil {
				return fmt.Errorf("failed to parse request stats: %w", err)
			}

			report.RequestStats = stats

			report.Summary.TotalChecks = stats.TotalRequests
			report.Summary.Successful = stats.Successful
			report.Summary.Failed = stats.Failed
			report.Summary.UptimePercent = stats.SuccessRate
		}

		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}

		if outputPath != "" {
			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write report: %w", err)
			}
			fmt.Printf("Report saved to %s\n", outputPath)
		} else {
			fmt.Println(string(data))
		}

		return nil
	},
}

var (
	outputPath       string
	reportType       string
	checkResultsFile string
	requestStatsFile string
)

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path (default is stdout)")
	reportCmd.Flags().StringVarP(&reportType, "type", "t", "health", "report type (health, load-test, combined)")
	reportCmd.Flags().StringVarP(&checkResultsFile, "checks", "c", "", "JSON file with check results")
	reportCmd.Flags().StringVarP(&requestStatsFile, "stats", "s", "", "JSON file with request statistics")
}
