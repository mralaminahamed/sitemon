package notify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewNotifierProviderDetection(t *testing.T) {
	cases := map[string]string{
		"https://hooks.slack.com/services/x":        "slack",
		"https://discord.com/api/webhooks/x":        "discord",
		"https://api.telegram.org/botX/sendMessage": "telegram",
		"https://example.com/hook":                  "generic",
	}
	for url, want := range cases {
		if got := NewNotifier(url).provider; got != want {
			t.Errorf("provider(%q)=%q want %q", url, got, want)
		}
	}
}

func TestSendTelegramRequiresChatID(t *testing.T) {
	n := NewNotifier("https://api.telegram.org/botX/sendMessage") // no TELEGRAM_CHAT_ID
	if err := n.SendAlert("https://x.com", "DOWN", 500, "down"); err == nil {
		t.Fatal("expected error when TELEGRAM_CHAT_ID unset")
	}
}

func TestSendTelegramIncludesChatID(t *testing.T) {
	var body TelegramPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := &Notifier{webhookURL: srv.URL, provider: "telegram", telegramChatID: "12345", client: srv.Client()}
	if err := n.SendAlert("https://x.com", "DOWN", 500, "down"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if body.ChatID != "12345" {
		t.Errorf("chat_id=%q want 12345", body.ChatID)
	}
}
