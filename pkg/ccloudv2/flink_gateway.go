package ccloudv2

import (
	"context"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	"github.com/confluentinc/cli/v3/pkg/errors/flink"
)

type GatewayClientInterface interface {
	DeleteStatement(environmentId, statementName, orgId string) error
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error)
	ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1beta1.SqlV1beta1StatementList, error)
	CreateStatement(statement flinkgatewayv1beta1.SqlV1beta1Statement, serviceAccountId, identityPoolId, environmentId, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error)
	GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1beta1.SqlV1beta1StatementResult, error)
	GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1beta1.SqlV1beta1StatementExceptionList, error)
}

type FlinkGatewayClient struct {
	*flinkgatewayv1beta1.APIClient
	AuthToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClient {
	cfg := flinkgatewayv1beta1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkgatewayv1beta1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &FlinkGatewayClient{
		APIClient: flinkgatewayv1beta1.NewAPIClient(cfg),
		AuthToken: authToken,
	}
}

func (c *FlinkGatewayClient) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1beta1.ContextAccessToken, c.AuthToken)
}

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	httpResp, err := c.StatementsSqlV1beta1Api.DeleteSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error) {
	resp, httpResp, err := c.StatementsSqlV1beta1Api.GetSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1beta1.SqlV1beta1StatementList, error) {
	req := c.StatementsSqlV1beta1Api.ListSqlv1beta1Statements(c.flinkGatewayApiContext(), orgId, environmentId).PageSize(ccloudV2ListPageSize)

	if computePoolId != "" {
		req = req.SpecComputePoolId(computePoolId)
	}

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListAllStatements(environmentId, orgId, computePoolId string) ([]flinkgatewayv1beta1.SqlV1beta1Statement, error) {
	var allStatements []flinkgatewayv1beta1.SqlV1beta1Statement
	pageToken := ""
	done := false
	for !done {
		statementListResponse, err := c.ListStatements(environmentId, orgId, pageToken, computePoolId)
		if err != nil {
			return nil, err
		}
		allStatements = append(allStatements, statementListResponse.GetData()...)
		nextUrl := statementListResponse.Metadata.GetNext()
		pageToken, done, err = extractNextPageToken(flinkgatewayv1beta1.NewNullableString(&nextUrl))
		if err != nil {
			return nil, err
		}
	}
	return allStatements, nil
}

func (c *FlinkGatewayClient) CreateStatement(statement flinkgatewayv1beta1.SqlV1beta1Statement, serviceAccountId, identityPoolId, environmentId, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error) {
	if serviceAccountId != "" {
		// add the service account header and remove it after the request
		statement.Principal = &serviceAccountId
	} else {
		// if this is also empty we have the interactive query/AUMP case
		statement.Spec.IdentityPoolId = &identityPoolId
	}
	resp, httpResp, err := c.StatementsSqlV1beta1Api.CreateSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId).SqlV1beta1Statement(statement).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementName, orgId, pageToken string) (flinkgatewayv1beta1.SqlV1beta1StatementResult, error) {
	req := c.StatementResultSqlV1beta1Api.GetSqlv1beta1StatementResult(c.flinkGatewayApiContext(), orgId, environmentId, statementName)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetExceptions(environmentId, statementName, orgId string) (flinkgatewayv1beta1.SqlV1beta1StatementExceptionList, error) {
	resp, httpResp, err := c.StatementExceptionsSqlV1beta1Api.GetSqlv1beta1StatementExceptions(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return resp, flink.CatchError(err, httpResp)
}
