package ccloudv2

import (
	"context"
	"net/http"

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

func (c *Client) CreateSqlStatement(environmentId string, sqlStatement flinkgatewayv1alpha1.SqlV1alpha1Statement) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId).SqlV1alpha1Statement(sqlStatement)
	resp, httpResp, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.CreateSqlV1alpha1StatementExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSqlStatement(environmentId, statementName string) error {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName)
	httpResp, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1StatementExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSqlStatement(environmentId, statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName)
	resp, httpResp, err := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.GetSqlV1alpha1StatementExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSqlStatements(environmentId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	var list []flinkgatewayv1alpha1.SqlV1alpha1Statement

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSqlStatements(pageToken, environmentId)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListSqlStatements(pageToken, environmentId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementList, *http.Response, error) {
	req := c.FlinkGatewayClient.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	req = req.PageSize(ccloudV2ListPageSize)
	return c.FlinkGatewayClient.StatementsSqlV1alpha1Api.ListSqlV1alpha1StatementsExecute(req)
}
