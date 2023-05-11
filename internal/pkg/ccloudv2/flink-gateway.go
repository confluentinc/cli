package ccloudv2

import (
	"context"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newFlinkGatewayClient(url, userAgent string, unsafeTrace bool) *flinkgatewayv1alpha1.APIClient {
	cfg := flinkgatewayv1alpha1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkgatewayv1alpha1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return flinkgatewayv1alpha1.NewAPIClient(cfg)
}

func (c *Client) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1alpha1.ContextAccessToken, c.AuthToken)
}

func (c *Client) DeleteStatement(environmentId, statementName string) error {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName)
	r, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1StatementExecute(req)
	return errors.CatchCCloudV2Error(err, r)
}

func (c *Client) GetStatement(environmentId, statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName)
	resp, r, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.GetSqlV1alpha1StatementExecute(req)
	return resp, errors.CatchCCloudV2Error(err, r)
}

func (c *Client) ListStatements(environmentId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId).PageSize(ccloudV2ListPageSize)
	resp, r, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.ListSqlV1alpha1StatementsExecute(req)
	return resp.GetData(), errors.CatchCCloudV2Error(err, r)
}
