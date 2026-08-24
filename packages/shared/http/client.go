package http

import (
	"crypto/tls"
	_ "embed"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

//go:embed user-agents.txt
var userAgentData string

var userAgents []string

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func init() {
	loadUserAgents()
}

// loadUserAgents parses the embedded user-agents list. Embedding makes the
// binary self-contained — no dependency on a data/ dir relative to cwd.
func loadUserAgents() {
	for _, line := range strings.Split(userAgentData, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			userAgents = append(userAgents, line)
		}
	}
	if len(userAgents) == 0 {
		userAgents = []string{defaultUserAgent}
	}
}

// getRandomUserAgent returns a random UA. The global rand source is
// auto-seeded since Go 1.20, so no manual Seed is needed (and Seed is
// deprecated).
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

type Client struct {
	client *resty.Client
}

type ClientOption func(*resty.Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *resty.Client) {
		c.SetTimeout(timeout)
	}
}

func WithRetry(count int, waitTime time.Duration) ClientOption {
	return func(c *resty.Client) {
		c.SetRetryCount(count)
		c.SetRetryWaitTime(waitTime)
	}
}

func WithTLSInsecure(skipVerify bool) ClientOption {
	return func(c *resty.Client) {
		c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: skipVerify})
	}
}

func WithHeader(key, value string) ClientOption {
	return func(c *resty.Client) {
		c.SetHeader(key, value)
	}
}

func WithHeaders(headers map[string]string) ClientOption {
	return func(c *resty.Client) {
		c.SetHeaders(headers)
	}
}

func WithProxy(proxyURL string) ClientOption {
	return func(c *resty.Client) {
		c.SetProxy(proxyURL)
	}
}

func WithFollowRedirects(follow bool) ClientOption {
	return func(c *resty.Client) {
		c.SetRedirectPolicy(resty.NoRedirectPolicy())
		if follow {
			c.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))
		}
	}
}

func WithCookie(cookie string) ClientOption {
	return func(c *resty.Client) {
		c.SetHeader("Cookie", cookie)
	}
}

func WithCloudflareBypass() ClientOption {
	ua := getRandomUserAgent()
	return func(c *resty.Client) {
		c.SetHeader("User-Agent", ua)
		c.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		c.SetHeader("Accept-Language", "en-US,en;q=0.9")
		c.SetHeader("Accept-Encoding", "gzip, deflate")
		c.SetHeader("DNT", "1")
		c.SetHeader("Connection", "keep-alive")
		c.SetHeader("Upgrade-Insecure-Requests", "1")
		c.SetHeader("Sec-Fetch-Dest", "document")
		c.SetHeader("Sec-Fetch-Mode", "navigate")
		c.SetHeader("Sec-Fetch-Site", "none")
		c.SetHeader("Sec-Fetch-User", "?1")
		c.SetHeader("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
		c.SetHeader("Sec-Ch-Ua-Mobile", "?0")
		c.SetHeader("Sec-Ch-Ua-Platform", `"Windows"`)
		c.SetHeader("Cache-Control", "max-age=0")
	}
}

func generateFakeCFCookies() map[string]string {
	cookies := make(map[string]string)
	cookies["cf_clearance"] = generateRandomString(43) + "." + generateRandomString(8)
	cookies["_cfuvid"] = generateRandomString(22)
	cookies["__cf_bm"] = generateRandomString(32)
	return cookies
}

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func NewClient(timeout time.Duration, opts ...ClientOption) *Client {
	c := resty.New().
		SetTimeout(timeout).
		SetRetryCount(3).
		SetRetryWaitTime(500 * time.Millisecond).
		SetTransport(&http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		})

	for _, opt := range opts {
		opt(c)
	}

	return &Client{client: c}
}

func NewFastClient(timeout time.Duration, opts ...ClientOption) *Client {
	c := resty.New().
		SetTimeout(timeout).
		SetRetryCount(0).
		SetHeader("Connection", "keep-alive").
		SetTransport(&http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     120 * time.Second,
		}).
		SetRetryWaitTime(100 * time.Millisecond)

	for _, opt := range opts {
		opt(c)
	}

	return &Client{client: c}
}

func (c *Client) Get(url string) (*resty.Response, error) {
	return c.client.R().Get(url)
}

func (c *Client) Post(url string, body interface{}) (*resty.Response, error) {
	return c.client.R().SetBody(body).Post(url)
}

func (c *Client) Put(url string, body interface{}) (*resty.Response, error) {
	return c.client.R().SetBody(body).Put(url)
}

func (c *Client) Patch(url string, body interface{}) (*resty.Response, error) {
	return c.client.R().SetBody(body).Patch(url)
}

func (c *Client) Delete(url string) (*resty.Response, error) {
	return c.client.R().Delete(url)
}

func (c *Client) Head(url string) (*resty.Response, error) {
	return c.client.R().Head(url)
}

func (c *Client) Options(url string) (*resty.Response, error) {
	return c.client.R().Options(url)
}

type RequestResult struct {
	StatusCode   int
	ResponseTime time.Duration
	ResponseSize int64
	Error        error
}

func (c *Client) SendHead(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Head(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendGet(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Get(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendPost(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Post(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendPut(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Put(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendPatch(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Patch(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendDelete(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Delete(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SendOptions(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Options(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		ResponseSize: responseSize(resp),
		Error:        err,
	}
}

func (c *Client) SetHeader(key, value string) *Client {
	c.client.SetHeader(key, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	c.client.SetHeaders(headers)
	return c
}

func statusCode(resp *resty.Response, err error) int {
	if err != nil || resp == nil {
		return 0
	}
	return resp.StatusCode()
}

func responseSize(resp *resty.Response) int64 {
	if resp == nil {
		return 0
	}
	return int64(len(resp.Body()))
}
