package hub

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/confluentinc/cli/v4/pkg/httpretry"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

type Client struct {
	URL       string
	Debug     bool
	UserAgent string
	Client    *http.Client
}

func NewClient(userAgent string, isTest, unsafeTrace bool) *Client {
	url := "https://api.hub.confluent.io"
	if isTest {
		url = testserver.TestHubUrl.String()
	}

	client := retryablehttp.NewClient()
	httpretry.Configure(client, unsafeTrace, httpretry.Options{})

	return &Client{
		URL:       url,
		Debug:     unsafeTrace,
		UserAgent: userAgent,
		Client:    client.StandardClient(),
	}
}
