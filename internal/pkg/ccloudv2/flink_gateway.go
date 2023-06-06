package ccloudv2

import (
	"context"

	"github.com/google/uuid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type GatewayClientInterface interface {
	DeleteStatement(environmentId, statementName, orgId string) error
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	ListStatements(environmentId, orgId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	CreateStatement(statement, computePoolId, identityPoolId string, properties map[string]string, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error)
	GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error)
}

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

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	r, err := c.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, r, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return resp, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) ListStatements(environmentId, orgId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, r, err := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId).OrgId(orgId).PageSize(ccloudV2ListPageSize).Execute()
	return resp.GetData(), errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) CreateStatement(statement, computePoolId, identityPoolId string, properties map[string]string, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	statementName := uuid.New().String()[:20]

	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName:  &statementName,
			Statement:      &statement,
			ComputePoolId:  &computePoolId,
			IdentityPoolId: &identityPoolId,
			Properties:     &properties,
		},
	}
	res, r, err := c.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId).SqlV1alpha1Statement(statementObj).OrgId(orgId).Execute()
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	req := c.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	res, r, err := req.Execute()
	return res, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error) {
	result, r, err := c.StatementExceptionsSqlV1alpha1Api.GetSqlV1alpha1StatementExceptions(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId).Execute()
	return result, errors.CatchCCloudV2Error(err, r)
}
