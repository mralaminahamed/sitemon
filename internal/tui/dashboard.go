package tui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mralaminahamed/portman/internal/http"
	"github.com/mralaminahamed/portman/internal/monitor"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))
)

type DashboardModel struct {
	urls      []string
	results   map[string]*monitor.HealthResult
	resultsMu sync.RWMutex
	interval  time.Duration
	client    *http.Client
	stopChan  chan bool
	width     int
	height    int
}

func NewDashboard(urls []string, interval time.Duration, timeout time.Duration) *DashboardModel {
	return &DashboardModel{
		urls:     urls,
		results:  make(map[string]*monitor.HealthResult),
		interval: interval,
		client:   http.NewClient(timeout),
		stopChan: make(chan bool),
	}
}

func (m *DashboardModel) Start() error {
	healthChecker := monitor.NewHealthChecker(m.client)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-m.stopChan:
				return
			default:
				for _, url := range m.urls {
					result, err := healthChecker.Check(url)
					if err == nil {
						m.resultsMu.Lock()
						m.results[url] = result
						m.resultsMu.Unlock()
					}
					time.Sleep(100 * time.Millisecond)
				}
				time.Sleep(m.interval)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		m.Stop()
	}()

	fmt.Println(m.renderDashboard())
	fmt.Println("\nPress Ctrl+C to stop...")

	wg.Wait()
	return nil
}

func (m *DashboardModel) Stop() {
	close(m.stopChan)
}

func (m *DashboardModel) renderDashboard() string {
	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()

	header := headerStyle.Render("╔══════════════════════════════════════════════════════════════════╗\n")
	header += headerStyle.Render("║                     Portman Health Monitor                      ║\n")
	header += headerStyle.Render("╠══════════════════════════════════════════════════════════════════╣\n")
	header += fmt.Sprintf("║ URLs: %d | Interval: %s | Time: %s ║\n",
		len(m.urls), m.interval, time.Now().Format("15:04:05"))
	header += headerStyle.Render("╚══════════════════════════════════════════════════════════════════╝\n\n")

	header += fmt.Sprintf("%-40s %-15s %-8s %s\n", "URL", "Status", "Code", "Latency")
	header += fmt.Sprintf("%s\n", "─────────────────────────────────────────────────────────────────────────────")

	for _, url := range m.urls {
		result, ok := m.results[url]
		if !ok {
			header += fmt.Sprintf("%-40s %-15s %-8s %s\n", url, "Checking...", "-", "-")
			continue
		}

		status := result.Status
		statusColor := infoStyle
		if status == "UP" {
			status = "✓ UP"
			statusColor = successStyle
		} else if status == "DOWN" {
			status = "✗ DOWN"
			statusColor = errorStyle
		} else if status == "WARNING" {
			status = "⚠ WARNING"
			statusColor = warningStyle
		}

		header += fmt.Sprintf("%-40s %-15s %-8d %s\n",
			url,
			statusColor.Render(status),
			result.StatusCode,
			result.ResponseTime.String(),
		)
	}

	return header
}
