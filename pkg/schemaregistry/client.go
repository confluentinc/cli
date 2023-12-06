package schemaregistry

import (
	"context"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type Client struct {
	*srsdk.APIClient
	apiKey srsdk.BasicAuth
	cfg    *config.Config
}

func NewClient(configuration *srsdk.Configuration, cfg *config.Config) *Client {
	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
		cfg:       cfg,
	}
}

func NewClientWithApiKey(configuration *srsdk.Configuration, apiKey srsdk.BasicAuth) *Client {
	return &Client{
		APIClient: srsdk.NewAPIClient(configuration),
		apiKey:    apiKey,
	}
}

func (c *Client) context() context.Context {
	ctx := context.Background()

	if c.apiKey.UserName != "" && c.apiKey.Password != "" {
		return context.WithValue(ctx, srsdk.ContextBasicAuth, c.apiKey)
	}

	if c.cfg.IsCloudLogin() {
		dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
		if err != nil {
			return ctx
		}
		return context.WithValue(ctx, srsdk.ContextAccessToken, dataplaneToken)
	} else if c.cfg.Context().GetState() != nil {
		return context.WithValue(ctx, srsdk.ContextAccessToken, c.cfg.Context().GetAuthToken())
	}

	return ctx
}

func (c *Client) Get() error {
	_, _, err := c.DefaultApi.Get(c.context()).Execute()
	return err
}

func (c *Client) GetTopLevelConfig() (srsdk.Config, error) {
	res, _, err := c.DefaultApi.GetTopLevelConfig(c.context()).Execute()
	return res, err
}

func (c *Client) DeleteTopLevelConfig() (string, error) {
	res, _, err := c.DefaultApi.DeleteTopLevelConfig(c.context()).Execute()
	return res, err
}

func (c *Client) UpdateTopLevelConfig(req srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateTopLevelConfig(c.context()).ConfigUpdateRequest(req).Execute()
	return res, err
}

func (c *Client) GetTopLevelMode() (srsdk.Mode, error) {
	res, _, err := c.DefaultApi.GetTopLevelMode(c.context()).Execute()
	return res, err
}

func (c *Client) UpdateTopLevelMode(req srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateTopLevelMode(c.context()).Body(req).Execute()
	return res, err
}

func (c *Client) UpdateMode(subject string, req srsdk.ModeUpdateRequest) (srsdk.ModeUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateMode(c.context(), subject).Body(req).Execute()
	return res, err
}

func (c *Client) GetSubjectLevelConfig(subject string) (srsdk.Config, error) {
	res, _, err := c.DefaultApi.GetSubjectLevelConfig(c.context(), subject).Execute()
	return res, err
}

func (c *Client) UpdateSubjectLevelConfig(subject string, req srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, error) {
	res, _, err := c.DefaultApi.UpdateSubjectLevelConfig(c.context(), subject).Body(req).Execute()
	return res, err
}

func (c *Client) DeleteSubjectLevelConfig(subject string) (string, error) {
	res, _, err := c.DefaultApi.DeleteSubjectConfig(c.context(), subject).Execute()
	return res, err
}

func (c *Client) TestCompatibilityBySubjectName(subject, version string, req srsdk.RegisterSchemaRequest) (srsdk.CompatibilityCheckResponse, error) {
	res, _, err := c.DefaultApi.TestCompatibilityBySubjectName(c.context(), subject, version).Body(req).Execute()
	return res, err
}

func (c *Client) CreateExporter(req srsdk.CreateExporterRequest) (srsdk.CreateExporterResponse, error) {
	res, _, err := c.DefaultApi.CreateExporter(c.context()).Body(req).Execute()
	return res, err
}

func (c *Client) DeleteExporter(name string) error {
	_, err := c.DefaultApi.DeleteExporter(c.context(), name).Execute()
	return err
}

func (c *Client) GetExporterInfo(name string) (srsdk.ExporterInfo, error) {
	res, _, err := c.DefaultApi.GetExporterInfo(c.context(), name).Execute()
	return res, err
}

func (c *Client) GetExporterConfig(name string) (map[string]string, error) {
	res, _, err := c.DefaultApi.GetExporterConfig(c.context(), name).Execute()
	return res, err
}

func (c *Client) ResumeExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResumeExporter(c.context(), name).Execute()
	return res, err
}

func (c *Client) ResetExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.ResetExporter(c.context(), name).Execute()
	return res, err
}

func (c *Client) PauseExporter(name string) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PauseExporter(c.context(), name).Execute()
	return res, err
}

func (c *Client) GetExporters() ([]string, error) {
	res, _, err := c.DefaultApi.GetExporters(c.context()).Execute()
	return res, err
}

func (c *Client) GetExporterStatus(name string) (srsdk.ExporterStatus, error) {
	res, _, err := c.DefaultApi.GetExporterStatus(c.context(), name).Execute()
	return res, err
}

func (c *Client) PutExporter(name string, req srsdk.UpdateExporterRequest) (srsdk.UpdateExporterResponse, error) {
	res, _, err := c.DefaultApi.PutExporter(c.context(), name).Body(req).Execute()
	return res, err
}

func (c *Client) Register(subject string, req srsdk.RegisterSchemaRequest, normalize bool) (srsdk.RegisterSchemaResponse, error) {
	res, _, err := c.DefaultApi.Register(c.context(), subject).Body(req).Normalize(normalize).Execute()
	return res, err
}

func (c *Client) GetSchema(id int32, subject string) (srsdk.SchemaString, error) {
	res, _, err := c.DefaultApi.GetSchema(c.context(), id).Subject(subject).Execute()
	return res, err
}

