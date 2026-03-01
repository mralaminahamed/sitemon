package tui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mralaminahamed/sitemon/internal/http"
	"github.com/mralaminahamed/sitemon/internal/models"
	"github.com/mralaminahamed/sitemon/internal/monitor"
)

var (
	cyan    = lipgloss.Color("75")
	green   = lipgloss.Color("82")
	red     = lipgloss.Color("196")
	yellow  = lipgloss.Color("226")
	gray    = lipgloss.Color("240")
	magenta = lipgloss.Color("201")
	white   = lipgloss.Color("255")
	blue    = lipgloss.Color("33")
	orange  = lipgloss.Color("208")

	successStyle  = lipgloss.NewStyle().Foreground(green).Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(red).Bold(true)
	warningStyle  = lipgloss.NewStyle().Foreground(yellow).Bold(true)
	infoStyle     = lipgloss.NewStyle().Foreground(cyan)
	disabledStyle = lipgloss.NewStyle().Foreground(gray)
	urlStyle      = lipgloss.NewStyle().Foreground(white)
	boldStyle     = lipgloss.NewStyle().Bold(true)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(magenta).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	statBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(gray).
			Padding(0, 1).
			MarginRight(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan)
)

type Dashboard struct {
	urls             []string
	results          map[string]*models.HealthResult
	checking         map[string]bool
	resultsMu        sync.RWMutex
	interval         time.Duration
	client           *http.Client
	stopChan         chan bool
	stats            DashboardStats
	lastUpdate       time.Time
	startTime        time.Time
	width            int
	height           int
	history          []HistoryEntry
	historyMu        sync.RWMutex
	maxHistory       int
	bypassCloudflare bool
}

type DashboardStats struct {
	TotalChecks   int
	SuccessCount  int
	FailureCount  int
	UptimePercent float64
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	TotalRequests int
	RPS           float64
}

type HistoryEntry struct {
	URL        string
	Status     string
	StatusCode int
	Latency    time.Duration
	Timestamp  time.Time
}

func NewDashboard(urls []string, interval time.Duration, timeout time.Duration, bypassCloudflare bool) *Dashboard {
	width, height := getTerminalSize()
	return &Dashboard{
		urls:             urls,
		results:          make(map[string]*models.HealthResult),
		checking:         make(map[string]bool),
		interval:         interval,
		client:           http.NewClient(timeout),
		stopChan:         make(chan bool),
		stats:            DashboardStats{},
		lastUpdate:       time.Now(),
		startTime:        time.Now(),
		width:            width,
		height:           height,
		history:          make([]HistoryEntry, 0, 100),
		maxHistory:       100,
		bypassCloudflare: bypassCloudflare,
	}
}

func getTerminalSize() (width, height int) {
	width = 120
	height = 40

	if envWidth := os.Getenv("TERMINAL_WIDTH"); envWidth != "" {
		if w, err := fmt.Sscanf(envWidth, "%d", &width); err == nil && w > 0 {
			return width, height
		}
	}

	if w, h, err := getTermSize(); err == nil {
		width = w
		height = h
	}
	return width, height
}

func getTermSize() (width, height int, err error) {
	width = 120
	height = 40
	return width, height, nil
}

func (d *Dashboard) Start() error {
	clientTimeout := 10 * time.Second
	if d.bypassCloudflare {
		d.client = http.NewClient(clientTimeout, http.WithCloudflareBypass())
	}
	healthChecker := monitor.NewHealthChecker(d.client)

	d.checkURLs(healthChecker)

	go d.monitorLoop(healthChecker)

	for {
		select {
		case <-d.stopChan:
			d.print(d.Goodbye())
			return nil
		case <-time.After(d.interval):
			width, height := getTerminalSize()
			d.width = width
			d.height = height
			d.print(d.Render())
		}
	}
}

func (d *Dashboard) print(s string) {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
	fmt.Print(s)
}

func (d *Dashboard) Goodbye() string {
	var output string
	output += "\n\n"
	output += successStyle.Render("  ╔══════════════════════════════════════════════════════════╗\n")
	output += successStyle.Render("  ║           Thanks for using Sitemon!                     ║\n")
	output += successStyle.Render("  ╚══════════════════════════════════════════════════════════╝\n")
	output += "\n"
	return output
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

	currentResults := make(map[string]*models.HealthResult)
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
				d.addToHistory(u, result)
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

	d.stats.TotalRequests++
	d.stats.RPS = float64(d.stats.TotalRequests) / time.Since(d.startTime).Seconds()

	d.lastUpdate = time.Now()
}

