package ccloudv2

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/v3/pkg/errors/flink"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type GatewayClientInterface interface {
	DeleteStatement(environmentId, statementName, orgId string) error
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementList, error)
	CreateStatement(statement, computePoolId string, properties map[string]string, serviceAccountId, identityPoolId, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error)
	GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error)
}

type FlinkGatewayClient struct {
	*flinkgatewayv1alpha1.APIClient
	AuthToken string
}

func NewFlinkGatewayClient(url, userAgent string, unsafeTrace bool, authToken string) *FlinkGatewayClient {
	cfg := flinkgatewayv1alpha1.NewConfiguration()
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
	cfg.Servers = flinkgatewayv1alpha1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &FlinkGatewayClient{
		APIClient: flinkgatewayv1alpha1.NewAPIClient(cfg),
		AuthToken: authToken,
	}
}

func (c *FlinkGatewayClient) flinkGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), flinkgatewayv1alpha1.ContextAccessToken, c.AuthToken)
}

func (c *FlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	httpResp, err := c.StatementsSqlV1alpha1Api.DeleteSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, httpResp, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementList, error) {
	req := c.StatementsSqlV1alpha1Api.ListSqlV1alpha1Statements(c.flinkGatewayApiContext(), environmentId).OrgId(orgId).PageSize(ccloudV2ListPageSize)

	if computePoolId != "" {
		req = req.ComputePoolId(computePoolId)
	}

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) ListAllStatements(environmentId, orgId, computePoolId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	var allStatements []flinkgatewayv1alpha1.SqlV1alpha1Statement
	pageToken := ""
	done := false
	for !done {
		statementListResponse, err := c.ListStatements(environmentId, orgId, pageToken, computePoolId)
		if err != nil {
			return nil, err
		}
		allStatements = append(allStatements, statementListResponse.GetData()...)
		nextUrl := statementListResponse.Metadata.GetNext()
		pageToken, done, err = extractNextPageToken(flinkgatewayv1alpha1.NewNullableString(&nextUrl))
		if err != nil {
			return nil, err
		}
	}
	return allStatements, nil
}

func (c *FlinkGatewayClient) CreateStatement(statement, computePoolId string, properties map[string]string, serviceAccountId, identityPoolId, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	statementName := uuid.New().String()[:18]

	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
			Statement:     &statement,
			ComputePoolId: &computePoolId,
			Properties:    &properties,
		},
	}
	if serviceAccountId != "" {
		// add the service account header and remove it after the request
		c.GetConfig().AddDefaultHeader("Service-Account-Id", serviceAccountId)
		defer delete(c.GetConfig().DefaultHeader, "Service-Account-Id")
	} else {
		// if this is also empty we have the interactive query/AUMP case
		statementObj.Spec.IdentityPoolId = &identityPoolId
	}
	resp, httpResp, err := c.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId).SqlV1alpha1Statement(statementObj).OrgId(orgId).Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	req := c.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, flink.CatchError(err, httpResp)
}

func (c *FlinkGatewayClient) GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error) {
	resp, httpResp, err := c.StatementExceptionsSqlV1alpha1Api.GetSqlV1alpha1StatementExceptions(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId).Execute()
	return resp, flink.CatchError(err, httpResp)
}
