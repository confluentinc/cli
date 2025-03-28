package mock

import (
	"fmt"
	"time"

	"pgregory.net/rapid"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
)

const (
	// Use `static;` to receive an example of results for a COMPLETED statement.
	// It will return a randomized set of data types and a different number of rows and columns every time you use it.
	staticQuery = "static;"
	// Use `dynamic;` to receive an example of results for a RUNNING statement.
	// It will return an integer counter that is incremented every second.
	dynamicQuery = "dynamic;"
)

type FakeFlinkGatewayClient struct {
	statement   flinkgatewayv1.SqlV1Statement
	statements  []flinkgatewayv1.SqlV1Statement
	connection  flinkgatewayv1.SqlV1Connection
	connections []flinkgatewayv1.SqlV1Connection
	fakeCount   int
}

func (c *FakeFlinkGatewayClient) GetAuthToken() string {
	return ""
}

func NewFakeFlinkGatewayClient() ccloudv2.GatewayClientInterface {
	return &FakeFlinkGatewayClient{}
}

func (c *FakeFlinkGatewayClient) CreateConnection(connection flinkgatewayv1.SqlV1Connection, _, _ string) (flinkgatewayv1.SqlV1Connection, error) {
	c.fakeCount = 0
	c.connection = connection
	c.connections = append(c.connections, c.connection)
	return c.connection, nil
}

func (c *FakeFlinkGatewayClient) ListConnections(_, _, _ string) ([]flinkgatewayv1.SqlV1Connection, error) {
	return c.connections, nil
}

func (c *FakeFlinkGatewayClient) GetConnection(_, _, _ string) (flinkgatewayv1.SqlV1Connection, error) {
	secondsToWait := time.Duration(rapid.IntRange(1, 3).Example())
	time.Sleep(secondsToWait * time.Second)
	c.connection.Status.Phase = "Completed"
	return c.connection, nil
}

func (c *FakeFlinkGatewayClient) DeleteConnection(_, _, _ string) error {
	return nil
}

func (c *FakeFlinkGatewayClient) UpdateConnection(_, _, _ string, _ flinkgatewayv1.SqlV1Connection) error {
	return nil
}

func (c *FakeFlinkGatewayClient) DeleteStatement(_, _, _ string) error {
	return nil
}

func (c *FakeFlinkGatewayClient) UpdateStatement(_, _, _ string, _ flinkgatewayv1.SqlV1Statement) error {
	return nil
}

func (c *FakeFlinkGatewayClient) GetStatement(_, _, _ string) (flinkgatewayv1.SqlV1Statement, error) {
	secondsToWait := time.Duration(rapid.IntRange(1, 3).Example())
	time.Sleep(secondsToWait * time.Second)
	c.statement.Status.Phase = "RUNNING"
	return c.statement, nil
}

func (c *FakeFlinkGatewayClient) ListStatements(_, _, _ string) ([]flinkgatewayv1.SqlV1Statement, error) {
	return c.statements, nil
}

func (c *FakeFlinkGatewayClient) CreateStatement(statement flinkgatewayv1.SqlV1Statement, _, _, _ string) (flinkgatewayv1.SqlV1Statement, error) {
	c.fakeCount = 0
	c.statement = statement
	c.statements = append(c.statements, c.statement)

	return c.statement, nil
}

func (c *FakeFlinkGatewayClient) getFakeResultSchema(statement string) []flinkgatewayv1.ColumnDetails {
	switch statement {
	case staticQuery:
		return c.getStaticFakeResultSchema()
	case dynamicQuery:
		return c.getDynamicFakeResultSchema()
	}
	return nil
}

func (c *FakeFlinkGatewayClient) getStaticFakeResultSchema() []flinkgatewayv1.ColumnDetails {
	return generators.MockResultColumns(5, 2).Example()
}

func (c *FakeFlinkGatewayClient) getDynamicFakeResultSchema() []flinkgatewayv1.ColumnDetails {
	return []flinkgatewayv1.ColumnDetails{
		{
			Name: "Count",
			Type: flinkgatewayv1.DataType{
				Nullable: false,
				Type:     "INTEGER",
			},
		},
	}
}

func (c *FakeFlinkGatewayClient) GetStatementResults(_, _, _, _ string) (flinkgatewayv1.SqlV1StatementResult, error) {
	resultData, nextUrl := c.getFakeResults()
	result := flinkgatewayv1.SqlV1StatementResult{
		Metadata: flinkgatewayv1.ResultListMeta{Next: &nextUrl},
		Results:  &flinkgatewayv1.SqlV1StatementResultResults{Data: &resultData},
	}
	return result, nil
}

func (c *FakeFlinkGatewayClient) getFakeResults() ([]any, string) {
	switch c.statement.Spec.GetStatement() {
	case staticQuery:
		return c.getFakeResultsCompletedTable()
	case dynamicQuery:
		return c.getFakeResultsRunningCounter()
	}
	return nil, ""
}

func (c *FakeFlinkGatewayClient) getFakeResultsCompletedTable() ([]any, string) {
	return rapid.SliceOfN(generators.MockResultRow(c.statement.Status.Traits.Schema.GetColumns()), 20, 50).Example(), ""
}

func (c *FakeFlinkGatewayClient) getFakeResultsRunningCounter() ([]any, string) {
	elapsedSeconds := int(time.Since(c.statement.Metadata.GetCreatedAt()).Seconds())
	if c.fakeCount >= elapsedSeconds {
		// we are live and there should be no more results
		return nil, fmt.Sprintf("https://devel.cpdev.cloud/some/results?page_token=%s", "not-empty")
	}

	var results []any
	// remove all previous entries
	for i := 0; i < c.fakeCount; i++ {
		// update before
		results = append(results, map[string]any{
			"op":  float64(1),
			"row": []any{fmt.Sprintf("%v", i)},
		})
	}

	// update after (add latest entry)
	results = append(results, map[string]any{
		"op":  float64(2),
		"row": []any{fmt.Sprintf("%v", c.fakeCount)},
	})
	c.fakeCount++

	return results, fmt.Sprintf("https://devel.cpdev.cloud/some/results?page_token=%s", "not-empty")
}

func (c *FakeFlinkGatewayClient) GetExceptions(_, _, _ string) ([]flinkgatewayv1.SqlV1StatementException, error) {
	return []flinkgatewayv1.SqlV1StatementException{}, nil
}
