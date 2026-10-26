package lightning

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"

	rtcesqlv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// nullField is a SQL NULL: "NULL" in the table, null in json/yaml.
var nullField = types.AtomicStatementResultField{Type: types.Null, Value: "NULL"}

// multisetCountType is the type of a MULTISET element's count.
var multisetCountType = rtcesqlv1.QueryV1DataType{Type: string(types.Integer)}

// convertResult converts a result to the field types `confluent flink query` uses, so both print values the same way.
func convertResult(result rtcesqlv1.QueryV1QueryResult) ([]string, [][]types.StatementResultField, error) {
	if format := result.GetResultFormat(); format != "JSON" {
		return nil, nil, fmt.Errorf(`unsupported result format "%s"`, format)
	}
	data := result.GetData()
	if data.String != nil {
		return nil, nil, errors.New("the query result's rows are not a JSON array")
	}

	columns := result.Schema.GetColumns()
	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.GetName()
	}

	var resultRows []rtcesqlv1.QueryV1ResultRow
	if data.ArrayOfQueryV1ResultRow != nil {
		resultRows = *data.ArrayOfQueryV1ResultRow
	}

	rows := make([][]types.StatementResultField, len(resultRows))
	for i, resultRow := range resultRows {
		if len(resultRow.Items) != len(columns) {
			return nil, nil, fmt.Errorf("result row %d has %d values, expected %d", i+1, len(resultRow.Items), len(columns))
		}

		row := make([]types.StatementResultField, len(columns))
		for j, value := range resultRow.Items {
			field, err := convertValue(value, columns[j].GetType())
			if err != nil {
				return nil, nil, fmt.Errorf(`failed to read column "%s": %w`, columns[j].GetName(), err)
			}
			row[j] = field
		}
		rows[i] = row
	}

	return headers, rows, nil
}

// convertValue converts one value; a JSON null is a SQL NULL, and a value that doesn't match its type is an error.
func convertValue(value rtcesqlv1.QueryV1ResultValue, dataType rtcesqlv1.QueryV1DataType) (types.StatementResultField, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, err
	}

	if value.String == nil && value.ArrayOfQueryV1ResultValue == nil {
		return nullField, nil
	}

	switch fieldType {
	case types.Array:
		return convertArray(value, dataType.GetElementType())
	case types.Multiset:
		return convertMap(value, fieldType, dataType.GetElementType(), multisetCountType)
	case types.Map:
		return convertMap(value, fieldType, dataType.GetKeyType(), dataType.GetValueType())
	case types.Row:
		return convertRow(value, dataType.GetFields())
	}

	if value.String == nil {
		return nil, fmt.Errorf("%s value is not a string", fieldType)
	}
	return types.AtomicStatementResultField{Type: fieldType, Value: *value.String}, nil
}

func convertArray(value rtcesqlv1.QueryV1ResultValue, elementType rtcesqlv1.QueryV1DataType) (types.StatementResultField, error) {
	elements, err := arrayElements(value, types.Array)
	if err != nil {
		return nil, err
	}

	elementFieldType, err := types.NewResultFieldType(elementType.GetType())
	if err != nil {
		return nil, err
	}

	values := make([]types.StatementResultField, len(elements))
	for i, element := range elements {
		if values[i], err = convertValue(element, elementType); err != nil {
			return nil, err
		}
	}

	return types.ArrayStatementResultField{Type: types.Array, ElementType: elementFieldType, Values: values}, nil
}

// convertMap converts a MAP, or a MULTISET as a map of element to count, from [key, value] pairs.
func convertMap(value rtcesqlv1.QueryV1ResultValue, fieldType types.StatementResultFieldType, keyType, valueType rtcesqlv1.QueryV1DataType) (types.StatementResultField, error) {
	pairs, err := arrayElements(value, fieldType)
	if err != nil {
		return nil, err
	}

	keyFieldType, err := types.NewResultFieldType(keyType.GetType())
	if err != nil {
		return nil, err
	}
	valueFieldType, err := types.NewResultFieldType(valueType.GetType())
	if err != nil {
		return nil, err
	}

	entries := make([]types.MapStatementResultFieldEntry, len(pairs))
	for i, pair := range pairs {
		if pair.ArrayOfQueryV1ResultValue == nil || len(*pair.ArrayOfQueryV1ResultValue) != 2 {
			return nil, fmt.Errorf("%s entry is not a [key, value] pair", fieldType)
		}
		keyAndValue := *pair.ArrayOfQueryV1ResultValue

		key, err := convertValue(keyAndValue[0], keyType)
		if err != nil {
			return nil, err
		}
		val, err := convertValue(keyAndValue[1], valueType)
		if err != nil {
			return nil, err
		}
		entries[i] = types.MapStatementResultFieldEntry{Key: key, Value: val}
	}

	return types.MapStatementResultField{Type: fieldType, KeyType: keyFieldType, ValueType: valueFieldType, Entries: entries}, nil
}

// convertRow converts a ROW, whose values are in the schema's field order.
func convertRow(value rtcesqlv1.QueryV1ResultValue, fields []rtcesqlv1.QueryV1RowFieldType) (types.StatementResultField, error) {
	elements, err := arrayElements(value, types.Row)
	if err != nil {
		return nil, err
	}
	if len(elements) != len(fields) {
		return nil, fmt.Errorf("%s value has %d fields but its type has %d", types.Row, len(elements), len(fields))
	}

	values := make([]types.StatementResultField, len(elements))
	elementTypes := make([]types.StatementResultFieldType, len(elements))
	for i, element := range elements {
		if values[i], err = convertValue(element, fields[i].GetFieldType()); err != nil {
			return nil, err
		}
		elementTypes[i] = values[i].GetType()
	}

	return types.RowStatementResultField{Type: types.Row, ElementTypes: elementTypes, Values: values}, nil
}

