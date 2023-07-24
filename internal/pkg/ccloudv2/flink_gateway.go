package ccloudv2

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type GatewayClientInterface interface {
	DeleteStatement(environmentId, statementName, orgId string) error
	GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
	ListStatements(environmentId, orgId, pageToken, computePoolId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementList, error)
	CreateStatement(statement, computePoolId, identityPoolId string, properties map[string]string, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error)
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
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
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
	return makeFlinkError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	resp, httpResp, err := c.StatementsSqlV1alpha1Api.GetSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId, statementName).OrgId(orgId).Execute()
	return resp, makeFlinkError(err, httpResp)
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
	return resp, makeFlinkError(err, httpResp)
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

func (c *FlinkGatewayClient) CreateStatement(statement, computePoolId, identityPoolId string, properties map[string]string, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	statementName := uuid.New().String()[:18]

	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName:  &statementName,
			Statement:      &statement,
			ComputePoolId:  &computePoolId,
			IdentityPoolId: &identityPoolId,
			Properties:     &properties,
		},
	}
	resp, httpResp, err := c.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(c.flinkGatewayApiContext(), environmentId).SqlV1alpha1Statement(statementObj).OrgId(orgId).Execute()
	return resp, makeFlinkError(err, httpResp)
}

func (c *FlinkGatewayClient) GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	req := c.StatementResultSqlV1alpha1Api.GetSqlV1alpha1StatementResult(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResp, err := req.Execute()
	return resp, makeFlinkError(err, httpResp)
}

func (c *FlinkGatewayClient) GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error) {
	resp, httpResp, err := c.StatementExceptionsSqlV1alpha1Api.GetSqlV1alpha1StatementExceptions(c.flinkGatewayApiContext(), environmentId, statementId).OrgId(orgId).Execute()
	return resp, makeFlinkError(err, httpResp)
}

// FlinkError extends the ErrorWithSuggestion with a status code.
type FlinkError struct {
	errorMsg       string
	suggestionsMsg string
	statusCode     int
}

func NewFlinkError(errorMsg string, suggestionsMsg string, statusCode int) FlinkError {
	return FlinkError{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
		statusCode:     statusCode,
	}
}

func (f FlinkError) StatusCode() int {
	return f.statusCode
}

func (f FlinkError) GetSuggestionsMsg() string {
	return f.suggestionsMsg
}

func (f FlinkError) Error() string {
	return f.errorMsg
}

type Coder interface {
	StatusCode() int
}

var _ Coder = (*FlinkError)(nil)
var _ errors.ErrorWithSuggestions = (*FlinkError)(nil)

// Extends error with status code, including suggestion if err type is ErrorWithSuggestion
func makeFlinkError(err error, r *http.Response) error {
	if err == nil {
		return nil
	}
	err = errors.CatchCCloudV2Error(err, r)
	suggestion := ""
	if suggester, ok := err.(errors.ErrorWithSuggestions); ok {
		suggestion = suggester.GetSuggestionsMsg()
	}
	var statusCode int
	if r != nil {
		statusCode = r.StatusCode
	}
	return FlinkError{
		statusCode:     statusCode,
		errorMsg:       err.Error(),
		suggestionsMsg: suggestion,
	}
}
