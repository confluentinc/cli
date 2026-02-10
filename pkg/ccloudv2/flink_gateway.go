package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	flinkgatewayinternalv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1"
	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/errors/flink"
	flinkerror "github.com/confluentinc/cli/v4/pkg/errors/flink"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type GatewayClientInterface interface {
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1.SqlV1Statement, error)
	ListStatements(environmentId, orgId, computePoolId string) ([]flinkgatewayv1.SqlV1Statement, error)
	CreateStatement(statement flinkgatewayv1.SqlV1Statement, principal, environmentId, orgId string) (flinkgatewayv1.SqlV1Statement, error)
	GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1.SqlV1StatementResult, error)
	GetExceptions(environmentId, statementId, orgId string) ([]flinkgatewayv1.SqlV1StatementException, error)
	DeleteStatement(environmentId, statementName, orgId string) error
	UpdateStatement(environmentId, statementName, orgId string, statement flinkgatewayv1.SqlV1Statement) error
	GetAuthToken() string
	CreateConnection(connection flinkgatewayv1.SqlV1Connection, environmentId, orgId string) (flinkgatewayv1.SqlV1Connection, error)
	ListConnections(environmentId, orgId, connectionType string) ([]flinkgatewayv1.SqlV1Connection, error)
	GetConnection(environmentId, connectionName, orgId string) (flinkgatewayv1.SqlV1Connection, error)
	DeleteConnection(environmentId, connectionName, orgId string) error
	UpdateConnection(environmentId, connectionName, organizationId string, connection flinkgatewayv1.SqlV1Connection) error
}

type FlinkGatewayClient struct {
	*flinkgatewayv1.APIClient
	AuthToken string
}

type FlinkGatewayClientInternal struct {
	*flinkgatewayinternalv1.APIClient
	AuthToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClient {
	cfg := flinkgatewayv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClientWithRedirect(unsafeTrace, checkRedirect)
	cfg.Servers = flinkgatewayv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &FlinkGatewayClient{
		APIClient: flinkgatewayv1.NewAPIClient(cfg),
		AuthToken: authToken,
	}
}

func NewFlinkGatewayClientInternal(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClientInternal {
	cfg := flinkgatewayinternalv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClientWithRedirect(unsafeTrace, checkRedirect)
	cfg.Servers = flinkgatewayinternalv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &FlinkGatewayClientInternal{
		APIClient: flinkgatewayinternalv1.NewAPIClient(cfg),
		AuthToken: authToken,
	}
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	// Default net/http implementation allows 10 redirects - https://go.dev/src/net/http/client.go.
	// Lowered the redirect limit to fail fast in case of redirect cycles
	const maxRedirects = 5

	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}

	if len(via) > 0 {
		if authorization := via[len(via)-1].Header.Get("Authorization"); authorization != "" {
			log.CliLogger.Debugf("Following redirect with authorization to %s", req.URL)
			// Copy Authorization header from previous request as the location returned by
			// Flink GW on a redirect won't be a subdomain or exact match of initial domain
			// to be copied automatically.
			req.Header.Add("Authorization", authorization)
		}
	}

	return nil
}

func (c *FlinkGatewayClient) GetAuthToken() string {
	return c.AuthToken
}

func (c *FlinkGatewayClient) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1.ContextAccessToken, c.AuthToken)
}

