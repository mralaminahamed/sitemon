package alert

import (
	"time"

	"github.com/mralaminahamed/portman/internal/monitor"
	"github.com/mralaminahamed/portman/internal/webhook"
)

type ThresholdConfig struct {
	MaxLatency    time.Duration
	MinStatusCode int
	MaxStatusCode int
	WebhookURL    string
}

type AlertManager struct {
	config       *ThresholdConfig
	lastStatus   map[string]string
	notifier     *webhook.Notifier
}

func NewAlertManager(config *ThresholdConfig) *AlertManager {
	am := &AlertManager{
		config:     config,
		lastStatus: make(map[string]string),
	}

	if config.WebhookURL != "" {
		am.notifier = webhook.NewNotifier(config.WebhookURL)
	}

	return am
}

func (am *AlertManager) CheckAndAlert(url string, result *monitor.HealthResult) error {
	if am.config == nil {
		return nil
	}

	if am.config.MaxLatency > 0 && result.ResponseTime > am.config.MaxLatency {
		if err := am.sendAlert(url, "WARNING", result.StatusCode, "latency_threshold"); err != nil {
			return err
		}
	}

	if am.config.MinStatusCode > 0 && result.StatusCode < am.config.MinStatusCode {
		if err := am.sendAlert(url, "WARNING", result.StatusCode, "status_threshold"); err != nil {
			return err
		}
	}

	if am.config.MaxStatusCode > 0 && result.StatusCode > am.config.MaxStatusCode {
		if err := am.sendAlert(url, "WARNING", result.StatusCode, "status_threshold"); err != nil {
			return err
		}
	}

	return nil
}

func (am *AlertManager) sendAlert(url string, status string, statusCode int, alertType string) error {
	if am.notifier == nil {
		return nil
	}

	previousStatus, exists := am.lastStatus[url]
	am.lastStatus[url] = status

	if !exists {
		return nil
	}

	if previousStatus == "UP" && status != "UP" {
		return am.notifier.SendAlert(url, status, statusCode, "down")
	}

	if previousStatus != "UP" && status == "UP" {
		return am.notifier.SendAlert(url, status, statusCode, "recovery")
	}

	return nil
}

func (am *AlertManager) GetLastStatus(url string) string {
	return am.lastStatus[url]
}
