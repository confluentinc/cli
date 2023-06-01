package ccloudv2

import (
	"context"
	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/google/uuid"
)

type GatewayClientInterface interface {
	DeleteStatement(environmentId, statementName string) error
	GetStatement(environmentId, statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	ListStatements(environmentId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	CreateStatement(environmentId, computePoolId, identityPoolId, statement string, properties map[string]string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	GetStatementResults(environmentId, statementId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error)
}

type FlinkGatewayClient struct {
	*flinkgatewayv1alpha1.APIClient
	authToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken, orgResourceId string) *FlinkGatewayClient {
	cfg := flinkgatewayv1alpha1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkgatewayv1alpha1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent
	cfg.DefaultHeader["Org-Id"] = orgResourceId

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

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, r, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).Execute()
	return resp, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) ListStatements(environmentId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, r, err := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId).PageSize(ccloudV2ListPageSize).Execute()
	return resp.GetData(), errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) CreateStatement(environmentId, computePoolId, identityPoolId, statement string, properties map[string]string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
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
	createdStatement, r, err := c.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId).SqlV1alpha1Statement(statementObj).Execute()
	return createdStatement, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	fetchResultsRequest := c.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(c.flinkGatewayApiContext(), environmentId, statementId)
	if pageToken != "" {
		fetchResultsRequest = fetchResultsRequest.PageToken(pageToken)
	}
	result, r, err := fetchResultsRequest.Execute()
	return result, errors.CatchCCloudV2Error(err, r)
}
