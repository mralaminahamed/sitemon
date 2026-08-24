package ssl

import "testing"

func TestExtractHost(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"https://example.com", "example.com"},
		{"https://example.com/path", "example.com"},
		{"http://sub.example.com/a/b?q=1", "sub.example.com"},
		{"example.com/x", "example.com"},
		{"https://example.com?q=1", "example.com"},
	}
	for _, tt := range tests {
		if got := extractHost(tt.in); got != tt.want {
			t.Errorf("extractHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
