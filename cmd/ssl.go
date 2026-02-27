package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/mralaminahamed/portman/internal/ssl"
)

var sslCmd = &cobra.Command{
	Use:   "ssl [url...]",
	Short: "Check SSL certificate details",
	Long: `Check SSL certificate information for URLs.
Shows issuer, validity dates, days remaining, protocol, and cipher suite.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		urls := sslURLs
		if len(args) > 0 {
			urls = append(urls, args...)
		}

		if len(urls) == 0 {
			return fmt.Errorf("at least one URL is required")
		}

		for _, url := range urls {
			info, err := ssl.CheckCertificate(url, sslTimeout)
			if err != nil {
				fmt.Printf("✗ %s - Error: %v\n", url, err)
				continue
			}

			daysColor := "✓"
			daysStatus := "Valid"
			if info.DaysRemaining < 7 {
				daysColor = "✗ CRITICAL"
				daysStatus = "Critical"
			} else if info.DaysRemaining < 30 {
				daysColor = "⚠ WARNING"
				daysStatus = "Expiring Soon"
			}

			fmt.Printf("\n=== SSL Certificate: %s ===\n", url)
			fmt.Printf("Status:     %s %s\n", daysColor, daysStatus)
			fmt.Printf("Issuer:     %s\n", info.Issuer)
			fmt.Printf("Subject:    %s\n", info.Subject)
			fmt.Printf("Valid From: %s\n", info.ValidFrom.Format("2006-01-02 15:04:05"))
			fmt.Printf("Valid Until: %s\n", info.ValidUntil.Format("2006-01-02 15:04:05"))
			fmt.Printf("Days Left:  %d days\n", info.DaysRemaining)
			fmt.Printf("Protocol:   %s\n", info.Protocol)
			if info.CipherSuite != "" {
				fmt.Printf("Cipher:     0x%s\n", info.CipherSuite)
			}
		}

		return nil
	},
}

var (
	sslURLs    []string
	sslTimeout = 10 * time.Second
)

func init() {
	rootCmd.AddCommand(sslCmd)

	sslCmd.Flags().StringSliceVarP(&sslURLs, "url", "u", []string{}, "URL(s) to check")
	sslCmd.Flags().DurationVar(&sslTimeout, "timeout", 10*time.Second, "request timeout")
}
