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
	cyan    = lipgloss.Color("75")
	green   = lipgloss.Color("82")
	red     = lipgloss.Color("196")
	yellow  = lipgloss.Color("226")
	gray    = lipgloss.Color("240")
	magenta = lipgloss.Color("201")
	white   = lipgloss.Color("255")

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(cyan)

	disabledStyle = lipgloss.NewStyle().
			Foreground(gray)

	urlStyle = lipgloss.NewStyle().
			Foreground(white)
)

type Dashboard struct {
	urls       []string
	results    map[string]*monitor.HealthResult
	checking   map[string]bool
	resultsMu  sync.RWMutex
	interval   time.Duration
	client     *http.Client
	stopChan   chan bool
	width      int
	height     int
	stats      DashboardStats
	lastUpdate time.Time
	startTime  time.Time
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
		urls:       urls,
		results:    make(map[string]*monitor.HealthResult),
		checking:   make(map[string]bool),
		interval:   interval,
		client:     http.NewClient(timeout),
		stopChan:   make(chan bool),
		stats:      DashboardStats{},
		lastUpdate: time.Now(),
		startTime:  time.Now(),
	}
}

func (d *Dashboard) Start() error {
	healthChecker := monitor.NewHealthChecker(d.client)

	d.checkURLs(healthChecker)

	go d.monitorLoop(healthChecker)

	fmt.Print(d.Render())
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
			fmt.Print(d.Render())
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

	d.resultsMu.Lock()
	for _, url := range d.urls {
		d.checking[url] = true
	}
	d.resultsMu.Unlock()

	for _, url := range d.urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result, err := healthChecker.Check(u)
			d.resultsMu.Lock()
			d.checking[u] = false
			d.resultsMu.Unlock()
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

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(magenta).
		Render("◈ Portman Dashboard")

	info := lipgloss.NewStyle().
		Foreground(gray).
		Render(fmt.Sprintf("%d URLs • %s • %s",
			len(d.urls),
			d.interval.String(),
			d.lastUpdate.Format("15:04:05")))

	output += title + "\n" + info + "\n"

	separator := lipgloss.NewStyle().Foreground(cyan).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	output += separator + "\n"

	output += d.renderStatsBar()

	numCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(3).AlignHorizontal(lipgloss.Left)
	urlCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(32)
	statusCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(12)
	codeCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(6)
	latencyCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(10)
	infoCol := lipgloss.NewStyle().Bold(true).Foreground(cyan).Width(14)

	output += numCol.Render("#") + "  " +
		urlCol.Render("URL") + "  " +
		statusCol.Render("STATUS") + "  " +
		codeCol.Render("CODE") + "  " +
		latencyCol.Render("LATENCY") + "  " +
		infoCol.Render("INFO") + "\n"

	sepStyle := lipgloss.NewStyle().Foreground(gray)
	output += sepStyle.Render("───") + "  " +
		sepStyle.Render("────────────────────────────────") + "  " +
		sepStyle.Render("────────────") + "  " +
		sepStyle.Render("──────") + "  " +
		sepStyle.Render("──────────") + "  " +
		sepStyle.Render("──────────────") + "\n"

	for i, url := range d.urls {
		result, ok := d.results[url]
		isChecking := d.checking[url]

		rowNum := numCol.Render(fmt.Sprintf("%d", i+1))
		urlCell := urlStyle.Render(truncateURL(url, 30))

		if isChecking {
			statusCell := infoStyle.Render("● Checking")
			codeCell := infoStyle.Render("...")
			latencyCell := infoStyle.Render("...")
			infoCell := infoStyle.Render("in progress")
			output += rowNum + "  " + urlCell + "  " + statusCell + "  " + codeCell + "  " + latencyCell + "  " + infoCell + "\n"
			continue
		}

		if !ok {
			statusCell := disabledStyle.Render("○ Pending")
			codeCell := disabledStyle.Render("---")
			latencyCell := disabledStyle.Render("---")
			infoCell := disabledStyle.Render("awaiting")
			output += rowNum + "  " + urlCell + "  " + statusCell + "  " + codeCell + "  " + latencyCell + "  " + infoCell + "\n"
			continue
		}

		var statusCell, codeCell, latencyCell, infoCell string

		switch result.Status {
		case "UP":
			statusCell = successStyle.Render("● UP")
			codeCell = successStyle.Render(fmt.Sprintf("%d", result.StatusCode))
			latencyCell = d.getLatencyStyle(result.ResponseTime).Render(d.formatLatency(result.ResponseTime))
			infoCell = successStyle.Render("healthy")
		case "DOWN":
			statusCell = errorStyle.Render("✗ DOWN")
			codeCell = errorStyle.Render(fmt.Sprintf("%d", result.StatusCode))
			latencyCell = errorStyle.Render(d.formatLatency(result.ResponseTime))
			if result.Error != "" {
				infoCell = errorStyle.Render(truncateError(result.Error, 12))
			} else {
				infoCell = errorStyle.Render("server error")
			}
		case "WARNING":
			statusCell = warningStyle.Render("⚠ WARN")
			codeCell = warningStyle.Render(fmt.Sprintf("%d", result.StatusCode))
			latencyCell = warningStyle.Render(d.formatLatency(result.ResponseTime))
			infoCell = warningStyle.Render("client error")
		case "REDIRECT":
			statusCell = infoStyle.Render("↪ REDIRECT")
			codeCell = infoStyle.Render(fmt.Sprintf("%d", result.StatusCode))
			latencyCell = infoStyle.Render(d.formatLatency(result.ResponseTime))
			infoCell = infoStyle.Render("redirected")
		}

		output += rowNum + "  " + urlCell + "  " + statusCell + "  " + codeCell + "  " + latencyCell + "  " + infoCell + "\n"
	}

	output += "\n" + disabledStyle.Render("Ctrl+C to exit • "+d.interval.String()+" refresh • "+d.uptime())

	return output
}

