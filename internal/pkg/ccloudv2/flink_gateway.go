package ccloudv2

import (
	"context"
	"os/user"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/google/uuid"
)

type GatewayClientInterface interface {
	DeleteStatement(statementName string) error
	GetStatement(statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	ListStatements() ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	CreateStatement(statement string, properties map[string]string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	GetStatementResults(statementId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error)
}

type FlinkGatewayClient struct {
	*flinkgatewayv1alpha1.APIClient
	authToken      func() string
	envId          string
	orgResourceId  string
	kafkaClusterId string
	computePoolId  string
	identityPoolId string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken func() string, envId, orgResourceId, kafkaClusterId, computePoolId, identityPoolId string) *FlinkGatewayClient {
	cfg := flinkgatewayv1alpha1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = flinkgatewayv1alpha1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent
	cfg.DefaultHeader["Org-Id"] = orgResourceId
	return &FlinkGatewayClient{
		APIClient:      flinkgatewayv1alpha1.NewAPIClient(cfg),
		authToken:      authToken,
		envId:          envId,
		orgResourceId:  orgResourceId,
		kafkaClusterId: kafkaClusterId,
		computePoolId:  computePoolId,
		identityPoolId: identityPoolId,
	}
}

func (c *FlinkGatewayClient) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1alpha1.ContextAccessToken, c.authToken())
}

func (c *FlinkGatewayClient) DeleteStatement(statementName string) error {
	req := c.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), c.envId, statementName)
	r, err := c.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1StatementExecute(req)
	return errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatement(statementName string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), c.envId, statementName)
	resp, r, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1StatementExecute(req)
	return resp, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) ListStatements() ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	req := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), c.envId).PageSize(ccloudV2ListPageSize)
	resp, r, err := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1StatementsExecute(req)
	return resp.GetData(), errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) CreateStatement(statement string, properties map[string]string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	statementName := uuid.New().String()[:20]
	properties = c.propsDefault(properties)

	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName:  &statementName,
			Statement:      &statement,
			ComputePoolId:  &c.computePoolId,
			IdentityPoolId: &c.identityPoolId,
			Properties:     &properties,
		},
	}
	req := c.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), c.envId).SqlV1alpha1Statement(statementObj)
	createdStatement, r, err := req.Execute()
	return createdStatement, errors.CatchCCloudV2Error(err, r)
}

func (c *FlinkGatewayClient) GetStatementResults(statementId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	fetchResultsRequest := c.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(c.flinkGatewayApiContext(), c.envId, statementId)
	if pageToken != "" {
		fetchResultsRequest = fetchResultsRequest.PageToken(pageToken)
	}
	result, r, err := fetchResultsRequest.Execute()
	return result, errors.CatchCCloudV2Error(err, r)
}

// Set properties default values if not set by the user
// We probably want to refactor the keys names and where they are stored. Maybe also the default values.
func (c *FlinkGatewayClient) propsDefault(propsWithoutDefault map[string]string) map[string]string {
	properties := make(map[string]string)
	for key, value := range propsWithoutDefault {
		properties[key] = value
	}

	if _, ok := properties[config.ConfigKeyCatalog]; !ok {
		properties[config.ConfigKeyCatalog] = c.envId
	}
	if _, ok := properties[config.ConfigKeyDatabase]; !ok {
		properties[config.ConfigKeyDatabase] = c.kafkaClusterId
	}
	if _, ok := properties[config.ConfigKeyOrgResourceId]; !ok {
		properties[config.ConfigKeyOrgResourceId] = c.orgResourceId
	}
	if _, ok := properties[config.ConfigKeyExecutionRuntime]; !ok {
		properties[config.ConfigKeyExecutionRuntime] = "streaming"
	}

	currentUser, _ := user.Current()

	// TODO: consider removing configKeyStatementOwner when shipping to customers
	if _, ok := properties[config.ConfigKeyStatementOwner]; !ok && currentUser != nil {
		properties[config.ConfigKeyStatementOwner] = currentUser.Username
	}

	// Here we delete locally used properties before sending it to the backend
	delete(properties, config.ConfigKeyResultsTimeout)

	return properties
}
