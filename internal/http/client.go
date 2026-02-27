package http

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: resty.New().
			SetTimeout(timeout).
			SetRetryCount(3).
			SetRetryWaitTime(500 * time.Millisecond).
			SetCloseConnection(true),
	}
}

func NewFastClient(timeout time.Duration) *Client {
	return &Client{
		client: resty.New().
			SetTimeout(timeout).
			SetRetryCount(0).
			SetHeader("Connection", "keep-alive"),
	}
}

func (c *Client) Get(url string) (*resty.Response, error) {
	return c.client.R().Get(url)
}

func (c *Client) Post(url string, body interface{}) (*resty.Response, error) {
	return c.client.R().SetBody(body).Post(url)
}

func (c *Client) Delete(url string) (*resty.Response, error) {
	return c.client.R().Delete(url)
}

func (c *Client) Head(url string) (*resty.Response, error) {
	return c.client.R().Head(url)
}

func (c *Client) SendHead(url string) (int, error) {
	resp, err := c.client.R().Head(url)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

func (c *Client) SendGet(url string) (int, error) {
	resp, err := c.client.R().Get(url)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

func (c *Client) SendPost(url string, body interface{}) (int, error) {
	resp, err := c.client.R().SetBody(body).Post(url)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

func (c *Client) SendDelete(url string) (int, error) {
	resp, err := c.client.R().Delete(url)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode(), nil
}

func (c *Client) SetHeader(key, value string) *Client {
	c.client.SetHeader(key, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	c.client.SetHeaders(headers)
	return c
}
