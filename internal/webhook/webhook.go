package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Payload struct {
	Text      string `json:"text"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	StatusCode int    `json:"status_code"`
	Timestamp string `json:"timestamp"`
	AlertType string `json:"alert_type"`
}

type SlackPayload struct {
	Text      string       `json:"text"`
	Blocks    []SlackBlock `json:"blocks,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackBlock struct {
	Type string `json:"type"`
	Text *SlackText `json:"text,omitempty"`
	Fields []SlackField `json:"fields,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackField struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Color string

type SlackAttachment struct {
	Color  string      `json:"color"`
	Blocks []SlackBlock `json:"blocks,omitempty"`
}

type DiscordPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Color       int            `json:"color"`
	Timestamp   string          `json:"timestamp"`
	Fields      []DiscordField  `json:"fields,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}

type TelegramPayload struct {
	ChatID    string `json:"chat_id"`
	ParseMode string `json:"parse_mode"`
	Text      string `json:"text"`
}

type Notifier struct {
	webhookURL string
	provider   string
	client     *http.Client
}

func NewNotifier(webhookURL string) *Notifier {
	provider := "generic"
	if strings.Contains(webhookURL, "slack.com") {
		provider = "slack"
	} else if strings.Contains(webhookURL, "discord.com") || strings.Contains(webhookURL, "discordapp.com") {
		provider = "discord"
	} else if strings.Contains(webhookURL, "telegram") {
		provider = "telegram"
	}

	return &Notifier{
		webhookURL: webhookURL,
		provider:   provider,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) SendAlert(url, status string, statusCode int, alertType string) error {
	timestamp := time.Now().Format(time.RFC3339)

	switch n.provider {
	case "slack":
		return n.sendSlack(url, status, statusCode, timestamp, alertType)
	case "discord":
		return n.sendDiscord(url, status, statusCode, timestamp, alertType)
	case "telegram":
		return n.sendTelegram(url, status, statusCode, timestamp, alertType)
	default:
		return n.sendGeneric(url, status, statusCode, timestamp, alertType)
	}
}

func (n *Notifier) sendSlack(url, status string, statusCode int, timestamp, alertType string) error {
	color := "#36a64f" // green
	if status == "DOWN" {
		color = "#ff0000" // red
	} else if status == "WARNING" {
		color = "#ff9900" // yellow
	}

	text := fmt.Sprintf("🚨 Portman Alert: %s", status)
	if alertType == "recovery" {
		text = fmt.Sprintf("✅ Portman Recovery: %s", status)
	}

	payload := SlackPayload{
		Text: text,
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: text,
				},
			},
			{
				Type: "section",
				Fields: []SlackField{
					{Type: "mrkdwn", Text: fmt.Sprintf("*URL:*\n%s", url)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Status:*\n%s (%d)", status, statusCode)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Time:*\n%s", timestamp)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Type:*\n%s", alertType)},
				},
			},
		},
		Attachments: []SlackAttachment{
			{
				Color: color,
			},
		},
	}

	return n.sendPayload(payload)
}

func (n *Notifier) sendDiscord(url, status string, statusCode int, timestamp, alertType string) error {
	color := 65280 // green
	if status == "DOWN" {
		color = 16711680 // red
	} else if status == "WARNING" {
		color = 16776960 // yellow
	}

	title := fmt.Sprintf("🚨 Portman Alert: %s", status)
	if alertType == "recovery" {
		title = fmt.Sprintf("✅ Portman Recovery: %s", status)
	}

	payload := DiscordPayload{
		Embeds: []DiscordEmbed{
			{
				Title:       title,
				Description: fmt.Sprintf("URL: %s", url),
				Color:       color,
				Timestamp:   timestamp,
				Fields: []DiscordField{
					{Name: "Status", Value: fmt.Sprintf("%s (%d)", status, statusCode), Inline: true},
					{Name: "Type", Value: alertType, Inline: true},
				},
				Footer: &DiscordFooter{
					Text: "Portman",
				},
			},
		},
	}

	return n.sendPayload(payload)
}

func (n *Notifier) sendTelegram(url, status string, statusCode int, timestamp, alertType string) error {
	emoji := "🔴"
	if status == "UP" {
		emoji = "🟢"
	} else if status == "WARNING" {
		emoji = "🟡"
	}

	text := fmt.Sprintf("%s *Portman Alert*\n\n*URL:* %s\n*Status:* %s (%d)\n*Time:* %s\n*Type:* %s",
		emoji, url, status, statusCode, timestamp, alertType)

	if alertType == "recovery" {
		text = fmt.Sprintf("✅ *Portman Recovery*\n\n%s", text)
	}

	payload := TelegramPayload{
		ParseMode: "Markdown",
		Text:      text,
	}

	return n.sendPayload(payload)
}

func (n *Notifier) sendGeneric(url, status string, statusCode int, timestamp, alertType string) error {
	payload := Payload{
		Text:      fmt.Sprintf("Portman Alert: %s - %s", status, url),
		URL:       url,
		Status:    status,
		StatusCode: statusCode,
		Timestamp: timestamp,
		AlertType: alertType,
	}

	return n.sendPayload(payload)
}

func (n *Notifier) sendPayload(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", n.webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}