func (d *Dashboard) addToHistory(url string, result *models.HealthResult) {
	d.historyMu.Lock()
	defer d.historyMu.Unlock()

	entry := HistoryEntry{
		URL:        url,
		Status:     result.Status,
		StatusCode: result.StatusCode,
		Latency:    result.ResponseTime,
		Timestamp:  result.Timestamp,
	}

	d.history = append(d.history, entry)
	if len(d.history) > d.maxHistory {
		d.history = d.history[1:]
	}
}

func (d *Dashboard) Stop() {
	close(d.stopChan)
}

func (d *Dashboard) Render() string {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()

	w := d.width
	if w < 80 {
		w = 80
	}

	var output string

	output += d.renderHeader(w)
	output += d.renderStatsBar(w)
	output += "\n"
	output += d.renderURLTable(w)
	output += "\n"
	output += d.renderRecentHistory(w)
	output += "\n"
	output += d.renderFooter(w)

	return output
}

func (d *Dashboard) renderHeader(w int) string {
	sep := lipgloss.NewStyle().Foreground(cyan).Render(string(make([]byte, w)))

	title := fmt.Sprintf(" ◈ Sitemon Dashboard ")
	output := titleStyle.Width(w).Render(title)
	output += "\n"
	output += sep + "\n"

	info := fmt.Sprintf(" URLs: %d | Interval: %s | Running: %s | Updated: %s",
		len(d.urls),
		d.interval.String(),
		d.uptime(),
		d.lastUpdate.Format("15:04:05"))
	output += lipgloss.NewStyle().Foreground(gray).Render(info) + "\n"
	output += sep + "\n"

	return output
}

func (d *Dashboard) renderStatsBar(w int) string {
	uptimeColor := green
	if d.stats.UptimePercent < 80 {
		uptimeColor = yellow
	}
	if d.stats.UptimePercent < 50 {
		uptimeColor = red
	}

	item := func(label string, value string, color lipgloss.Color) string {
		return lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Render(fmt.Sprintf("%s:%s", label, value))
	}

	stats := "  "
	stats += item(" TOTAL ", fmt.Sprintf("%d ", d.stats.TotalChecks), cyan)
	stats += item(" UP ", fmt.Sprintf("%d ", d.stats.SuccessCount), green)
	stats += item(" DOWN ", fmt.Sprintf("%d ", d.stats.FailureCount), red)
	stats += item(" UPTIME ", fmt.Sprintf("%.1f%% ", d.stats.UptimePercent), uptimeColor)

	stats += "\n  "
	stats += item(" AVG ", fmt.Sprintf("%s ", formatLatency(d.stats.AvgLatency)), cyan)
	stats += item(" MIN ", fmt.Sprintf("%s ", formatLatency(d.stats.MinLatency)), green)
	stats += item(" MAX ", fmt.Sprintf("%s ", formatLatency(d.stats.MaxLatency)), yellow)
	stats += item(" RPS ", fmt.Sprintf("%.1f ", d.stats.RPS), orange)

	return stats
}

