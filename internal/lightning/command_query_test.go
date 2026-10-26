package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	rtcesqlv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func TestResolveDisplayCap(t *testing.T) {
	tests := []struct {
		name           string
		format         output.Format
		maxRowsChanged bool
		maxRows        int
		wantCap        int
		wantHumanCap   bool
	}{
		{"human default caps at 100", output.Human, false, 0, defaultHumanRows, true},
		{"human with explicit --max-rows uses it", output.Human, true, 50, 50, false},
		{"human with explicit --max-rows 0 is uncapped", output.Human, true, 0, 0, false},
		{"json stays uncapped by default", output.JSON, false, 0, 0, false},
		{"yaml stays uncapped by default", output.YAML, false, 0, 0, false},
		{"yaml with explicit --max-rows uses it", output.YAML, true, 200, 200, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotCap, gotHumanCap := resolveDisplayCap(test.format, test.maxRowsChanged, test.maxRows)
			require.Equal(t, test.wantCap, gotCap)
			require.Equal(t, test.wantHumanCap, gotHumanCap)
		})
	}
}

func TestResolveMaxRows(t *testing.T) {
	newCmd := func(maxRows string) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().Int("max-rows", 0, "")
		require.NoError(t, cmd.Flags().Set("max-rows", maxRows))
		return cmd
	}

	maxRows, err := resolveMaxRows(newCmd("25"))
	require.NoError(t, err)
	require.Equal(t, 25, maxRows)

	maxRows, err = resolveMaxRows(newCmd("0"))
	require.NoError(t, err)
	require.Equal(t, 0, maxRows)

	_, err = resolveMaxRows(newCmd("-1"))
	require.EqualError(t, err, "the `--max-rows` flag must not be negative")

	_, err = resolveMaxRows(&cobra.Command{})
	require.Error(t, err)
}

func TestMaxResultRows(t *testing.T) {
	tests := []struct {
		maxRows int
		want    int32
	}{
		{0, 0},
		{1, 2},
		{defaultHumanRows, defaultHumanRows + 1},
		{1000, 1001},
		{math.MaxInt32, math.MaxInt32},
	}
	for _, test := range tests {
		require.Equal(t, test.want, maxResultRows(test.maxRows), "maxRows %d", test.maxRows)
	}
}

func TestResolveSQL(t *testing.T) {
	newCmd := func(sql, file string) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().String("sql", "", "")
		cmd.Flags().String("file", "", "")
		require.NoError(t, cmd.Flags().Set("sql", sql))
		require.NoError(t, cmd.Flags().Set("file", file))
		return cmd
	}

	dir := t.TempDir()
	queryFile := filepath.Join(dir, "query.sql")
	require.NoError(t, os.WriteFile(queryFile, []byte("SELECT * FROM orders\n"), 0600))
	blankFile := filepath.Join(dir, "blank.sql")
	require.NoError(t, os.WriteFile(blankFile, []byte(" \n\t"), 0600))

	sql, err := resolveSQL(newCmd("SELECT 1", ""))
	require.NoError(t, err)
	require.Equal(t, "SELECT 1", sql)

	sql, err = resolveSQL(newCmd("", queryFile))
	require.NoError(t, err)
	require.Equal(t, "SELECT * FROM orders\n", sql)

	missingFile := filepath.Join(dir, "missing.sql")
	_, readErr := os.ReadFile(missingFile)
	_, err = resolveSQL(newCmd("", missingFile))
	require.EqualError(t, err, fmt.Sprintf(`failed to read the SQL statement from "%s": %v`, missingFile, readErr))
	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, "Check that the `--file` path is correct and the file is readable.", errorWithSuggestions.GetSuggestionsMsg())

	_, err = resolveSQL(newCmd("", blankFile))
	require.EqualError(t, err, "the SQL statement is empty")

	_, err = resolveSQL(newCmd("  ", ""))
	require.EqualError(t, err, "the SQL statement is empty")

	_, err = resolveSQL(&cobra.Command{})
	require.Error(t, err)

	cmd := &cobra.Command{}
	cmd.Flags().String("sql", "", "")
	_, err = resolveSQL(cmd)
	require.Error(t, err, "a missing file flag is an error")
}