func (d *Dashboard) getLatencyStyle(latency time.Duration) lipgloss.Style {
	ms := latency.Milliseconds()
	switch {
	case ms < 200:
		return successStyle
	case ms < 500:
		return infoStyle
	case ms < 1000:
		return warningStyle
	default:
		return errorStyle
	}
}

func (d *Dashboard) formatLatency(latency time.Duration) string {
	ms := latency.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

func truncateURL(url string, maxLen int) string {
	if len(url) > maxLen {
		return url[:maxLen-3] + "..."
	}
	return url
}

func truncateError(err string, maxLen int) string {
	if len(err) > maxLen {
		return err[:maxLen-3] + "..."
	}
	return err
}

func (d *Dashboard) uptime() string {
	return time.Since(d.startTime).Round(time.Second).String()
}

func (d *Dashboard) renderStatsBar() string {
	uptimeColor := green
	if d.stats.UptimePercent < 80 {
		uptimeColor = yellow
	}
	if d.stats.UptimePercent < 50 {
		uptimeColor = red
	}

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(gray).
		Padding(0, 1).
		MarginRight(1)

	totalBox := boxStyle.Copy().BorderForeground(cyan).Render(fmt.Sprintf(" TOTAL %d ", d.stats.TotalChecks))
	upBox := boxStyle.Copy().BorderForeground(green).Render(fmt.Sprintf(" UP %d ", d.stats.SuccessCount))
	downBox := boxStyle.Copy().BorderForeground(red).Render(fmt.Sprintf(" DOWN %d ", d.stats.FailureCount))
	uptimeBox := boxStyle.Copy().BorderForeground(uptimeColor).Render(fmt.Sprintf(" %.1f%% ", d.stats.UptimePercent))

	row1 := "  " + totalBox + upBox + downBox + uptimeBox + "\n"

	avgBox := boxStyle.Copy().BorderForeground(cyan).Render(fmt.Sprintf(" AVG %s ", d.stats.AvgLatency.String()))
	minBox := boxStyle.Copy().BorderForeground(green).Render(fmt.Sprintf(" MIN %s ", d.stats.MinLatency.String()))
	maxBox := boxStyle.Copy().BorderForeground(yellow).Render(fmt.Sprintf(" MAX %s ", d.stats.MaxLatency.String()))

	row2 := "  " + avgBox + minBox + maxBox + "\n"

	return row1 + row2
}

func (d *Dashboard) GetResults() map[string]*monitor.HealthResult {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()
	return d.results
}

func (d *Dashboard) GetStats() DashboardStats {
	return d.stats
}
