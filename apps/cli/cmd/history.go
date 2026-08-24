package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/storage"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View health check history",
	Long: `View and manage health check history from the database.
Shows past check results with statistics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := storage.NewDatabase(historyDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		if exportHistory {
			data, err := db.ExportJSON()
			if err != nil {
				return fmt.Errorf("failed to export: %w", err)
			}
			if historyOutput != "" {
				return os.WriteFile(historyOutput, data, 0644) // FIX: was os.WriteFile
			}
			fmt.Println(string(data))
			return nil
		}

		if urlStats != "" {
			stats, err := db.GetStats(urlStats)
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			fmt.Printf("\n=== Statistics for %s ===\n", urlStats)
			fmt.Printf("Total Checks:    %d\n", int64(stats["total_checks"].(int64)))
			fmt.Printf("Successful:      %d\n", int64(stats["successful"].(int64)))
			fmt.Printf("Failed:          %d\n", int64(stats["failed"].(int64)))
			fmt.Printf("Uptime:          %.2f%%\n", stats["uptime_percentage"].(float64))
			fmt.Printf("Avg Response:    %.2fms\n", stats["avg_response_ms"].(float64))
			fmt.Println()
			return nil
		}

		fromTime, _ := time.Parse("2006-01-02", historyFrom)
		toTime, _ := time.Parse("2006-01-02", historyTo)

		var from, to *time.Time
		if historyFrom != "" {
			from = &fromTime
		}
		if historyTo != "" {
			to = &toTime
		}

		checks, err := db.GetChecks(historyLimit, from, to)
		if err != nil {
			return fmt.Errorf("failed to get checks: %w", err)
		}

		fmt.Printf("\n=== Health Check History ===\n\n")
		for _, c := range checks {
			statusIcon := "✓"
			if c.Status != "UP" {
				statusIcon = "✗"
			}
			fmt.Printf("[%s] %s %s - %d in %dms\n",
				c.Timestamp.Format("2006-01-02 15:04"),
				statusIcon,
				c.URL,
				c.StatusCode,
				c.ResponseTime,
			)
		}

		if len(checks) == 0 {
			fmt.Println("No records found.")
		}

		return nil
	},
}

var (
	historyDBPath string
	historyLimit  int
	historyFrom   string
	historyTo     string
	historyOutput string
	urlStats      string
	exportHistory bool
)

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().StringVar(&historyDBPath, "db", "", "path to history database")
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 50, "number of records to show")
	historyCmd.Flags().StringVar(&historyFrom, "from", "", "start date (YYYY-MM-DD)")
	historyCmd.Flags().StringVar(&historyTo, "to", "", "end date (YYYY-MM-DD)")
	historyCmd.Flags().StringVarP(&historyOutput, "output", "o", "", "output file for export")
	historyCmd.Flags().StringVar(&urlStats, "stats", "", "show statistics for URL")
	historyCmd.Flags().BoolVar(&exportHistory, "export", false, "export all history as JSON")
}