func parseResult(t *testing.T, raw string) rtcesqlv1.QueryV1QueryResult {
	var result rtcesqlv1.QueryV1QueryResult
	require.NoError(t, json.Unmarshal([]byte(raw), &result))
	return result
}

func TestConvertResultScalarTypes(t *testing.T) {
	result := parseResult(t, `{
		"result_format": "JSON",
		"schema": {"columns": [
			{"name": "c_boolean", "type": {"type": "BOOLEAN", "nullable": true}},
			{"name": "c_integer", "type": {"type": "INTEGER", "nullable": true}},
			{"name": "c_bigint", "type": {"type": "BIGINT", "nullable": true}},
			{"name": "c_double", "type": {"type": "DOUBLE", "nullable": true}},
			{"name": "c_decimal", "type": {"type": "DECIMAL", "nullable": true, "precision": 10, "scale": 2}},
			{"name": "c_varchar", "type": {"type": "VARCHAR", "nullable": true}},
			{"name": "c_timestamp", "type": {"type": "TIMESTAMP_WITHOUT_TIME_ZONE", "nullable": true}},
			{"name": "c_null", "type": {"type": "INTEGER", "nullable": true}}
		]},
		"data": [["TRUE", "70000", "9007199254740993", "3.14159", "149.00", "hello", "2026-10-12 13:45:30.123456", null]]
	}`)

	headers, rows, err := convertResult(result)
	require.NoError(t, err)
	require.Equal(t, []string{"c_boolean", "c_integer", "c_bigint", "c_double", "c_decimal", "c_varchar", "c_timestamp", "c_null"}, headers)
	require.Len(t, rows, 1)

	rendered := make([]string, 0, len(rows[0]))
	serialized := make([]any, 0, len(rows[0]))
	for _, field := range rows[0] {
		rendered = append(rendered, field.ToString())
		serialized = append(serialized, field.ToSerializedValue())
	}
	require.Equal(t, []string{"TRUE", "70000", "9007199254740993", "3.14159", "149.00", "hello", "2026-10-12 13:45:30.123456", "NULL"}, rendered)
	// As in `confluent flink query`, BIGINT and DECIMAL stay strings to keep precision.
	require.Equal(t, []any{true, int64(70000), "9007199254740993", 3.14159, "149.00", "hello", "2026-10-12 13:45:30.123456", nil}, serialized)
}

func TestConvertResultComplexTypes(t *testing.T) {
	result := parseResult(t, `{
		"result_format": "JSON",
		"schema": {"columns": [
			{"name": "c_array", "type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "VARCHAR", "nullable": true}}},
			{"name": "c_map", "type": {"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}},
			{"name": "c_int_map", "type": {"type": "MAP", "nullable": true, "key_type": {"type": "INTEGER", "nullable": false}, "value_type": {"type": "VARCHAR", "nullable": true}}},
			{"name": "c_multiset", "type": {"type": "MULTISET", "nullable": true, "element_type": {"type": "VARCHAR", "nullable": true}}},
			{"name": "c_row", "type": {"type": "ROW", "nullable": true, "fields": [
				{"name": "USER_ID", "field_type": {"type": "VARCHAR", "nullable": false}},
				{"name": "AGE", "field_type": {"type": "INTEGER", "nullable": false}}
			]}},
			{"name": "c_nested", "type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "INTEGER", "nullable": true}}}},
			{"name": "c_empty_array", "type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "INTEGER", "nullable": true}}},
			{"name": "c_null_array", "type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "INTEGER", "nullable": true}}}
		]},
		"data": [[
			["a", null, "c"],
			[["latency", "42"], ["errors", "0"]],
			[["1", "one"], ["2", "two"]],
			[["x", "2"]],
			["u-1", "42"],
			[["1", "2"], []],
			[],
			null
		]]
	}`)

	_, rows, err := convertResult(result)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	rendered := make([]string, 0, len(rows[0]))
	serialized := make([]any, 0, len(rows[0]))
	for _, field := range rows[0] {
		rendered = append(rendered, field.ToString())
		serialized = append(serialized, field.ToSerializedValue())
	}
	require.Equal(t, []string{"[a, NULL, c]", "{latency=42, errors=0}", "{1=one, 2=two}", "{x=2}", "(u-1, 42)", "[[1, 2], []]", "[]", "NULL"}, rendered)
	require.Equal(t, []any{
		[]any{"a", nil, "c"},
		map[string]any{"latency": int64(42), "errors": int64(0)},
		[]any{map[string]any{"key": int64(1), "value": "one"}, map[string]any{"key": int64(2), "value": "two"}},
		map[string]any{"x": int64(2)},
		[]any{"u-1", int64(42)},
		[]any{[]any{int64(1), int64(2)}, []any{}},
		[]any{},
		nil,
	}, serialized)

	row, ok := rows[0][4].(types.RowStatementResultField)
	require.True(t, ok)
	require.Equal(t, []types.StatementResultFieldType{types.Varchar, types.Integer}, row.ElementTypes)
}

