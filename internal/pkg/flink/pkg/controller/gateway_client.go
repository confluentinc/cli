package controller

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
)

type GatewayClientInterface interface {
	CreateStatement(ctx context.Context, statement string, properties map[string]string) (v1.SqlV1alpha1Statement, *http.Response, error)
	DeleteStatement(ctx context.Context, statementName string) (*http.Response, error)
	GetStatement(ctx context.Context, statementName string) (v1.SqlV1alpha1Statement, *http.Response, error)
	GetStatementResults(ctx context.Context, statementId, pageToken string) (v1.SqlV1alpha1StatementResult, *http.Response, error)
}

type GatewayClient struct {
	authToken         string
	envId             string
	orgResourceId     string
	kafkaClusterId    string
	computePoolId     string
	defaultProperties map[string]string
	client            *v1.APIClient
}

func (c *GatewayClient) CreateStatement(ctx context.Context, statement string, properties map[string]string) (v1.SqlV1alpha1Statement, *http.Response, error) {
	statementName := uuid.New().String()[:20]
	properties = c.propsDefault(properties)

	statementObj := v1.SqlV1alpha1Statement{
		Spec: &v1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
			Statement:     &statement,
			ComputePoolId: &c.computePoolId,
			Properties:    &properties,
		},
	}

	ctx = context.WithValue(ctx, v1.ContextAccessToken, c.authToken)
	req := c.client.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(ctx, c.envId).SqlV1alpha1Statement(statementObj)

	createdStatement, resp, err := req.Execute()

	return createdStatement, resp, err
}

func (c *GatewayClient) GetStatement(ctx context.Context, statementName string) (v1.SqlV1alpha1Statement, *http.Response, error) {
	ctx = context.WithValue(ctx, v1.ContextAccessToken, c.authToken)
	createdStatement, resp, err := c.client.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(ctx, c.envId, statementName).Execute()

	return createdStatement, resp, err
}

func (c *GatewayClient) DeleteStatement(ctx context.Context, statementName string) (*http.Response, error) {
	ctx = context.WithValue(ctx, v1.ContextAccessToken, c.authToken)
	resp, err := c.client.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(ctx, c.envId, statementName).Execute()

	return resp, err
}

func (c *GatewayClient) GetStatementResults(ctx context.Context, statementId, pageToken string) (v1.SqlV1alpha1StatementResult, *http.Response, error) {
	fetchResultsRequest := c.client.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(ctx, c.envId, statementId)
	if pageToken != "" {
		fetchResultsRequest = fetchResultsRequest.PageToken(pageToken)
	}
	result, resp, err := fetchResultsRequest.Execute()
	return result, resp, err
}

// Set properties default values if not set by the user
// We probably want to refactor the keys names and where they are stored. Maybe also the default values.
func (c *GatewayClient) propsDefault(properties map[string]string) map[string]string {
	if _, ok := properties[configKeyCatalog]; !ok {
		properties[configKeyCatalog] = c.envId
	}
	if _, ok := properties[configKeyDatabase]; !ok {
		properties[configKeyDatabase] = c.kafkaClusterId
	}
	if _, ok := properties[configKeyOrgResourceId]; !ok {
		properties[configKeyOrgResourceId] = c.orgResourceId
	}
	if _, ok := properties[configKeyExecutionRuntime]; !ok {
		properties[configKeyExecutionRuntime] = "streaming"
	}

	return properties
}

func NewGatewayClient(envId, orgResourceId, kafkaClusterId, computePoolId, authToken string, appOptions *ApplicationOptions) GatewayClientInterface {
	cfg := v1.NewConfiguration()
	unsafeTrace := false
	defaultProperties := make(map[string]string)
	if appOptions != nil {
		if appOptions.HTTP_CLIENT_UNSAFE_TRACE {
			unsafeTrace = true
		}
		if appOptions.FLINK_GATEWAY_URL != "" {
			cfg.Servers = v1.ServerConfigurations{{URL: appOptions.FLINK_GATEWAY_URL}}
		}
	}

	cfg.Debug = unsafeTrace

	client := &GatewayClient{
		envId:             envId,
		orgResourceId:     orgResourceId,
		kafkaClusterId:    kafkaClusterId,
		computePoolId:     computePoolId,
		authToken:         authToken,
		defaultProperties: defaultProperties,
		client:            v1.NewAPIClient(cfg),
	}

	return client
}
