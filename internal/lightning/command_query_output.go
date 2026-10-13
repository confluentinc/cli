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

	rtcesqlv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1alpha1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// nullField is a SQL NULL: "NULL" in the table, null in json/yaml.
var nullField = types.AtomicStatementResultField{Type: types.Null, Value: "NULL"}

// multisetCountType is the type of the count in each [element, count] pair of a MULTISET.
var multisetCountType = rtcesqlv1alpha1.QueryV1alpha1DataType{Type: string(types.Integer)}

// convertResult reads the column names and rows of an inline result. Values become the result
// field types `confluent flink query` renders, so both commands print a value the same way.
func convertResult(result rtcesqlv1alpha1.QueryV1alpha1QueryResult) ([]string, [][]types.StatementResultField, error) {
	// Rows are requested as JSON; anything else (including a response without a result) can't be read.
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

	var resultRows []rtcesqlv1alpha1.QueryV1alpha1ResultRow
	if data.ArrayOfQueryV1alpha1ResultRow != nil {
		resultRows = *data.ArrayOfQueryV1alpha1ResultRow
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

// convertValue converts one value of the given type. Scalars arrive as strings, and ARRAY, MAP,
// MULTISET, and ROW values as nested arrays; a JSON null is a SQL NULL at any depth. A value whose
// shape doesn't match its type is an error rather than a silent NULL.
func convertValue(value rtcesqlv1alpha1.QueryV1alpha1ResultValue, dataType rtcesqlv1alpha1.QueryV1alpha1DataType) (types.StatementResultField, error) {
	fieldType, err := types.NewResultFieldType(dataType.GetType())
	if err != nil {
		return nil, err
	}

	if value.String == nil && value.ArrayOfQueryV1alpha1ResultValue == nil {
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

func convertArray(value rtcesqlv1alpha1.QueryV1alpha1ResultValue, elementType rtcesqlv1alpha1.QueryV1alpha1DataType) (types.StatementResultField, error) {
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

// convertMap converts a MAP, or a MULTISET as a map from each element to its count. Both arrive as an
// array of two-element [key, value] arrays.
func convertMap(value rtcesqlv1alpha1.QueryV1alpha1ResultValue, fieldType types.StatementResultFieldType, keyType, valueType rtcesqlv1alpha1.QueryV1alpha1DataType) (types.StatementResultField, error) {
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
		if pair.ArrayOfQueryV1alpha1ResultValue == nil || len(*pair.ArrayOfQueryV1alpha1ResultValue) != 2 {
			return nil, fmt.Errorf("%s entry is not a [key, value] pair", fieldType)
		}
		keyAndValue := *pair.ArrayOfQueryV1alpha1ResultValue

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

// convertRow converts a ROW, whose values are positional: field names appear only in the schema.
func convertRow(value rtcesqlv1alpha1.QueryV1alpha1ResultValue, fields []rtcesqlv1alpha1.QueryV1alpha1RowFieldType) (types.StatementResultField, error) {
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

func arrayElements(value rtcesqlv1alpha1.QueryV1alpha1ResultValue, fieldType types.StatementResultFieldType) ([]rtcesqlv1alpha1.QueryV1alpha1ResultValue, error) {
	if value.ArrayOfQueryV1alpha1ResultValue == nil {
		return nil, fmt.Errorf("%s value is not an array", fieldType)
	}
	return *value.ArrayOfQueryV1alpha1ResultValue, nil
}

// printResult prints the rows to stdout as a table for -o human, or as a bare array of row objects for
// -o json and -o yaml. Notices go to stderr, so the output stays parseable, with the same wording as
// `confluent flink query`.
func printResult(stdout, stderr io.Writer, format output.Format, result queryResult) error {
	switch {
	case result.truncated && !result.humanDefaultCap:
		// The implicit -o human cap isn't user-requested, so it gets the notice after the table instead.
		fmt.Fprintf(stderr, "Warning: stopped after %d rows because of the `--max-rows` flag. The result set below is truncated.\n", result.maxRows)
	case result.serverLimited:
		fmt.Fprintf(stderr, "Warning: the result may be truncated by the server's row limit, which is %d rows by default. Narrow the query, for example with a more selective \"WHERE\" clause, to see the rest.\n", serverRowLimit)
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
		fmt.Fprintf(stderr, "\nOnly showing the first %d rows. Use `--max-rows` to show more, or `-o json`/`-o yaml` for every row the server returns.\n", result.maxRows)
	}
	return nil
}

// printTable renders the -o human table, escaping control characters in both headers and values so
// a result value can't inject terminal sequences. No column is truncated: the table may be piped, and
// shortening a value would corrupt whatever reads it.
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

// escapeControlChars neutralizes control characters (e.g. raw ANSI escapes) in a table header or value
// before it reaches the terminal.
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

// needsEscape reports whether r is a control character to neutralize. Tab is left as-is; a raw
// newline, carriage return, or ANSI escape in a table cell breaks the layout or injects fake rows.
func needsEscape(r rune) bool {
	return unicode.IsControl(r) && r != '\t'
}

// rowField is one column's serialized value paired with its name.
type rowField struct {
	key   string
	value any
}

// orderedRow is a row object whose keys keep the column (SELECT) order, which a map would sort.
type orderedRow []rowField

// serializedRows pairs each row's serialized values with their column names for -o json/-o yaml.
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

// printJSON prints the rows as a bare JSON array. Each row is pretty-printed on its own and indented
// under the array, as `confluent flink query` does, so the output is the same byte for byte.
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
	// A bufio.Writer keeps its first write error, so Flush reports any of them.
	return bw.Flush()
}

// printYAML prints the rows as a bare YAML list. Each row is rendered on its own under its "- "
// marker, as `confluent flink query` does, so nested values are indented the same way.
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

// yamlListItem renders one row as a YAML sequence item. A yaml.Node mapping keeps the column order,
// which yaml.Marshal of a map would sort.
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

	// Only the final newline is trimmed: more would be part of a multi-line string value.
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
