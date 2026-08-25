package notify

import (
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

type ThresholdConfig struct {
	MaxLatency    time.Duration
	MinStatusCode int
	MaxStatusCode int
	WebhookURL    string
}

type AlertManager struct {
	config     *ThresholdConfig
	lastStatus map[string]string
	notifier   *Notifier
}

func NewAlertManager(config *ThresholdConfig) *AlertManager {
	am := &AlertManager{
		config:     config,
		lastStatus: make(map[string]string),
	}

	if config.WebhookURL != "" {
		am.notifier = NewNotifier(config.WebhookURL)
	}

	return am
}

func (am *AlertManager) CheckAndAlert(url string, result *models.HealthResult) error {
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

	previousStatus := am.lastStatus[url]
	am.lastStatus[url] = status

	if alert := Classify(previousStatus, status); alert != "" {
		return am.notifier.SendAlert(url, status, statusCode, alert)
	}
	return nil
}

// Evaluate records the latest status for url and sends a webhook only when the
// status transitions (UP -> not-UP fires a "down" alert, not-UP -> UP fires a
// "recovery" alert). Calling it every tick is intentional and cheap: without
// this transition gate the watch/schedule loops fired an alert on every single
// poll while a site stayed down, and never sent recovery. Safe to call on a
// nil receiver or with no notifier configured.
func (am *AlertManager) Evaluate(url, status string, statusCode int) error {
	if am == nil || am.notifier == nil {
		return nil
	}

	prev := am.lastStatus[url]
	am.lastStatus[url] = status

	if alert := Classify(prev, status); alert != "" {
		return am.notifier.SendAlert(url, status, statusCode, alert)
	}
	return nil
}

func (am *AlertManager) GetLastStatus(url string) string {
	return am.lastStatus[url]
}
