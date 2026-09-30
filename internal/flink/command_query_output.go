package flink

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// queryColumnOut is one column of the result schema in the -o json/yaml output.
type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

// escapeControlChars neutralizes control characters (e.g. raw ANSI escapes) in a
// value before it reaches the terminal. JSON/YAML serializers already do this;
// this is the equivalent for the plain-table renderer.
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

// needsEscape reports whether r is a control character to neutralize. Tab is
// left as-is (legitimate, harmless); newline/CR and ANSI escapes are not — a raw
// one in a table cell breaks layout or injects fake rows.
func needsEscape(r rune) bool {
	return unicode.IsControl(r) && r != '\t'
}

// newResultStreamer picks the page-by-page streamer for a serialized format
// (json/yaml). Human output is buffered (see printHumanTable), not handled here.
func newResultStreamer(format output.Format, raw bool) resultStreamer {
	if format == output.YAML {
		return newYAMLStreamer(os.Stdout, raw)
	}
	if raw {
		return newRawJSONArrayStreamer(os.Stdout)
	}
	return newJSONEnvelopeStreamer(os.Stdout)
}

// resultStreamer prints a serialized result one page at a time. setColumns runs
// once before any rows, writeRows per page, close with the post-drain metadata.
type resultStreamer interface {
	setColumns(columns []flinkgatewayv1.ColumnDetails) error
	writeRows(rows []types.StatementResultRow) error
	close(phase string, rowCount int, truncated bool) error
}

func columnNames(columns []flinkgatewayv1.ColumnDetails) []string {
	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.GetName()
	}
	return headers
}

func columnOuts(columns []flinkgatewayv1.ColumnDetails) []queryColumnOut {
	out := make([]queryColumnOut, len(columns))
	for i, column := range columns {
		columnType := column.GetType()
		out[i] = queryColumnOut{Name: column.GetName(), Type: columnType.GetType()}
	}
	return out
}

// rowMap keys a row's serialized values by column name.
func rowMap(headers []string, row types.StatementResultRow) map[string]any {
	fields := make(map[string]any, len(headers))
	for j, field := range row.GetFields() {
		fields[headers[j]] = field.ToSerializedValue()
	}
	return fields
}

// ---- JSON: bare --raw array -------------------------------------------------

// rawJSONArrayStreamer writes rows as a pretty-printed JSON array, one page at a
// time. Output is byte-for-byte identical to the buffered --raw path.
type rawJSONArrayStreamer struct {
	w        *bufio.Writer
	headers  []string
	wroteRow bool
	opened   bool
}

func newRawJSONArrayStreamer(w io.Writer) *rawJSONArrayStreamer {
	return &rawJSONArrayStreamer{w: bufio.NewWriter(w)}
}

func (s *rawJSONArrayStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = columnNames(columns)
	return nil
}

func (s *rawJSONArrayStreamer) writeRows(rows []types.StatementResultRow) error {
	if !s.opened {
		if _, err := s.w.WriteString("[\n"); err != nil {
			return err
		}
		s.opened = true
	}
	for _, row := range rows {
		encoded, err := json.Marshal(rowMap(s.headers, row))
		if err != nil {
			return err
		}
		object := indentLines(bytes.TrimRight(pretty.Pretty(encoded), "\n"), "  ")
		if s.wroteRow {
			if _, err := s.w.WriteString(",\n"); err != nil {
				return err
			}
		}
		if _, err := s.w.Write(object); err != nil {
			return err
		}
		s.wroteRow = true
	}
	return s.w.Flush()
}

func (s *rawJSONArrayStreamer) close(string, int, bool) error {
	if !s.opened {
		if _, err := s.w.WriteString("[]\n"); err != nil {
			return err
		}
		return s.w.Flush()
	}
	if _, err := s.w.WriteString("\n]\n"); err != nil {
		return err
	}
	return s.w.Flush()
}

// ---- JSON: default envelope -------------------------------------------------

// jsonEnvelopeStreamer streams the schema+rows envelope. phase/row_count/truncated
// are known only after the drain, so they are emitted after the rows (a key-order
// change from the buffered form; JSON objects are unordered so it's harmless).
type jsonEnvelopeStreamer struct {
	w        *bufio.Writer
	headers  []string
	columns  []queryColumnOut
	wroteRow bool
	opened   bool
}

func newJSONEnvelopeStreamer(w io.Writer) *jsonEnvelopeStreamer {
	// columns starts as a non-nil empty slice so a schema-less statement (setColumns
	// never called) still marshals "columns": [], not "columns": null.
	return &jsonEnvelopeStreamer{w: bufio.NewWriter(w), columns: make([]queryColumnOut, 0)}
}

func (s *jsonEnvelopeStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = columnNames(columns)
	s.columns = columnOuts(columns)
	return nil
}

func (s *jsonEnvelopeStreamer) open() error {
	colsJSON, err := json.Marshal(s.columns)
	if err != nil {
		return err
	}
	cols := indentLines(bytes.TrimRight(pretty.Pretty(colsJSON), "\n"), "  ")
	if _, err := fmt.Fprintf(s.w, "{\n  \"columns\": %s,\n  \"rows\": [\n", bytes.TrimLeft(cols, " ")); err != nil {
		return err
	}
	s.opened = true
	return nil
}

