// Package urlguard blocks server-side fetches of unsafe targets (SSRF).
package urlguard

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Check rejects a target URL that is not http/https or that resolves to a
// loopback, private, link-local, or unspecified address. Set
// ALLOW_PRIVATE_TARGETS=true to bypass (local dev / tests).
func Check(raw string) error {
	u, err := parse(raw)
	if err != nil {
		return err
	}
	if allowPrivate() {
		return nil
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", host, err)
	}
	for _, ip := range ips {
		if blocked(ip) {
			return fmt.Errorf("target %s resolves to blocked address %s", host, ip)
		}
	}
	return nil
}

func parse(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	return u, nil
}

func blocked(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func allowPrivate() bool {
	return strings.EqualFold(os.Getenv("ALLOW_PRIVATE_TARGETS"), "true")
}
