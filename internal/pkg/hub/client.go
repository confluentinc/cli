package hub

import (
	"net/http"
	"time"

	testserver "github.com/confluentinc/cli/test/test-server"
)

type Client struct {
	URL    string
	Debug  bool
	Client *http.Client
}

func NewClient(isTest, unsafeTrace bool) *Client {
	url := "https://api.hub.confluent.io"
	if isTest {
		url = testserver.TestHubUrl.String()
	}

	return &Client{
		URL:    url,
		Debug:  unsafeTrace,
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}