func TestConvertResultWithoutRows(t *testing.T) {
	headers, rows, err := convertResult(parseResult(t, `{"result_format": "JSON", "schema": {"columns": [{"name": "id", "type": {"type": "INTEGER", "nullable": true}}]}, "data": []}`))
	require.NoError(t, err)
	require.Equal(t, []string{"id"}, headers)
	require.Empty(t, rows)

	headers, rows, err = convertResult(parseResult(t, `{"result_format": "JSON", "schema": {"columns": []}}`))
	require.NoError(t, err)
	require.Empty(t, headers)
	require.Empty(t, rows)
}

func TestConvertResultErrors(t *testing.T) {
	// oneValue is a result with one column of dataType and one row holding value.
	oneValue := func(dataType, value string) string {
		return fmt.Sprintf(`{"result_format": "JSON", "schema": {"columns": [{"name": "c", "type": %s}]}, "data": [[%s]]}`, dataType, value)
	}
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{
			name:    "arrow result",
			raw:     `{"result_format": "ARROW_STREAM", "schema": {"columns": []}, "data": "QVJST1c="}`,
			wantErr: `unsupported result format "ARROW_STREAM"`,
		},
		{
			name:    "no result",
			raw:     `{}`,
			wantErr: `unsupported result format ""`,
		},
		{
			name:    "rows that aren't an array",
			raw:     `{"result_format": "JSON", "schema": {"columns": []}, "data": "W10="}`,
			wantErr: "the query result's rows are not a JSON array",
		},
		{
			name:    "row shorter than the schema",
			raw:     `{"result_format": "JSON", "schema": {"columns": [{"name": "a", "type": {"type": "INTEGER", "nullable": true}}, {"name": "b", "type": {"type": "INTEGER", "nullable": true}}]}, "data": [["1"]]}`,
			wantErr: "result row 1 has 1 values, expected 2",
		},
		{
			name:    "unknown column type",
			raw:     oneValue(`{"type": "NOT_A_REAL_TYPE", "nullable": true}`, `"1"`),
			wantErr: `failed to read column "c": unsupported result field type "NOT_A_REAL_TYPE": this CLI may be out of date`,
		},
		{
			name:    "scalar sent as an array",
			raw:     oneValue(`{"type": "INTEGER", "nullable": true}`, `["1"]`),
			wantErr: `failed to read column "c": INTEGER value is not a string`,
		},
		{
			name:    "array sent as a string",
			raw:     oneValue(`{"type": "ARRAY", "nullable": true, "element_type": {"type": "INTEGER", "nullable": true}}`, `"1"`),
			wantErr: `failed to read column "c": ARRAY value is not an array`,
		},
		{
			name:    "NULL of an unknown type",
			raw:     oneValue(`{"type": "NOT_A_REAL_TYPE", "nullable": true}`, `null`),
			wantErr: `failed to read column "c": unsupported result field type "NOT_A_REAL_TYPE"`,
		},
		{
			name:    "unknown array element type",
			raw:     oneValue(`{"type": "ARRAY", "nullable": true, "element_type": {"type": "NOT_A_REAL_TYPE", "nullable": true}}`, `[]`),
			wantErr: `unsupported result field type "NOT_A_REAL_TYPE"`,
		},
		{
			name:    "bad array element",
			raw:     oneValue(`{"type": "ARRAY", "nullable": true, "element_type": {"type": "INTEGER", "nullable": true}}`, `[["1"]]`),
			wantErr: "INTEGER value is not a string",
		},
		{
			name:    "map entry that isn't a pair",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `[["k"]]`),
			wantErr: "MAP entry is not a [key, value] pair",
		},
		{
			name:    "map entry with more than a key and a value",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `[["k", "1", "2"]]`),
			wantErr: "MAP entry is not a [key, value] pair",
		},
		{
			name:    "map entry that isn't an array",
			raw:     oneValue(`{"type": "MULTISET", "nullable": true, "element_type": {"type": "VARCHAR", "nullable": true}}`, `["k"]`),
			wantErr: "MULTISET entry is not a [key, value] pair",
		},
		{
			name:    "unknown map key type",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "NOT_A_REAL_TYPE", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `[]`),
			wantErr: `unsupported result field type "NOT_A_REAL_TYPE"`,
		},
		{
			name:    "unknown map value type",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "NOT_A_REAL_TYPE", "nullable": true}}`, `[]`),
			wantErr: `unsupported result field type "NOT_A_REAL_TYPE"`,
		},
		{
			name:    "bad map key",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `[[["k"], "1"]]`),
			wantErr: "VARCHAR value is not a string",
		},
		{
			name:    "bad map value",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `[["k", ["1"]]]`),
			wantErr: "INTEGER value is not a string",
		},
		{
			name:    "map sent as a string",
			raw:     oneValue(`{"type": "MAP", "nullable": true, "key_type": {"type": "VARCHAR", "nullable": false}, "value_type": {"type": "INTEGER", "nullable": true}}`, `"k=1"`),
			wantErr: "MAP value is not an array",
		},
		{
			name:    "row with the wrong number of fields",
			raw:     oneValue(`{"type": "ROW", "nullable": true, "fields": [{"name": "a", "field_type": {"type": "INTEGER", "nullable": true}}]}`, `["1", "2"]`),
			wantErr: "ROW value has 2 fields but its type has 1",
		},
		{
			name:    "row with fewer values than fields",
			raw:     oneValue(`{"type": "ROW", "nullable": true, "fields": [{"name": "a", "field_type": {"type": "INTEGER", "nullable": true}}, {"name": "b", "field_type": {"type": "INTEGER", "nullable": true}}]}`, `["1"]`),
			wantErr: "ROW value has 1 fields but its type has 2",
		},
		{
			name:    "row sent as a string",
			raw:     oneValue(`{"type": "ROW", "nullable": true, "fields": [{"name": "a", "field_type": {"type": "INTEGER", "nullable": true}}]}`, `"(1)"`),
			wantErr: "ROW value is not an array",
		},
		{
			name:    "bad row field",
			raw:     oneValue(`{"type": "ROW", "nullable": true, "fields": [{"name": "a", "field_type": {"type": "INTEGER", "nullable": true}}]}`, `[["1"]]`),
			wantErr: "INTEGER value is not a string",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := convertResult(parseResult(t, test.raw))
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func testRows(t *testing.T) ([]string, [][]types.StatementResultField) {
	headers, rows, err := convertResult(parseResult(t, `{
		"result_format": "JSON",
		"schema": {"columns": [
			{"name": "status", "type": {"type": "VARCHAR", "nullable": true}},
			{"name": "order_id", "type": {"type": "INTEGER", "nullable": true}},
			{"name": "tags", "type": {"type": "ARRAY", "nullable": true, "element_type": {"type": "VARCHAR", "nullable": true}}}
		]},
		"data": [["SHIPPED", "1021", ["a", "b"]], ["PENDING", null, []]]
	}`))
	require.NoError(t, err)
	return headers, rows
}

func TestPrintJSON(t *testing.T) {
	headers, rows := testRows(t)

	var buf bytes.Buffer
	require.NoError(t, printJSON(&buf, serializedRows(headers, rows)))
	// Keys keep the column order instead of being sorted.
	require.Equal(t, `[
  {
    "status": "SHIPPED",
    "order_id": 1021,
    "tags": ["a", "b"]
  },
  {
    "status": "PENDING",
    "order_id": null,
    "tags": []
  }
]
`, buf.String())

	buf.Reset()
	require.NoError(t, printJSON(&buf, serializedRows(headers, nil)))
	require.Equal(t, "[]\n", buf.String())
}

func TestPrintJSONMatchesFlinkQueryForLongArrays(t *testing.T) {
	// Rows are pretty-printed one at a time, as in `confluent flink query`, so this array isn't wrapped.
	row := orderedRow{{key: "c_ids", value: []any{strings.Repeat("x", 64)}}}

	var buf bytes.Buffer
	require.NoError(t, printJSON(&buf, []orderedRow{row}))
	require.Equal(t, "[\n  {\n    \"c_ids\": [\""+strings.Repeat("x", 64)+"\"]\n  }\n]\n", buf.String())
}

func TestPrintYAML(t *testing.T) {
	headers, rows := testRows(t)

	var buf bytes.Buffer
	require.NoError(t, printYAML(&buf, serializedRows(headers, rows)))
	// Matches `confluent flink query`.
	require.Equal(t, `- status: SHIPPED
  order_id: 1021
  tags:
      - a
      - b
