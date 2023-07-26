package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

type Client struct {
	*srsdk.APIClient
	authToken string
}

func NewClient(configuration *srsdk.Configuration, authToken string) *Client {
	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
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

func (c *Client) UpdateTopLevelConfig(req srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	req, _, err := c.DefaultApi.UpdateTopLevelConfig(c.context(), req)
	return req, err
}

func (c *Client) GetTopLevelMode() (srsdk.Mode, error) {
	mode, _, err := c.DefaultApi.GetTopLevelMode(c.context())
	return mode, err
}

func (c *Client) UpdateTopLevelMode(req srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	req, _, err := c.DefaultApi.UpdateTopLevelMode(c.context(), req)
	return req, err
}

func (c *Client) GetSubjectLevelConfig(subject string) (srsdk.Config, error) {
	config, _, err := c.DefaultApi.GetSubjectLevelConfig(c.context(), subject, nil)
	return config, err
}

func (c *Client) TestCompatibilityBySubjectName(subject, version string, body srsdk.RegisterSchemaRequest) (srsdk.CompatibilityCheckResponse, error) {
	res, _, err := c.DefaultApi.TestCompatibilityBySubjectName(c.context(), subject, version, body, nil)
	return res, err
}
