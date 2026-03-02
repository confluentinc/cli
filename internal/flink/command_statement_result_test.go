package flink

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func newTestCommand(outputFormat string) (*cobra.Command, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.Flags().StringP(output.FlagName, "o", "", "output format")
	if outputFormat != "" {
		_ = cmd.Flags().Set(output.FlagName, outputFormat)
	}
	return cmd, buf
}

func TestPrintStatementResults_HumanNoRows(t *testing.T) {
	cmd, buf := newTestCommand("")
	data := &statementResultData{
		Headers: []string{"id", "name"},
	}

	err := printStatementResults(cmd, data)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No results found.")
}

func TestPrintStatementResults_HumanNilData(t *testing.T) {
	cmd, buf := newTestCommand("")

	err := printStatementResults(cmd, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No results found.")
}

func TestPrintStatementResults_HumanWithRows(t *testing.T) {
	cmd, buf := newTestCommand("")
	data := &statementResultData{
		Headers: []string{"id", "name"},
		Rows: [][]string{
			{"1", "alice"},
			{"2", "bob"},
		},
	}

	err := printStatementResults(cmd, data)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "id")
	assert.Contains(t, out, "name")
	assert.Contains(t, out, "alice")
	assert.Contains(t, out, "bob")
}

func TestPrintStatementResults_SerializedNoRows(t *testing.T) {
	cmd, _ := newTestCommand("json")
	data := &statementResultData{
		Headers: []string{"id", "name"},
	}

	// SerializedOutput writes to os.Stdout directly, so we just verify no error
	err := printStatementResults(cmd, data)
	require.NoError(t, err)
}

func TestPrintStatementResults_SerializedNilData(t *testing.T) {
	cmd, _ := newTestCommand("json")

	err := printStatementResults(cmd, nil)
	require.NoError(t, err)
}

func TestPrintStatementResults_SerializedWithRows(t *testing.T) {
	cmd, _ := newTestCommand("json")
	data := &statementResultData{
		Headers: []string{"id", "name"},
		Rows: [][]string{
			{"1", "alice"},
			{"2", "bob"},
		},
	}

	err := printStatementResults(cmd, data)
	require.NoError(t, err)
}

func makeSchema(colNames ...string) flinkgatewayv1.SqlV1ResultSchema {
	cols := make([]flinkgatewayv1.ColumnDetails, len(colNames))
	for i, name := range colNames {
		cols[i] = flinkgatewayv1.ColumnDetails{Name: name}
	}
	return flinkgatewayv1.SqlV1ResultSchema{Columns: &cols}
}

func makeResultResponse(rows [][]any, nextPageToken string) flinkgatewayv1.SqlV1StatementResult {
	data := make([]any, len(rows))
	for i, row := range rows {
		data[i] = map[string]any{"row": row}
	}

	results := flinkgatewayv1.SqlV1StatementResultResults{Data: &data}
	meta := flinkgatewayv1.ResultListMeta{}
	if nextPageToken != "" {
		next := fmt.Sprintf("https://api.example.com/results?page_token=%s", nextPageToken)
		meta.Next = &next
	}

	return flinkgatewayv1.SqlV1StatementResult{
		Metadata: meta,
		Results:  &results,
	}
}

func TestFetchAllResults_SinglePage(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id", "name")
	resp := makeResultResponse([][]any{
		{"1", "alice"},
		{"2", "bob"},
	}, "")

	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(resp, nil)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, result.Headers)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, []string{"1", "alice"}, result.Rows[0])
	assert.Equal(t, []string{"2", "bob"}, result.Rows[1])
}

func TestFetchAllResults_MultiplePages(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	page1 := makeResultResponse([][]any{{"1"}, {"2"}}, "token2")
	page2 := makeResultResponse([][]any{{"3"}}, "")

	gomock.InOrder(
		client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(page1, nil),
		client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "token2").Return(page2, nil),
	)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 100)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 3)
	assert.Equal(t, []string{"1"}, result.Rows[0])
	assert.Equal(t, []string{"2"}, result.Rows[1])
	assert.Equal(t, []string{"3"}, result.Rows[2])
}

func TestFetchAllResults_MaxRowsTruncation(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	resp := makeResultResponse([][]any{{"1"}, {"2"}, {"3"}, {"4"}, {"5"}}, "token2")

	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(resp, nil)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 3)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 3)
	assert.Equal(t, []string{"1"}, result.Rows[0])
	assert.Equal(t, []string{"3"}, result.Rows[2])
}

func TestFetchAllResults_MaxRowsZeroFetchesAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	resp := makeResultResponse([][]any{{"1"}, {"2"}, {"3"}}, "")

	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(resp, nil)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 0)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 3)
}

func TestFetchAllResults_NullFieldRendersAsNULL(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id", "name")
	resp := makeResultResponse([][]any{{"1", nil}}, "")

	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(resp, nil)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 100)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "NULL"}, result.Rows[0])
}

func TestFetchAllResults_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(flinkgatewayv1.SqlV1StatementResult{}, fmt.Errorf("connection refused"))

	_, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestFieldToString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, "NULL"},
		{"hello", "hello"},
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			assert.Equal(t, tt.expected, fieldToString(tt.input))
		})
	}
}

func TestExtractResultPageToken(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		hasError bool
	}{
		{"empty URL", "", "", false},
		{"URL with token", "https://api.example.com/results?page_token=abc123", "abc123", false},
		{"URL without token", "https://api.example.com/results", "", false},
		{"URL with multiple params", "https://api.example.com/results?foo=bar&page_token=xyz&baz=qux", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := extractResultPageToken(tt.url)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, token)
			}
		})
	}
}

func TestFetchAllResults_MaxRowsAcrossPages(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	page1 := makeResultResponse([][]any{{"1"}, {"2"}}, "token2")
	page2 := makeResultResponse([][]any{{"3"}, {"4"}, {"5"}}, "token3")

	gomock.InOrder(
		client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(page1, nil),
		client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "token2").Return(page2, nil),
	)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 4)
	require.NoError(t, err)
	assert.Len(t, result.Rows, 4)
}

func TestFetchAllResults_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock.NewMockGatewayClientInterface(ctrl)

	schema := makeSchema("id")
	resp := makeResultResponse([][]any{}, "")

	client.EXPECT().GetStatementResults("env1", "stmt1", "org1", "").Return(resp, nil)

	result, err := fetchAllResults(client, "env1", "stmt1", "org1", schema, 100)
	require.NoError(t, err)
	assert.Empty(t, result.Rows)
	assert.Equal(t, []string{"id"}, result.Headers)
}

func TestPrintStatementResults_HumanTableFormat(t *testing.T) {
	cmd, buf := newTestCommand("")
	data := &statementResultData{
		Headers: []string{"count"},
		Rows:    [][]string{{"42"}},
	}

	err := printStatementResults(cmd, data)
	require.NoError(t, err)
	out := buf.String()
	// Table output should contain the header and value
	assert.True(t, strings.Contains(out, "count") && strings.Contains(out, "42"), "expected table with 'count' header and '42' value, got: %s", out)
}
