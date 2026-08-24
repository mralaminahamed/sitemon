package http

import (
	"testing"
	"time"
)

func TestUserAgentsEmbedded(t *testing.T) {
	if len(userAgents) == 0 {
		t.Fatal("no user agents loaded from embedded list")
	}
	if getRandomUserAgent() == "" {
		t.Error("getRandomUserAgent returned empty string")
	}
}

func TestNewFastClient(t *testing.T) {
	c := NewFastClient(2*time.Second, WithHeader("X-Test", "1"))
	if c == nil || c.client == nil {
		t.Fatal("NewFastClient returned nil client")
	}
}
