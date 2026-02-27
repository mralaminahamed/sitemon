package http

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

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

func NewFastClient(timeout time.Duration) *Client {
	return &Client{
		client: resty.New().
			SetTimeout(timeout).
			SetRetryCount(0).
			SetHeader("Connection", "keep-alive").
			SetTransport(&http.Transport{
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 200,
				IdleConnTimeout:     120 * time.Second,
			}).
			SetRetryWaitTime(100 * time.Millisecond),
	}
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
	Error        error
}

func (c *Client) SendHead(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Head(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendGet(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Get(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendPost(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Post(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendPut(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Put(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendPatch(url string, body interface{}) RequestResult {
	start := time.Now()
	resp, err := c.client.R().SetBody(body).Patch(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendDelete(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Delete(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
	}
}

func (c *Client) SendOptions(url string) RequestResult {
	start := time.Now()
	resp, err := c.client.R().Options(url)
	return RequestResult{
		StatusCode:   statusCode(resp, err),
		ResponseTime: time.Since(start),
		Error:       err,
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
