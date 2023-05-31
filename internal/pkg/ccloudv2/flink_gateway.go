package ccloudv2

import (
	"context"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type FlinkGatewayClient struct {
	*flinkgatewayv1alpha1.APIClient
	authToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClient {
	cfg := flinkgatewayv1alpha1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkgatewayv1alpha1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &FlinkGatewayClient{
		APIClient: flinkgatewayv1alpha1.NewAPIClient(cfg),
		authToken: authToken,
	}
}

func (c *FlinkGatewayClient) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1alpha1.ContextAccessToken, c.authToken)
}

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName string) error {
	r, err := c.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).Execute()
	return errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	res, r, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) ListStatements(environmentId, orgId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	res, r, err := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId).OrgId(orgId).PageSize(ccloudV2ListPageSize).Execute()
	return res.GetData(), errors.CatchCCloudV2Error(err, r)
}
