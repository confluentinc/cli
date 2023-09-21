package ccloudv2

import (
	"context"
	"errors"
	"net/http"

	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	"github.com/confluentinc/cli/v3/pkg/errors/flink"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type GatewayClientInterface interface {
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error)
	ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1beta1.SqlV1beta1StatementList, error)
	CreateStatement(statement flinkgatewayv1beta1.SqlV1beta1Statement, principal, environmentId, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error)
	GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1beta1.SqlV1beta1StatementResult, error)
	GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1beta1.SqlV1beta1StatementExceptionList, error)
	DeleteStatement(environmentId, statementName, orgId string) error
	UpdateStatement(environmentId, statementName, orgId string, statement flinkgatewayv1beta1.SqlV1beta1Statement) error
}

type FlinkGatewayClient struct {
	*flinkgatewayv1beta1.APIClient
	AuthToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClient {
	cfg := flinkgatewayv1beta1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClientWithRedirect(unsafeTrace,
		func(req *http.Request, via []*http.Request) error {
			// Default net/http implementation allows 10 redirects - https://go.dev/src/net/http/client.go.
			// Lowered the redirect limit to fail fast in case of redirect cycles
			if len(via) >= 5 {
				return errors.New("stopped after 5 redirects")
			}
			log.CliLogger.Debugf("Following redirect with authorization to %s", req.URL)
			// Customize the redirect to add authorization header on 307 Redirect.
			// This is required as the Location header returned by the gateway may not be an exact subdomain and Authorization
			// header is copied only when the hostname is exactly the same or an exact subdomain of the hostname
			// Ideally, we should check the status code returned in via[0].Response. However, via[0].Response is nil in underlying
			// retryable http implementation
			req.Header.Add("Authorization", "Bearer "+authToken)

			return nil
		})
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

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error) {
	resp, httpResp, err := c.StatementsSqlV1beta1Api.GetSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) UpdateStatement(environmentId, statementName, orgId string, statement flinkgatewayv1beta1.SqlV1beta1Statement) error {
	httpResp, err := c.StatementsSqlV1beta1Api.UpdateSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).SqlV1beta1Statement(statement).Execute()
	return flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	httpResp, err := c.StatementsSqlV1beta1Api.DeleteSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId, statementName).Execute()
	return flink.CatchError(err, httpResp)
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

func (c *FlinkGatewayClient) CreateStatement(statement flinkgatewayv1beta1.SqlV1beta1Statement, principal, environmentId, orgId string) (flinkgatewayv1beta1.SqlV1beta1Statement, error) {
	if principal != "" {
		statement.Spec.Principal = flinkgatewayv1beta1.PtrString(principal)
	}
	resp, httpResp, err := c.StatementsSqlV1beta1Api.CreateSqlv1beta1Statement(c.flinkGatewayApiContext(), orgId, environmentId).SqlV1beta1Statement(statement).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) UpdateStatement(environmentId, statementName, organizationId string, statement flinkgatewayv1beta1.SqlV1beta1Statement) error {
	httpResp, err := c.StatementsSqlV1beta1Api.UpdateSqlv1beta1Statement(c.flinkGatewayApiContext(), organizationId, environmentId, statementName).SqlV1beta1Statement(statement).Execute()
	return flink.CatchError(err, httpResp)
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
