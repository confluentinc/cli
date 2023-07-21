package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

type Client struct {
	*srsdk.APIClient
	authToken string
}

func NewClient(url, userAgent string, unsafeTrace bool, authToken, clusterId string) *Client {
	cfg := srsdk.NewConfiguration()
	cfg.BasePath = url
	cfg.Debug = unsafeTrace
	cfg.UserAgent = userAgent
	cfg.HTTPClient = ccloudv2.NewRetryableHttpClient(unsafeTrace)
	cfg.DefaultHeader = map[string]string{"target-sr-cluster": clusterId}

	return &Client{
		APIClient: srsdk.NewAPIClient(cfg),
		authToken: authToken,
	}
}

func (c *Client) context() context.Context {
	return context.WithValue(context.Background(), srsdk.ContextAccessToken, c.authToken)
}

func (c *Client) GetTopLevelConfig() (srsdk.Config, error) {
	config, _, err := c.DefaultApi.GetTopLevelConfig(c.context())
	return config, err
}

func (c *Client) GetTopLevelMode() (srsdk.Mode, error) {
	mode, _, err := c.DefaultApi.GetTopLevelMode(c.context())
	return mode, err
}