func arrayElements(value rtcesqlv1.QueryV1ResultValue, fieldType types.StatementResultFieldType) ([]rtcesqlv1.QueryV1ResultValue, error) {
	if value.ArrayOfQueryV1ResultValue == nil {
		return nil, fmt.Errorf("%s value is not an array", fieldType)
	}
	return *value.ArrayOfQueryV1ResultValue, nil
}

// printResult prints rows to stdout and notices to stderr, worded as in `confluent flink query`.
func printResult(stdout, stderr io.Writer, format output.Format, result queryResult) error {
	// The -o human default cap gets a notice after the table instead.
	if result.truncated && !result.humanDefaultCap {
		fmt.Fprintf(stderr, "Warning: stopped after %d rows because of the `--max-rows` flag. The result set below is truncated.\n", result.maxRows)
	}

	switch format {
	case output.JSON:
		return printJSON(stdout, serializedRows(result.headers, result.rows))
	case output.YAML:
		return printYAML(stdout, serializedRows(result.headers, result.rows))
	}

	if len(result.headers) == 0 || len(result.rows) == 0 {
		fmt.Fprintln(stderr, "The query returned no rows.")
		return nil
	}

	printTable(stdout, result.headers, result.rows)

	if result.truncated && result.humanDefaultCap {
		fmt.Fprintf(stderr, "\nOnly showing the first %d rows. Use `--max-rows` to show more, or `-o json`/`-o yaml` for the full result.\n", result.maxRows)
	}
	return nil
}

// printTable prints the -o human table with control characters escaped and no value truncated.
func printTable(w io.Writer, headers []string, rows [][]types.StatementResultField) {
	escapedHeaders := make([]string, len(headers))
	for i, header := range headers {
		escapedHeaders[i] = escapeControlChars(header)
	}

	tableRows := make([][]string, len(rows))
	for i, row := range rows {
		values := make([]string, len(row))
		for j, field := range row {
			values[j] = escapeControlChars(field.ToString())
		}
		tableRows[i] = values
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(escapedHeaders)
	table.AppendBulk(tableRows)
	table.Render()
}

func escapeControlChars(s string) string {
	if !strings.ContainsFunc(s, needsEscape) {
		return s
	}

	var sb strings.Builder
	for _, r := range s {
		if needsEscape(r) {
			fmt.Fprintf(&sb, "\\x%02x", r)
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// needsEscape keeps tabs; other control characters could break the table or inject terminal sequences.
func needsEscape(r rune) bool {
	return unicode.IsControl(r) && r != '\t'
}

type rowField struct {
	key   string
	value any
}

// orderedRow keeps columns in SELECT order, which a map would sort.
type orderedRow []rowField

func serializedRows(headers []string, rows [][]types.StatementResultField) []orderedRow {
	out := make([]orderedRow, len(rows))
	for i, row := range rows {
		fields := make(orderedRow, len(headers))
		for j, header := range headers {
			fields[j] = rowField{key: header, value: row[j].ToSerializedValue()}
		}
		out[i] = fields
	}
	return out
}

func (r orderedRow) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, field := range r {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := json.Marshal(field.key)
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(field.value)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(value)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// printJSON prints a JSON array formatted exactly like `confluent flink query` output.
func printJSON(w io.Writer, rows []orderedRow) error {
	if len(rows) == 0 {
		_, err := io.WriteString(w, "[]\n")
		return err
	}

	bw := bufio.NewWriter(w)
	_, _ = bw.WriteString("[\n")
	for i, row := range rows {
		if i > 0 {
			_, _ = bw.WriteString(",\n")
		}
		object, err := row.MarshalJSON()
		if err != nil {
			return err
		}
		_, _ = bw.Write(indentLines(bytes.TrimRight(pretty.Pretty(object), "\n"), "  "))
	}
	_, _ = bw.WriteString("\n]\n")
	// Flush returns any earlier write error.
	return bw.Flush()
}

// printYAML prints a YAML list formatted like `confluent flink query` output.
func printYAML(w io.Writer, rows []orderedRow) error {
	if len(rows) == 0 {
		_, err := io.WriteString(w, "[]\n")
		return err
	}

	bw := bufio.NewWriter(w)
	for _, row := range rows {
		item, err := yamlListItem(row)
		if err != nil {
			return err
		}
		_, _ = bw.Write(item)
	}
	return bw.Flush()
}

// yamlListItem renders one row as a YAML list item, keeping the column order.
func yamlListItem(row orderedRow) ([]byte, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, field := range row {
		var key, value yaml.Node
		if err := key.Encode(field.key); err != nil {
			return nil, err
		}
		if err := value.Encode(field.value); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, &key, &value)
	}

	b, err := yaml.Marshal(node)
	if err != nil {
		return nil, err
	}

	// Trim only the final newline; more may belong to a multi-line string.
	var out bytes.Buffer
	for i, line := range bytes.Split(bytes.TrimSuffix(b, []byte("\n")), []byte("\n")) {
		if i == 0 {
			out.WriteString("- ")
		} else {
			out.WriteString("  ")
		}
		out.Write(line)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

func indentLines(b []byte, indent string) []byte {
	lines := bytes.Split(b, []byte("\n"))
	var out bytes.Buffer
	out.Grow(len(b) + len(indent)*len(lines))
	for i, line := range lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(indent)
		out.Write(line)
	}
	return out.Bytes()
}
