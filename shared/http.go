package shared

import (
	"io"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	config *Config
}

func NewHTTPClient(config *Config) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
		config: config,
	}
}

func (c *Client) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer " + c.config.AuthToken)
	return c.Do(req)
}

func (c *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer " + c.config.AuthToken)
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