func (c *FlinkGatewayClientInternal) flinkGatewayApiContextInternal() context.Context {
	return context.WithValue(context.Background(), flinkgatewayinternalv1.ContextAccessToken, c.AuthToken)
}

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	httpResp, err := c.StatementsSqlV1Api.DeleteSqlv1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1.SqlV1Statement, error) {
	resp, httpResp, err := c.StatementsSqlV1Api.GetSqlv1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListStatements(environmentId, orgId, computePoolId string) ([]flinkgatewayv1.SqlV1Statement, error) {
	var allStatements []flinkgatewayv1.SqlV1Statement
	pageToken := ""
	done := false
	for !done {
		statementListResponse, err := c.executeListStatements(environmentId, orgId, pageToken, computePoolId)
		if err != nil {
			return nil, err
		}
		allStatements = append(allStatements, statementListResponse.GetData()...)
		nextUrl := statementListResponse.Metadata.GetNext()
		pageToken, done, err = extractNextPageToken(flinkgatewayv1.NewNullableString(&nextUrl))
		if err != nil {
			return nil, err
		}
	}
	return allStatements, nil
}

func (c *FlinkGatewayClient) executeListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1.SqlV1StatementList, error) {
	req := c.StatementsSqlV1Api.ListSqlv1Statements(c.flinkGatewayApiContext(), orgId, environmentId).PageSize(ccloudV2ListPageSize)

	if computePoolId != "" {
		req = req.SpecComputePoolId(computePoolId)
	}

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) CreateStatement(statement flinkgatewayv1.SqlV1Statement, principal, environmentId, orgId string) (flinkgatewayv1.SqlV1Statement, error) {
	if principal != "" {
		statement.Spec.Principal = flinkgatewayv1.PtrString(principal)
	}
	resp, httpResp, err := c.StatementsSqlV1Api.CreateSqlv1Statement(c.flinkGatewayApiContext(), orgId, environmentId).SqlV1Statement(statement).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) UpdateStatement(environmentId, statementName, organizationId string, statement flinkgatewayv1.SqlV1Statement) error {
	httpResp, err := c.StatementsSqlV1Api.UpdateSqlv1Statement(c.flinkGatewayApiContext(), organizationId, environmentId, statementName).SqlV1Statement(statement).Execute()
	return flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementName, orgId, pageToken string) (flinkgatewayv1.SqlV1StatementResult, error) {
	req := c.StatementResultsSqlV1Api.GetSqlv1StatementResult(c.flinkGatewayApiContext(), orgId, environmentId, statementName)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetExceptions(environmentId, statementName, orgId string) ([]flinkgatewayv1.SqlV1StatementException, error) {
	resp, httpResp, err := c.StatementExceptionsSqlV1Api.GetSqlv1StatementExceptions(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	if err != nil {
		return nil, flinkerror.CatchError(err, httpResp)
	}
	return resp.GetData(), nil
}

func (c *FlinkGatewayClientInternal) DescribeMaterializedTable(environmentId, orgId, kafkaId, tableName string) (flinkgatewayinternalv1.SqlV1MaterializedTable, error) {
	resp, httpResp, err := c.MaterializedTablesSqlV1Api.GetSqlv1MaterializedTable(c.flinkGatewayApiContextInternal(), orgId, environmentId, kafkaId, tableName).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClientInternal) ListMaterializedTable(environmentId, orgId string) ([]flinkgatewayinternalv1.SqlV1MaterializedTable, error) {
	var allTables []flinkgatewayinternalv1.SqlV1MaterializedTable
	pageToken := ""
	done := false

	for !done {
		tableListResponse, err := c.executeMaterializedTablesList(environmentId, orgId, pageToken)
		if err != nil {
			return nil, err
		}
		allTables = append(allTables, tableListResponse.GetData()...)
		nextUrl := tableListResponse.Metadata.GetNext()
		pageToken, done, err = extractNextPageToken(flinkgatewayv1.NewNullableString(&nextUrl))
		if err != nil {
			return nil, err
		}
	}
	return allTables, nil
}

func (c *FlinkGatewayClientInternal) executeMaterializedTablesList(environmentId, orgId, pageToken string) (flinkgatewayinternalv1.SqlV1MaterializedTableList, error) {
	req := c.MaterializedTablesSqlV1Api.ListSqlv1MaterializedTables(c.flinkGatewayApiContextInternal(), orgId, environmentId).PageSize(ccloudV2ListPageSize)

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClientInternal) UpdateMaterializedTable(table flinkgatewayinternalv1.SqlV1MaterializedTable, environmentId, orgId, kafkaId, tableName string) (flinkgatewayinternalv1.SqlV1MaterializedTable, error) {
	resp, httpResp, err := c.MaterializedTablesSqlV1Api.UpdateSqlv1MaterializedTable(c.flinkGatewayApiContextInternal(), orgId, environmentId, kafkaId, tableName).SqlV1MaterializedTable(table).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClientInternal) DeleteMaterializedTable(environmentId, orgId, kafkaId, tableName string) error {
	httpResp, err := c.MaterializedTablesSqlV1Api.DeleteSqlv1MaterializedTable(c.flinkGatewayApiContextInternal(), orgId, environmentId, kafkaId, tableName).Execute()
	return flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) CreateConnection(connection flinkgatewayv1.SqlV1Connection, environmentId, orgId string) (flinkgatewayv1.SqlV1Connection, error) {
	resp, httpResp, err := c.ConnectionsSqlV1Api.CreateSqlv1Connection(c.flinkGatewayApiContext(), orgId, environmentId).SqlV1Connection(connection).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClientInternal) CreateMaterializedTable(table flinkgatewayinternalv1.SqlV1MaterializedTable, environmentId, orgId, kafkaId string) (flinkgatewayinternalv1.SqlV1MaterializedTable, error) {
	resp, httpResp, err := c.MaterializedTablesSqlV1Api.CreateSqlv1MaterializedTable(c.flinkGatewayApiContextInternal(), orgId, environmentId, kafkaId).SqlV1MaterializedTable(table).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) executeListConnection(environmentId, orgId, pageToken, connectionType string) (flinkgatewayv1.SqlV1ConnectionList, error) {
	req := c.ConnectionsSqlV1Api.ListSqlv1Connections(c.flinkGatewayApiContext(), orgId, environmentId).PageSize(ccloudV2ListPageSize)

	if connectionType != "" {
		req = req.SpecConnectionType(connectionType)
	}

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListConnections(environmentId, orgId, connectionType string) ([]flinkgatewayv1.SqlV1Connection, error) {
	var allConnections []flinkgatewayv1.SqlV1Connection
	pageToken := ""
	done := false
	for !done {
		connectionListResponse, err := c.executeListConnection(environmentId, orgId, pageToken, connectionType)
		if err != nil {
			return nil, err
		}
		allConnections = append(allConnections, connectionListResponse.GetData()...)
		nextUrl := connectionListResponse.Metadata.GetNext()
		pageToken, done, err = extractNextPageToken(flinkgatewayv1.NewNullableString(&nextUrl))
		if err != nil {
			return nil, err
		}
	}
	return allConnections, nil
}

func (c *FlinkGatewayClient) GetConnection(environmentId, connectionName, orgId string) (flinkgatewayv1.SqlV1Connection, error) {
	resp, httpResp, err := c.ConnectionsSqlV1Api.GetSqlv1Connection(c.flinkGatewayApiContext(), orgId, environmentId, connectionName).Execute()
	return resp, flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) DeleteConnection(environmentId, connectionName, orgId string) error {
	httpResp, err := c.ConnectionsSqlV1Api.DeleteSqlv1Connection(c.flinkGatewayApiContext(), orgId, environmentId, connectionName).Execute()
	return flinkerror.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) UpdateConnection(environmentId, connectionName, organizationId string, connection flinkgatewayv1.SqlV1Connection) error {
	httpResp, err := c.ConnectionsSqlV1Api.UpdateSqlv1Connection(c.flinkGatewayApiContext(), organizationId, environmentId, connectionName).SqlV1Connection(connection).Execute()
	return flink.CatchError(err, httpResp)
}
