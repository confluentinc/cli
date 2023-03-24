package controller

import (
	"context"
	"net/http"
	"strings"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type GatewayClient struct {
	authToken     string
	envId         string
	computePoolId string
	client        *v1.APIClient
}

func (c *GatewayClient) CreateStatement(ctx context.Context, statement string, properties map[string]string) (v1.SqlV1alpha1Statement, *http.Response, error) {
	statementName, ok := properties["pipeline.name"]
	if !ok || strings.TrimSpace(statementName) == "" {
		statementName = uuid.New().String()
	}

	propsSlice := lo.MapToSlice(properties, func(key, val string) v1.SqlV1alpha1Property { return v1.SqlV1alpha1Property{Key: &key, Value: &val} })

	statementObj := v1.SqlV1alpha1Statement{
		Spec: &v1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
			Statement:     &statement,
			ComputePoolId: &c.computePoolId,
			Properties:    &propsSlice,
		},
	}

	ctx = context.WithValue(ctx, v1.ContextAccessToken, c.authToken)
	createdStatement, resp, err := c.client.StatementsSqlV1alpha1Api.CreateSqlV1alpha1Statement(ctx, c.envId).SqlV1alpha1Statement(statementObj).Execute()
	return createdStatement, resp, err
}

// TODO result handling: https://confluentinc.atlassian.net/wiki/spaces/FLINK/pages/3004703887/WIP+Flink+Gateway+-+Results+handling
func (s *Store) FetchStatementResults(envId, statementId string) (*StatementResult, error) {
	return &StatementResult{
		Status:  "Completed",
		Columns: []string{},
		Rows:    [][]string{{}},
	}, nil
}

func NewGatewayClient(envId, computePoolId, authToken string) *GatewayClient {
	return &GatewayClient{
		envId:         envId,
		computePoolId: computePoolId,
		authToken:     authToken,
		client:        v1.NewAPIClient(v1.NewConfiguration()),
	}
}