- status: PENDING
  order_id: null
  tags: []
`, buf.String())

	buf.Reset()
	require.NoError(t, printYAML(&buf, serializedRows(headers, nil)))
	require.Equal(t, "[]\n", buf.String())
}

func TestPrintTable(t *testing.T) {
	headers, rows := testRows(t)

	var buf bytes.Buffer
	printTable(&buf, headers, rows)
	require.Equal(t, `+---------+----------+--------+
| status  | order_id |  tags  |
+---------+----------+--------+
| SHIPPED | 1021     | [a, b] |
| PENDING | NULL     | []     |
+---------+----------+--------+
`, buf.String())
}

func TestPrintTableDoesNotWrap(t *testing.T) {
	// A long value with spaces isn't wrapped.
	value := strings.TrimSpace(strings.Repeat("word ", 10))
	var buf bytes.Buffer
	printTable(&buf, []string{"note"}, [][]types.StatementResultField{{types.AtomicStatementResultField{Type: types.Varchar, Value: value}}})
	require.Contains(t, buf.String(), "| "+value+" |")
}

func TestPrintTableEscapesControlCharacters(t *testing.T) {
	headers := []string{"note\x1b[2J"}
	rows := [][]types.StatementResultField{{types.AtomicStatementResultField{Type: types.Varchar, Value: "\x1b[31mred\x1b[0m\nfake row"}}}

	var buf bytes.Buffer
	printTable(&buf, headers, rows)
	require.NotContains(t, buf.String(), "\x1b")
	require.Contains(t, buf.String(), `\x1b[31mred\x1b[0m\x0afake row`)
	require.Contains(t, buf.String(), `note\x1b[2J`)
}

func TestEscapeControlChars(t *testing.T) {
	require.Equal(t, "plain text", escapeControlChars("plain text"))
	require.Equal(t, "tab\tkept", escapeControlChars("tab\tkept"))
	require.Equal(t, `line\x0abreak\x0d`, escapeControlChars("line\nbreak\r"))
	require.Equal(t, `\x1b[0m`, escapeControlChars("\x1b[0m"))
	require.Equal(t, `a\x7fb`, escapeControlChars("a\x7fb"))
	require.Equal(t, `\x9b31m`, escapeControlChars("\u009b31m"))
}

func TestPrintYAMLKeepsTrailingNewlines(t *testing.T) {
	// Trailing blank lines in a string value are kept.
	for _, value := range []string{"a\n\n", "line1\nline2\n\n\n", "plain"} {
		var buf bytes.Buffer
		require.NoError(t, printYAML(&buf, []orderedRow{{{key: "first", value: value}, {key: "last", value: value}}}))

		var parsed []map[string]string
		require.NoError(t, yaml.Unmarshal(buf.Bytes(), &parsed))
		require.Equal(t, []map[string]string{{"first": value, "last": value}}, parsed, "%q", value)
	}
}

func TestPrintWriteErrors(t *testing.T) {
	headers, rows := testRows(t)
	require.Error(t, printJSON(failingWriter{}, serializedRows(headers, rows)))
	require.Error(t, printJSON(failingWriter{}, nil))
	require.Error(t, printYAML(failingWriter{}, serializedRows(headers, rows)))
	require.Error(t, printYAML(failingWriter{}, nil))
}

func idRows(count int) [][]types.StatementResultField {
	rows := make([][]types.StatementResultField, count)
	for i := range rows {
		rows[i] = []types.StatementResultField{types.AtomicStatementResultField{Type: types.Integer, Value: strconv.Itoa(i + 1)}}
	}
	return rows
}

func TestApplyRowLimits(t *testing.T) {
	tests := []struct {
		name          string
		rows          int
		maxRows       int
		wantRows      int
		wantTruncated bool
	}{
		{"the extra row asked for shows truncation", 101, 100, 100, true},
		{"exactly the cap is not truncated", 100, 100, 100, false},
		{"no cap keeps every row", 300, 0, 300, false},
		{"no rows", 0, 0, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := applyRowLimits([]string{"id"}, idRows(test.rows), test.maxRows, false)
			require.Len(t, result.rows, test.wantRows)
			require.Equal(t, test.wantTruncated, result.truncated)
		})
	}
}

func TestPrintResult(t *testing.T) {
	tests := []struct {
		name       string
		format     output.Format
		result     queryResult
		wantStdout string
		wantStderr string
	}{
		{
			name:       "--max-rows warning goes to stderr, before json rows on stdout",
			format:     output.JSON,
			result:     applyRowLimits([]string{"id"}, idRows(3), 2, false),
			wantStdout: "[\n  {\n    \"id\": 1\n  },\n  {\n    \"id\": 2\n  }\n]\n",
			wantStderr: "Warning: stopped after 2 rows because of the `--max-rows` flag. The result set below is truncated.\n",
		},
		{
			name:       "preview notice goes to stderr, after the table on stdout",
			format:     output.Human,
			result:     applyRowLimits([]string{"id"}, idRows(2), 1, true),
			wantStdout: "+----+\n| id |\n+----+\n| 1  |\n+----+\n",
			wantStderr: "\nOnly showing the first 1 rows. Use `--max-rows` to show more, or `-o json`/`-o yaml` for the full result.\n",
		},
		{
			name:       "no rows leaves stdout empty",
			format:     output.Human,
			result:     applyRowLimits([]string{"id"}, nil, 100, true),
			wantStderr: "The query returned no rows.\n",
		},
		{
			name:       "no rows is an empty json array",
			format:     output.JSON,
			result:     applyRowLimits([]string{"id"}, nil, 0, false),
			wantStdout: "[]\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			require.NoError(t, printResult(&stdout, &stderr, test.format, test.result))
			require.Equal(t, test.wantStdout, stdout.String())
			require.Equal(t, test.wantStderr, stderr.String())
		})
	}
}

func TestPrintResultRejectsRepeatedColumns(t *testing.T) {
	headers := []string{"KEY", "ID", "KEY"}
	row := []types.StatementResultField{
		types.AtomicStatementResultField{Type: types.Varchar, Value: "a"},
		types.AtomicStatementResultField{Type: types.Integer, Value: "1"},
		types.AtomicStatementResultField{Type: types.Varchar, Value: "b"},
	}
	result := applyRowLimits(headers, [][]types.StatementResultField{row, row}, 1, false)

	for _, format := range []output.Format{output.JSON, output.YAML} {
		var stdout, stderr bytes.Buffer
		err := printResult(&stdout, &stderr, format, result)
		require.EqualError(t, err, `the result has more than one column named "KEY"`)
		require.Empty(t, stdout.String())
		require.Empty(t, stderr.String(), "no --max-rows warning before the error")
	}

	var stdout, stderr bytes.Buffer
	require.NoError(t, printResult(&stdout, &stderr, output.Human, result))
	require.Contains(t, stdout.String(), "| KEY | ID | KEY |")
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
