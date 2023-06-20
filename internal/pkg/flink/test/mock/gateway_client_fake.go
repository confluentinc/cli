package mock

import (
	"fmt"
	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/google/uuid"
	"pgregory.net/rapid"
	"time"
)

const (
	staticQuery  = "static;"
	dynamicQuery = "dynamic;"
)

type FakeFlinkGatewayClient struct {
	statement  flinkgatewayv1alpha1.SqlV1alpha1Statement
	statements []flinkgatewayv1alpha1.SqlV1alpha1Statement
	fakeCount  int
}

func NewFakeFlinkGatewayClient() ccloudv2.GatewayClientInterface {
	return &FakeFlinkGatewayClient{}
}

func (c *FakeFlinkGatewayClient) DeleteStatement(environmentId, statementName, orgId string) error {
	return nil
}

func (c *FakeFlinkGatewayClient) GetStatement(environmentId, statementName, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	secondsToWait := time.Duration(rapid.IntRange(1, 3).Example())
	time.Sleep(secondsToWait * time.Second)
	c.statement.Status.Phase = "RUNNING"
	return c.statement, nil
}

func (c *FakeFlinkGatewayClient) ListStatements(environmentId, orgId string) ([]flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	return c.statements, nil
}

func (c *FakeFlinkGatewayClient) CreateStatement(statement, computePoolId, identityPoolId string, properties map[string]string, environmentId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1Statement, error) {
	columnDetails := c.resultsSchema(statement)
	statementName := uuid.New().String()[:20]

	creationTime := time.Now()
	c.statement = flinkgatewayv1alpha1.SqlV1alpha1Statement{
		ApiVersion: nil,
		Kind:       nil,
		Metadata: &flinkgatewayv1alpha1.ObjectMeta{
			Self:      "",
			CreatedAt: &creationTime,
			UpdatedAt: nil,
		},
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName:  &statementName,
			Statement:      &statement,
			Properties:     &properties,
			ComputePoolId:  &computePoolId,
			IdentityPoolId: &identityPoolId,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:        "PENDING",
			ResultSchema: &flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{Columns: &columnDetails},
			Detail:       nil,
		},
	}
	c.statements = append(c.statements, c.statement)

	return c.statement, nil
}

func (c *FakeFlinkGatewayClient) GetStatementResults(environmentId, statementId, orgId, pageToken string) (flinkgatewayv1alpha1.SqlV1alpha1StatementResult, error) {
	resultData, nextUrl := c.results()
	return flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
		ApiVersion: "",
		Kind:       "",
		Metadata: flinkgatewayv1alpha1.ResultListMeta{
			Self:      nil,
			Next:      &nextUrl,
			CreatedAt: nil,
		},
		Results: &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{Data: &resultData},
	}, nil
}

func (c *FakeFlinkGatewayClient) GetExceptions(environmentId, statementId, orgId string) (flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList, error) {
	return flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList{}, nil
}

func (c *FakeFlinkGatewayClient) resultsSchema(statement string) []flinkgatewayv1alpha1.ColumnDetails {
	switch statement {
	case staticQuery:
		return c.staticTableSchema()
	case dynamicQuery:
		return c.dynamicTableSchema()
	}
	return nil
}

func (c *FakeFlinkGatewayClient) staticTableSchema() []flinkgatewayv1alpha1.ColumnDetails {
	return generators.MockResultColumns(5, 2).Example()
}

func (c *FakeFlinkGatewayClient) dynamicTableSchema() []flinkgatewayv1alpha1.ColumnDetails {
	return []flinkgatewayv1alpha1.ColumnDetails{
		{
			Name: "Count",
			Type: flinkgatewayv1alpha1.DataType{
				Nullable: false,
				Type:     "INTEGER",
			},
		},
	}
}

func (c *FakeFlinkGatewayClient) results() ([]any, string) {
	switch c.statement.Spec.GetStatement() {
	case staticQuery:
		return c.staticTableResults()
	case dynamicQuery:
		return c.dynamicTableResults()
	}
	return nil, ""
}

func (c *FakeFlinkGatewayClient) staticTableResults() ([]any, string) {
	return rapid.SliceOfN(generators.MockResultRow(c.statement.Status.ResultSchema.GetColumns()), 20, 50).Example(), ""
}

func (c *FakeFlinkGatewayClient) dynamicTableResults() ([]any, string) {
	elapsedSeconds := int(time.Since(c.statement.Metadata.GetCreatedAt()).Seconds())
	// sync count and elapsed seconds once at the start
	if c.fakeCount == 0 {
		c.fakeCount = elapsedSeconds
	} else if c.fakeCount == elapsedSeconds {
		// if count and elapsed seconds are in sync, that means we are live and there should be no new results
		return nil, fmt.Sprintf("https://devel.cpdev.cloud/some/results?page_token=%s", "not-empty")
	}

	var results []any
	for i := 0; i < c.fakeCount; i++ {
		// update before
		results = append(results, map[string]any{
			"op":  int32(1),
			"row": []any{fmt.Sprintf("%v", i)},
		})
	}
	// update after
	results = append(results, map[string]any{
		"op":  int32(2),
		"row": []any{fmt.Sprintf("%v", c.fakeCount)},
	})
	c.fakeCount++

	return results, fmt.Sprintf("https://devel.cpdev.cloud/some/results?page_token=%s", "not-empty")
}
