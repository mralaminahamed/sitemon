package tui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/monitor"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	tableRowStyle = lipgloss.NewStyle()
)

type Dashboard struct {
	urls       []string
	results    map[string]*monitor.HealthResult
	resultsMu  sync.RWMutex
	interval   time.Duration
	client     *http.Client
	stopChan   chan bool
	width      int
	height     int
	stats      DashboardStats
	lastUpdate time.Time
}

type DashboardStats struct {
	TotalChecks   int
	SuccessCount  int
	FailureCount  int
	UptimePercent float64
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
}

func NewDashboard(urls []string, interval time.Duration, timeout time.Duration) *Dashboard {
	return &Dashboard{
		urls:      urls,
		results:   make(map[string]*monitor.HealthResult),
		interval:  interval,
		client:    http.NewClient(timeout),
		stopChan:  make(chan bool),
		stats:     DashboardStats{},
		lastUpdate: time.Now(),
	}
}

func (d *Dashboard) Start() error {
	healthChecker := monitor.NewHealthChecker(d.client)

	d.checkURLs(healthChecker)

	go d.monitorLoop(healthChecker)

	fmt.Println(d.Render())
	fmt.Println(disabledStyle.Render("\n  Press Ctrl+C to exit  "))

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			return nil
		case <-ticker.C:
			fmt.Print("\033[2J")
			fmt.Print("\033[H")
			fmt.Println(d.Render())
			fmt.Println(disabledStyle.Render("\n  Press Ctrl+C to exit  "))
		}
	}
}

func (d *Dashboard) monitorLoop(healthChecker *monitor.HealthChecker) {
	for {
		select {
		case <-d.stopChan:
			return
		case <-time.After(d.interval):
			d.checkURLs(healthChecker)
		}
	}
}

func (d *Dashboard) checkURLs(healthChecker *monitor.HealthChecker) {
	var wg sync.WaitGroup

	currentResults := make(map[string]*monitor.HealthResult)
	var totalLatency int64
	var minLatency int64 = -1
	var maxLatency int64
	successCount := 0
	failureCount := 0

	for _, url := range d.urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result, err := healthChecker.Check(u)
			if err == nil {
				currentResults[u] = result
			}
		}(url)
	}

	wg.Wait()

	d.resultsMu.Lock()
	d.results = currentResults
	d.resultsMu.Unlock()

	for _, result := range currentResults {
		latencyMs := result.ResponseTime.Milliseconds()
		totalLatency += latencyMs

		if minLatency == -1 || latencyMs < minLatency {
			minLatency = latencyMs
		}
		if latencyMs > maxLatency {
			maxLatency = latencyMs
		}

		if result.Status == "UP" {
			successCount++
		} else {
			failureCount++
		}
	}

	total := len(currentResults)
	if total > 0 {
		d.stats = DashboardStats{
			TotalChecks:   total,
			SuccessCount:  successCount,
			FailureCount:  failureCount,
			UptimePercent: float64(successCount) / float64(total) * 100,
			AvgLatency:    time.Duration(totalLatency/int64(total)) * time.Millisecond,
			MinLatency:    time.Duration(minLatency) * time.Millisecond,
			MaxLatency:    time.Duration(maxLatency) * time.Millisecond,
		}
	}

	d.lastUpdate = time.Now()
}

func (d *Dashboard) Stop() {
	close(d.stopChan)
}

func (d *Dashboard) Render() string {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()

	var output string

	output += headerStyle.Render(" ╔══════════════════════════════════════════════════════════════════╗ \n")
	output += headerStyle.Render(" ║                   Portman Health Monitor                    ║ \n")
	output += headerStyle.Render(" ╠══════════════════════════════════════════════════════════════════╣ \n")
	output += fmt.Sprintf(" ║  URLs: %d  │  Interval: %s  │  Updated: %s  ║\n",
		len(d.urls), d.interval, d.lastUpdate.Format("15:04:05"))
	output += headerStyle.Render(" ╚══════════════════════════════════════════════════════════════════╝ \n\n")

	output += d.renderStatsBar()

	output += tableHeaderStyle.Render("  #  │ URL                                    │ Status    │ Code │ Latency\n")
	output += tableHeaderStyle.Render(" ────┼────────────────────────────────────────┼───────────┼──────┼─────────\n")

	for i, url := range d.urls {
		result, ok := d.results[url]

		if !ok {
			output += fmt.Sprintf("  %d  │ %-36s │ %-9s │ %4s │ %s\n",
				i+1,
				truncateURL(url),
				disabledStyle.Render("Pending"),
				"-",
				disabledStyle.Render("-"))
			continue
		}

		statusStr := result.Status
		statusStyle := infoStyle
		statusIcon := "○"

		switch result.Status {
		case "UP":
			statusStyle = successStyle
			statusIcon = "●"
		case "DOWN":
			statusStyle = errorStyle
			statusIcon = "✗"
		case "WARNING":
			statusStyle = warningStyle
			statusIcon = "⚠"
		}

		output += fmt.Sprintf("  %d  │ %-36s │ %s %-6s │ %4d │ %s\n",
			i+1,
			truncateURL(url),
			statusStyle.Render(statusIcon),
			statusStyle.Render(statusStr),
			result.StatusCode,
			result.ResponseTime.String(),
		)
	}

	return output
}

func (d *Dashboard) renderStatsBar() string {
	uptimeColor := lipgloss.Color("82")
	if d.stats.UptimePercent < 80 {
		uptimeColor = lipgloss.Color("226")
	}
	if d.stats.UptimePercent < 50 {
		uptimeColor = lipgloss.Color("196")
	}

	uptimeStyle := lipgloss.NewStyle().Foreground(uptimeColor)

	stats := fmt.Sprintf("  ┌──────────────┬──────────────┬──────────────┬──────────────┐\n")

	stats += fmt.Sprintf("  │ Total: %-5d│ UP: %-7d│ DOWN: %-6d│ Uptime: %s│\n",
		d.stats.TotalChecks,
		d.stats.SuccessCount,
		d.stats.FailureCount,
		uptimeStyle.Render(fmt.Sprintf("%5.1f%%", d.stats.UptimePercent)))

	stats += fmt.Sprintf("  ├──────────────┼──────────────┼──────────────┼──────────────┤\n")

	if d.stats.TotalChecks > 0 {
		stats += fmt.Sprintf("  │ Avg: %-8s│ Min: %-8s│ Max: %-8s│            │\n",
			d.stats.AvgLatency.String(),
			d.stats.MinLatency.String(),
			d.stats.MaxLatency.String())
	} else {
		stats += fmt.Sprintf("  │ %-14s│ %-14s│ %-14s│            │\n",
			disabledStyle.Render("Avg: -"),
			disabledStyle.Render("Min: -"),
			disabledStyle.Render("Max: -"))
	}

	stats += fmt.Sprintf("  └──────────────┴──────────────┴──────────────┴──────────────┘\n\n")

	return stats
}

func truncateURL(url string) string {
	if len(url) > 36 {
		return url[:33] + "..."
	}
	return url
}

func (d *Dashboard) GetResults() map[string]*monitor.HealthResult {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()
	return d.results
}

func (d *Dashboard) GetStats() DashboardStats {
	return d.stats
}
