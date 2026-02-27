package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"
)

type CertificateInfo struct {
	URL           string    `json:"url"`
	Host          string    `json:"host"`
	Issuer        string    `json:"issuer"`
	Subject       string    `json:"subject"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidUntil    time.Time `json:"valid_until"`
	DaysRemaining int       `json:"days_remaining"`
	IsValid       bool      `json:"is_valid"`
	Protocol      string    `json:"protocol"`
	CipherSuite   string    `json:"cipher_suite"`
}

func CheckCertificate(url string, timeout time.Duration) (*CertificateInfo, error) {
	config := &tls.Config{InsecureSkipVerify: true}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", extractHost(url)+":443", config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	err = conn.Handshake()
	if err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := certs[0]
	now := time.Now()

	info := &CertificateInfo{
		URL:        url,
		Host:       extractHost(url),
		Issuer:     getCertIssuer(cert),
		Subject:    getCertSubject(cert),
		ValidFrom:  cert.NotBefore,
		ValidUntil: cert.NotAfter,
		IsValid:    now.After(cert.NotBefore) && now.Before(cert.NotAfter),
		Protocol:   fmt.Sprintf("TLS %d", state.Version),
	}

	info.CipherSuite = fmt.Sprintf("%x", state.CipherSuite)

	info.DaysRemaining = int(cert.NotAfter.Sub(now).Hours() / 24)

	return info, nil
}

func CheckCertificateHTTP(url string, timeout time.Duration) (*CertificateInfo, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: timeout,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.TLS == nil {
		return nil, fmt.Errorf("no TLS connection")
	}

	certs := resp.TLS.PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := certs[0]
	now := time.Now()

	info := &CertificateInfo{
		URL:        url,
		Host:       resp.TLS.ServerName,
		Issuer:     getCertIssuer(cert),
		Subject:    getCertSubject(cert),
		ValidFrom:  cert.NotBefore,
		ValidUntil: cert.NotAfter,
		IsValid:    now.After(cert.NotBefore) && now.Before(cert.NotAfter),
		Protocol:   fmt.Sprintf("TLS %d", resp.TLS.Version),
	}

	info.CipherSuite = fmt.Sprintf("%x", resp.TLS.CipherSuite)

	info.DaysRemaining = int(cert.NotAfter.Sub(now).Hours() / 24)

	return info, nil
}

func extractHost(url string) string {
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}

	for i, c := range url {
		if c == '/' || c == '?' {
			url = url[:i]
			break
		}
	}

	return url
}

func getCertIssuer(cert *x509.Certificate) string {
	if len(cert.Issuer.Organization) > 0 {
		return cert.Issuer.Organization[0]
	}
	if len(cert.Issuer.OrganizationalUnit) > 0 {
		return cert.Issuer.OrganizationalUnit[0]
	}
	return cert.Issuer.CommonName
}

func getCertSubject(cert *x509.Certificate) string {
	if len(cert.Subject.Organization) > 0 {
		return cert.Subject.Organization[0]
	}
	return cert.Subject.CommonName
}
