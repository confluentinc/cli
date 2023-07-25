package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

type Client struct {
	*srsdk.APIClient
	context context.Context
}

func NewClient(configuration *srsdk.Configuration, authToken string) *Client {
	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
		context:   context.WithValue(context.Background(), srsdk.ContextAccessToken, authToken),
	}
}

func NewClientWithApiKey(configuration *srsdk.Configuration, apiKey string) *Client {
	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
		context:   context.WithValue(context.Background(), srsdk.ContextAPIKey, apiKey),
	}
}

func (c *Client) GetTopLevelConfig() (srsdk.Config, error) {
	config, _, err := c.DefaultApi.GetTopLevelConfig(c.context)
	return config, err
}

func (c *Client) UpdateTopLevelConfig(req srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	req, _, err := c.DefaultApi.UpdateTopLevelConfig(c.context, req)
	return req, err
}

func (c *Client) GetTopLevelMode() (srsdk.Mode, error) {
	mode, _, err := c.DefaultApi.GetTopLevelMode(c.context)
	return mode, err
}

func (c *Client) UpdateTopLevelMode(req srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	req, _, err := c.DefaultApi.UpdateTopLevelMode(c.context, req)
	return req, err
}

func (c *Client) GetSubjectLevelConfig(subject string) (srsdk.Config, error) {
	config, _, err := c.DefaultApi.GetSubjectLevelConfig(c.context, subject, nil)
	return config, err
}

func (c *Client) UpdateSubjectLevelConfig(subject string, req srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	req, _, err := c.DefaultApi.UpdateSubjectLevelConfig(c.context, subject, req)
	return req, err
}

func (c *Client) TestCompatibilityBySubjectName(subject, version string, body srsdk.RegisterSchemaRequest) (srsdk.CompatibilityCheckResponse, error) {
	res, _, err := c.DefaultApi.TestCompatibilityBySubjectName(c.context, subject, version, body, nil)
	return res, err
}

func (c *Client) CreateExporter(req srsdk.CreateExporterRequest) (srsdk.CreateExporterResponse, error) {
	res, _, err := c.DefaultApi.CreateExporter(c.context, req)
	return res, err
}

func (c *Client) DeleteExporter(name string) error {
	_, err := c.DefaultApi.DeleteExporter(c.context, name)
	return err
}

func (c *Client) GetExporterInfo(name string) (srsdk.ExporterInfo, error) {
	info, _, err := c.DefaultApi.GetExporterInfo(c.context, name)
	return info, err
}

func (c *Client) GetExporterConfig(name string) (map[string]string, error) {
	config, _, err := c.DefaultApi.GetExporterConfig(c.context, name)
	return config, err
}

func (c *Client) ResumeExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResumeExporter(c.context, name)
	return res, err
}

func (c *Client) ResetExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResetExporter(c.context, name)
	return res, err
}

func (c *Client) PauseExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PauseExporter(c.context, name)
	return res, err
}

func (c *Client) GetExporters() ([]string, error) {
	exporters, _, err := c.DefaultApi.GetExporters(c.context)
	return exporters, err
}

func (c *Client) GetExporterStatus(name string) (srsdk.ExporterStatus, error) {
	status, _, err := c.DefaultApi.GetExporterStatus(c.context, name)
	return status, err
}

func (c *Client) PutExporter(name string, req srsdk.UpdateExporterRequest) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PutExporter(c.context, name, req)
	return res, err
}

func (c *Client) Register(subject string, req srsdk.RegisterSchemaRequest, opts *srsdk.RegisterOpts) (srsdk.RegisterSchemaResponse, error) {
	res, _, err := c.DefaultApi.Register(c.context, subject, req, opts)
	return res, err
}

func (c *Client) GetSchema(id int32, opts *srsdk.GetSchemaOpts) (srsdk.SchemaString, error) {
	schema, _, err := c.DefaultApi.GetSchema(c.context, id, opts)
	return schema, err
}

func (c *Client) GetSchemaByVersion(subject, version string) (srsdk.Schema, error) {
	schema, _, err := c.DefaultApi.GetSchemaByVersion(c.context, subject, version, nil)
	return schema, err
}

func (c *Client) PartialUpdateByUniqueAttributes(opts *srsdk.PartialUpdateByUniqueAttributesOpts) error {
	_, err := c.DefaultApi.PartialUpdateByUniqueAttributes(c.context, opts)
	return err
}

func (c *Client) CreateTags(opts *srsdk.CreateTagsOpts) ([]srsdk.TagResponse, error) {
	res, _, err := c.DefaultApi.CreateTags(c.context, opts)
	return res, err
}

func (c *Client) CreateTagDefs(opts *srsdk.CreateTagDefsOpts) ([]srsdk.TagDefResponse, error) {
	res, _, err := c.DefaultApi.CreateTagDefs(c.context, opts)
	return res, err
}

func (c *Client) GetTags(typeName, qualifiedName string) ([]srsdk.TagResponse, error) {
	res, _, err := c.DefaultApi.GetTags(c.context, typeName, qualifiedName)
	return res, err
}

func (c *Client) List() ([]string, error) {
	subjects, _, err := c.DefaultApi.List(c.context, nil)
	return subjects, err
}

func (c *Client) GetByUniqueAttributes(typeName, qualifiedName string) (srsdk.AtlasEntityWithExtInfo, error) {
	res, _, err := c.DefaultApi.GetByUniqueAttributes(c.context, typeName, qualifiedName, nil)
	return res, err
}

func (c *Client) AsyncapiPut() error {
	_, err := c.DefaultApi.AsyncapiPut(c.context)
	return err
}