func (d *Dashboard) renderURLTable(w int) string {
	numWidth := 4
	urlWidth := (w - numWidth - 14 - 12 - 12) / 2
	if urlWidth < 20 {
		urlWidth = 20
	}

	header := ""
	header += headerStyle.Width(numWidth).AlignHorizontal(lipgloss.Left).Render("#")
	header += "  "
	header += headerStyle.Width(urlWidth).Render("URL")
	header += "  "
	header += headerStyle.Width(14).Render("STATUS")
	header += "  "
	header += headerStyle.Width(12).Render("CODE")
	header += "  "
	header += headerStyle.Width(12).Render("LATENCY")

	output := header + "\n"

	sep := ""
	sep += lipgloss.NewStyle().Foreground(gray).Render(string(make([]byte, numWidth)))
	sep += "  "
	sep += lipgloss.NewStyle().Foreground(gray).Render(string(make([]byte, urlWidth)))
	sep += "  "
	sep += lipgloss.NewStyle().Foreground(gray).Render(string(make([]byte, 14)))
	sep += "  "
	sep += lipgloss.NewStyle().Foreground(gray).Render(string(make([]byte, 12)))
	sep += "  "
	sep += lipgloss.NewStyle().Foreground(gray).Render(string(make([]byte, 12)))
	output += sep + "\n"

	for i, url := range d.urls {
		result, ok := d.results[url]
		isChecking := d.checking[url]

		row := ""
		row += boldStyle.Width(numWidth).Render(fmt.Sprintf("%d", i+1))
		row += "  "

		truncatedURL := url
		if len(truncatedURL) > urlWidth {
			truncatedURL = truncatedURL[:urlWidth-3] + "..."
		}
		row += urlStyle.Width(urlWidth).Render(truncatedURL)
		row += "  "

		if isChecking {
			row += infoStyle.Width(14).Render("● Checking")
			row += "  "
			row += infoStyle.Width(12).Render("...")
			row += "  "
			row += infoStyle.Width(12).Render("...")
			output += row + "\n"
			continue
		}

		if !ok {
			row += disabledStyle.Width(14).Render("○ Pending")
			row += "  "
			row += disabledStyle.Width(12).Render("---")
			row += "  "
			row += disabledStyle.Width(12).Render("---")
			output += row + "\n"
			continue
		}

		var statusStyle, codeStyle, latencyStyle lipgloss.Style
		var statusText, codeText, latencyText string

		switch result.Status {
		case "UP":
			statusStyle = successStyle
			codeStyle = successStyle
			latencyStyle = d.getLatencyStyle(result.ResponseTime)
			statusText = "● UP"
			codeText = fmt.Sprintf("%d", result.StatusCode)
			latencyText = formatLatency(result.ResponseTime)
		case "DOWN":
			statusStyle = errorStyle
			codeStyle = errorStyle
			latencyStyle = errorStyle
			statusText = "✗ DOWN"
			codeText = fmt.Sprintf("%d", result.StatusCode)
			latencyText = formatLatency(result.ResponseTime)
		case "WARNING":
			statusStyle = warningStyle
			codeStyle = warningStyle
			latencyStyle = warningStyle
			statusText = "⚠ WARN"
			codeText = fmt.Sprintf("%d", result.StatusCode)
			latencyText = formatLatency(result.ResponseTime)
		case "REDIRECT":
			statusStyle = infoStyle
			codeStyle = infoStyle
			latencyStyle = infoStyle
			statusText = "↪ REDIRECT"
			codeText = fmt.Sprintf("%d", result.StatusCode)
			latencyText = formatLatency(result.ResponseTime)
		default:
			statusStyle = disabledStyle
			codeStyle = disabledStyle
			latencyStyle = disabledStyle
			statusText = "○ UNKNOWN"
			codeText = "---"
			latencyText = "---"
		}

		row += statusStyle.Width(14).Render(statusText)
		row += "  "
		row += codeStyle.Width(12).Render(codeText)
		row += "  "
		row += latencyStyle.Width(12).Render(latencyText)

		output += row + "\n"
	}

	return output
}

func (d *Dashboard) renderRecentHistory(w int) string {
	d.historyMu.RLock()
	defer d.historyMu.RUnlock()

	if len(d.history) == 0 {
		return ""
	}

	header := boldStyle.Render(" Recent Activity ")
	output := header + "\n"

	recent := d.history
	if len(recent) > 10 {
		recent = recent[len(recent)-10:]
	}

	for _, entry := range recent {
		timestamp := entry.Timestamp.Format("15:04:05")
		var statusIcon string
		var statusColor lipgloss.Color

		switch entry.Status {
		case "UP":
			statusIcon = "✓"
			statusColor = green
		case "DOWN":
			statusIcon = "✗"
			statusColor = red
		case "WARNING":
			statusIcon = "⚠"
			statusColor = yellow
		default:
			statusIcon = "○"
			statusColor = gray
		}

		truncatedURL := entry.URL
		if len(truncatedURL) > 40 {
			truncatedURL = truncatedURL[:37] + "..."
		}

		row := fmt.Sprintf(" %s [%s] %s - %d (%s) in %s",
			lipgloss.NewStyle().Foreground(gray).Render(timestamp),
			lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon),
			truncatedURL,
			entry.StatusCode,
			entry.Status,
			formatLatency(entry.Latency))
		output += row + "\n"
	}

	return output
}

func (d *Dashboard) renderFooter(w int) string {
	sep := lipgloss.NewStyle().Foreground(cyan).Render(string(make([]byte, w)))
	footer := sep + "\n"

	help := " Ctrl+C to exit "
	footer += disabledStyle.Render(help)

	footer += " | "
	footer += lipgloss.NewStyle().Foreground(gray).Render("Interval: " + d.interval.String())

	footer += " | "
	footer += lipgloss.NewStyle().Foreground(gray).Render("Uptime: " + d.uptime())

	return footer
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

func formatLatency(latency time.Duration) string {
	ms := latency.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

func (d *Dashboard) uptime() string {
	return time.Since(d.startTime).Round(time.Second).String()
}

func (d *Dashboard) GetResults() map[string]*models.HealthResult {
	d.resultsMu.RLock()
	defer d.resultsMu.RUnlock()
	return d.results
}

func (d *Dashboard) GetStats() DashboardStats {
	return d.stats
}