func (s *jsonEnvelopeStreamer) writeRows(rows []types.StatementResultRow) error {
	if !s.opened {
		if err := s.open(); err != nil {
			return err
		}
	}
	for _, row := range rows {
		encoded, err := json.Marshal(rowMap(s.headers, row))
		if err != nil {
			return err
		}
		object := indentLines(bytes.TrimRight(pretty.Pretty(encoded), "\n"), "    ")
		if s.wroteRow {
			if _, err := s.w.WriteString(",\n"); err != nil {
				return err
			}
		}
		if _, err := s.w.Write(object); err != nil {
			return err
		}
		s.wroteRow = true
	}
	return s.w.Flush()
}

func (s *jsonEnvelopeStreamer) close(phase string, rowCount int, truncated bool) error {
	if !s.opened {
		if err := s.open(); err != nil {
			return err
		}
	}
	if s.wroteRow {
		if _, err := s.w.WriteString("\n"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(s.w, "  ],\n  \"phase\": %q,\n  \"row_count\": %d,\n  \"truncated\": %t\n}\n", phase, rowCount, truncated); err != nil {
		return err
	}
	return s.w.Flush()
}

// ---- YAML -------------------------------------------------------------------

// yamlStreamer streams the -o yaml forms (bare list with --raw, else envelope).
type yamlStreamer struct {
	w       *bufio.Writer
	headers []string
	columns []queryColumnOut
	raw     bool
	opened  bool
}

func newYAMLStreamer(w io.Writer, raw bool) *yamlStreamer {
	return &yamlStreamer{w: bufio.NewWriter(w), raw: raw}
}

func (s *yamlStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = columnNames(columns)
	s.columns = columnOuts(columns)
	return nil
}

// yamlListItem renders v as a YAML sequence item ("- ...") at the given indent.
func yamlListItem(v any, indent string) ([]byte, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(bytes.TrimRight(b, "\n"), []byte("\n"))
	var out bytes.Buffer
	for i, line := range lines {
		if i == 0 {
			out.WriteString(indent + "- ")
		} else {
			out.WriteString(indent + "  ")
		}
		out.Write(line)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

func (s *yamlStreamer) writeColumnsBlock() error {
	if len(s.columns) == 0 {
		// Empty must serialize as `columns: []`; a bare `columns:` parses as null.
		_, err := s.w.WriteString("columns: []\n")
		return err
	}
	if _, err := s.w.WriteString("columns:\n"); err != nil {
		return err
	}
	for _, c := range s.columns {
		item, err := yamlListItem(c, "  ")
		if err != nil {
			return err
		}
		if _, err := s.w.Write(item); err != nil {
			return err
		}
	}
	return nil
}

func (s *yamlStreamer) writeRows(rows []types.StatementResultRow) error {
	if !s.opened {
		if !s.raw {
			if err := s.writeColumnsBlock(); err != nil {
				return err
			}
			if _, err := s.w.WriteString("rows:\n"); err != nil {
				return err
			}
		}
		s.opened = true
	}
	indent := ""
	if !s.raw {
		indent = "  "
	}
	for _, row := range rows {
		item, err := yamlListItem(rowMap(s.headers, row), indent)
		if err != nil {
			return err
		}
		if _, err := s.w.Write(item); err != nil {
			return err
		}
	}
	return s.w.Flush()
}

func (s *yamlStreamer) close(phase string, rowCount int, truncated bool) error {
	if !s.opened {
		if s.raw {
			if _, err := s.w.WriteString("[]\n"); err != nil {
				return err
			}
			return s.w.Flush()
		}
		if err := s.writeColumnsBlock(); err != nil {
			return err
		}
		// No rows were ever written; emit an empty list so `rows` isn't null.
		if _, err := s.w.WriteString("rows: []\n"); err != nil {
			return err
		}
		s.opened = true
	}
	if s.raw {
		return s.w.Flush()
	}
	if _, err := fmt.Fprintf(s.w, "phase: %s\nrow_count: %d\ntruncated: %t\n", phase, rowCount, truncated); err != nil {
		return err
	}
	return s.w.Flush()
}

// ---- Human table (buffered) -------------------------------------------------

// printHumanResult renders the buffered -o human table after the drain (it needs
// the full row set to align columns).
func printHumanResult(w io.Writer, name string, result *query.Result, isAppendOnly, appendOnlyKnown bool) error {
	headers := columnNames(result.Columns)
	showOperation := appendOnlyKnown && !isAppendOnly
	return printHumanTable(w, name, headers, showOperation, result)
}

// printHumanTable renders the default -o human table, escaping control characters
// in both headers and values so a result value can't inject terminal sequences.
// The no-rows notice goes to stderr, not w, so a redirected table stays clean.
func printHumanTable(w io.Writer, name string, headers []string, showOperation bool, result *query.Result) error {
	if len(headers) == 0 || len(result.Rows) == 0 {
		output.ErrPrintf(false, "The query returned no rows. Statement \"%s\" is in phase %s.\n", name, result.Phase())
		return nil
	}

	if showOperation {
		headers = append([]string{"Operation"}, headers...)
	}
	for i, header := range headers {
		headers[i] = escapeControlChars(header)
	}

	rows := make([][]string, len(result.Rows))
	for i, row := range result.Rows {
		fields := make([]string, 0, len(headers))
		if showOperation {
			fields = append(fields, row.Operation.String())
		}
		for _, field := range row.GetFields() {
			fields = append(fields, escapeControlChars(field.ToString()))
		}
		rows[i] = fields
	}

	// No column truncation: this command is expected to be piped, and shortening
	// a value would corrupt whatever reads it.
	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	table.AppendBulk(rows)
	table.Render()

	return nil
}

// ---- helpers ----------------------------------------------------------------

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
