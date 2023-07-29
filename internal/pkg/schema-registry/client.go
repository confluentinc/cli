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

func NewClientWithApiKey(configuration *srsdk.Configuration, key, secret string) *Client {
	basicAuth := srsdk.BasicAuth{
		UserName: key,
		Password: secret,
	}

	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
		context:   context.WithValue(context.Background(), srsdk.ContextBasicAuth, basicAuth),
	}
}

func (c *Client) Get() error {
	_, _, err := c.DefaultApi.Get(c.context)
	return err
}

func (c *Client) GetTopLevelConfig() (srsdk.Config, error) {
	res, _, err := c.DefaultApi.GetTopLevelConfig(c.context)
	return res, err
}

func (c *Client) UpdateTopLevelConfig(res srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateTopLevelConfig(c.context, res)
	return res, err
}

func (c *Client) GetTopLevelMode() (srsdk.Mode, error) {
	res, _, err := c.DefaultApi.GetTopLevelMode(c.context)
	return res, err
}

func (c *Client) UpdateTopLevelMode(res srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateTopLevelMode(c.context, res)
	return res, err
}

func (c *Client) UpdateMode(subject string, req srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateMode(c.context, subject, req)
	return res, err
}

func (c *Client) GetSubjectLevelConfig(subject string) (srsdk.Config, error) {
	res, _, err := c.DefaultApi.GetSubjectLevelConfig(c.context, subject, nil)
	return res, err
}

func (c *Client) UpdateSubjectLevelConfig(subject string, res srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateSubjectLevelConfig(c.context, subject, res)
	return res, err
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
	res, _, err := c.DefaultApi.GetExporterInfo(c.context, name)
	return res, err
}

func (c *Client) GetExporterConfig(name string) (map[string]string, error) {
	res, _, err := c.DefaultApi.GetExporterConfig(c.context, name)
	return res, err
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
	res, _, err := c.DefaultApi.GetExporters(c.context)
	return res, err
}

func (c *Client) GetExporterStatus(name string) (srsdk.ExporterStatus, error) {
	res, _, err := c.DefaultApi.GetExporterStatus(c.context, name)
	return res, err
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
	res, _, err := c.DefaultApi.GetSchema(c.context, id, opts)
	return res, err
}

func (c *Client) GetSchemaByVersion(subject, version string, opts *srsdk.GetSchemaByVersionOpts) (srsdk.Schema, error) {
	res, _, err := c.DefaultApi.GetSchemaByVersion(c.context, subject, version, opts)
	return res, err
}

func (c *Client) ListVersions(subject string, opts *srsdk.ListVersionsOpts) ([]int32, error) {
	res, _, err := c.DefaultApi.ListVersions(c.context, subject, opts)
	return res, err
}

func (c *Client) DeleteSchemaVersion(subject, version string, opts *srsdk.DeleteSchemaVersionOpts) (int32, error) {
	res, _, err := c.DefaultApi.DeleteSchemaVersion(c.context, subject, version, opts)
	return res, err
}

func (c *Client) GetSchemas(opts *srsdk.GetSchemasOpts) ([]srsdk.Schema, error) {
	res, _, err := c.DefaultApi.GetSchemas(c.context, opts)
	return res, err
}

func (c *Client) DeleteSubject(subject string, opts *srsdk.DeleteSubjectOpts) ([]int32, error) {
	res, _, err := c.DefaultApi.DeleteSubject(c.context, subject, opts)
	return res, err
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

func (c *Client) List(opts *srsdk.ListOpts) ([]string, error) {
	res, _, err := c.DefaultApi.List(c.context, opts)
	return res, err
}

func (c *Client) GetByUniqueAttributes(typeName, qualifiedName string) (srsdk.AtlasEntityWithExtInfo, error) {
	res, _, err := c.DefaultApi.GetByUniqueAttributes(c.context, typeName, qualifiedName, nil)
	return res, err
}

func (c *Client) AsyncapiPut() error {
	_, err := c.DefaultApi.AsyncapiPut(c.context)
	return err
}