func (c *Client) GetSchemaByVersion(subject, version string, deleted bool) (srsdk.Schema, error) {
	res, _, err := c.DefaultApi.GetSchemaByVersion(c.context(), subject, version).Deleted(deleted).Execute()
	return res, err
}

func (c *Client) ListVersions(subject string, deleted bool) ([]int32, error) {
	res, _, err := c.DefaultApi.ListVersions(c.context(), subject).Deleted(deleted).Execute()
	return res, err
}

func (c *Client) DeleteSchemaVersion(subject, version string, permanent bool) (int32, error) {
	res, _, err := c.DefaultApi.DeleteSchemaVersion(c.context(), subject, version).Permanent(permanent).Execute()
	return res, err
}

func (c *Client) GetSchemas(subjectPrefix string, deleted bool) ([]srsdk.Schema, error) {
	res, _, err := c.DefaultApi.GetSchemas(c.context()).SubjectPrefix(subjectPrefix).Deleted(deleted).Execute()
	return res, err
}

func (c *Client) DeleteSubject(subject string, permanent bool) ([]int32, error) {
	res, _, err := c.DefaultApi.DeleteSubject(c.context(), subject).Permanent(permanent).Execute()
	return res, err
}

func (c *Client) PartialUpdateByUniqueAttributes(atlasEntity srsdk.AtlasEntityWithExtInfo) error {
	_, err := c.DefaultApi.PartialUpdateByUniqueAttributes(c.context()).AtlasEntityWithExtInfo(atlasEntity).Execute()
	return err
}

func (c *Client) CreateTags(tag []srsdk.Tag) ([]srsdk.TagResponse, error) {
	res, _, err := c.DefaultApi.CreateTags(c.context()).Tag(tag).Execute()
	return res, err
}

func (c *Client) CreateTagDefs(tagDef []srsdk.TagDef) ([]srsdk.TagDefResponse, error) {
	res, _, err := c.DefaultApi.CreateTagDefs(c.context()).TagDef(tagDef).Execute()
	return res, err
}

func (c *Client) GetTags(typeName, qualifiedName string) ([]srsdk.TagResponse, error) {
	res, _, err := c.DefaultApi.GetTags(c.context(), typeName, qualifiedName).Execute()
	return res, err
}

func (c *Client) List(subjectPrefix string, deleted bool) ([]string, error) {
	res, _, err := c.DefaultApi.List(c.context()).SubjectPrefix(subjectPrefix).Deleted(deleted).Execute()
	return res, err
}

func (c *Client) GetByUniqueAttributes(typeName, qualifiedName string) (srsdk.AtlasEntityWithExtInfo, error) {
	res, _, err := c.DefaultApi.GetByUniqueAttributes(c.context(), typeName, qualifiedName).Execute()
	return res, err
}

// kek first.

func (c *Client) CreateKek(name string, createReq srsdk.CreateKekRequest) (srsdk.Kek, error) {
	res, _, err := c.DefaultApi.CreateKek(c.context()).CreateKekRequest(createReq).Execute()
	return res, err
}

func (c *Client) DeleteKek(name string, permanent bool) error {
	_, err := c.DefaultApi.DeleteKek(c.context(), name).Permanent(permanent).Execute()
	return err
}

func (c *Client) ListKeks(deleted bool) ([]string, error) {
	res, _, err := c.DefaultApi.GetKekNames(c.context()).Deleted(deleted).Execute() // no page token?
	return res, err
}

func (c *Client) DescribeKek(name string, deleted bool) (srsdk.Kek, error) {
	res, _, err := c.DefaultApi.GetKek(c.context(), name).Deleted(deleted).Execute()
	return res, err
}

func (c *Client) UpdateKek(name string, updateReq srsdk.UpdateKekRequest) (srsdk.Kek, error) {
	res, _, err := c.DefaultApi.PutKek(c.context(), name).UpdateKekRequest(updateReq).Execute()
	return res, err
}

func (c *Client) CreateDek(name string, createReq srsdk.CreateDekRequest) (srsdk.Dek, error) {
	res, _, err := c.DefaultApi.CreateDek(c.context(), name).CreateDekRequest(createReq).Execute()
	return res, err
}

func (c *Client) DeleteDekVersion(name, subject, version string) error {
	_, err := c.DefaultApi.DeleteDekVersion(c.context(), name, subject, version).Execute()
	return err
}

func (c *Client) DeleteDekVersions(name, subject string) error {
	_, err := c.DefaultApi.DeleteDekVersions(c.context(), name, subject).Execute()
	return err
}

func (c *Client) GetDek(name, subject string) (srsdk.Dek, error) {
	res, _, err := c.DefaultApi.GetDek(c.context(), name, subject).Execute()
	return res, err
}

func (c *Client) GetDekByVersion(name, subject, version string) (srsdk.Dek, error) {
	res, _, err := c.DefaultApi.GetDekByVersion(c.context(), name, subject, version).Execute()
	return res, err
}

func (c *Client) GetDeKVersions(name, subject string) ([]int32, error) {
	res, _, err := c.DefaultApi.GetDekVersions(c.context(), name, subject).Execute()
	return res, err
}

func (c *Client) GetDekSubjects(name string) ([]string, error) {
	res, _, err := c.DefaultApi.GetDekSubjects(c.context(), name).Execute()
	return res, err
}

func (c *Client) UndeleteDekVersion(name, subject, version string) error {
	_, err := c.DefaultApi.UndeleteDekVersion(c.context(), name, subject, version).Execute()
	return err
}
