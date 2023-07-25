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

func (c *Client) CreateExporter(req srsdk.CreateExporterRequest) (srsdk.CreateExporterResponse, error) {
	res, _, err := c.DefaultApi.CreateExporter(c.context(), req)
	return res, err
}

func (c *Client) DeleteExporter(name string) error {
	_, err := c.DefaultApi.DeleteExporter(c.context(), name)
	return err
}

func (c *Client) GetExporterInfo(name string) (srsdk.ExporterInfo, error) {
	info, _, err := c.DefaultApi.GetExporterInfo(c.context(), name)
	return info, err
}

func (c *Client) GetExporterConfig(name string) (map[string]string, error) {
	config, _, err := c.DefaultApi.GetExporterConfig(c.context(), name)
	return config, err
}

func (c *Client) ResumeExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResumeExporter(c.context(), name)
	return res, err
}

func (c *Client) ResetExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResetExporter(c.context(), name)
	return res, err
}

func (c *Client) PauseExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PauseExporter(c.context(), name)
	return res, err
}

func (c *Client) GetExporters() ([]string, error) {
	exporters, _, err := c.DefaultApi.GetExporters(c.context())
	return exporters, err
}

func (c *Client) GetExporterStatus(name string) (srsdk.ExporterStatus, error) {
	status, _, err := c.DefaultApi.GetExporterStatus(c.context(), name)
	return status, err
}

func (c *Client) PutExporter(name string, req srsdk.UpdateExporterRequest) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PutExporter(c.context(), name, req)
	return res, err
}
