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

	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// queryColumnOut is one column of the result schema in the -o json/yaml output.
type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

// escapeControlChars neutralizes control characters (e.g. raw ANSI escape codes)
// before a field value reaches the terminal. JSON/YAML already escape these via
// their own serializers; this is the equivalent for the plain-table renderer,
// which otherwise lets the terminal execute them (recoloring, moving the cursor).
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

// needsEscape reports whether r is a control character this renderer neutralizes.
// Tab is left as-is: it's a legitimate data character (e.g. inside a multi-line
// VARCHAR) and carries no cursor/injection risk. Newline and carriage return
// stay escaped despite also being "legitimate" — a raw one in a table cell
// breaks the layout or lets a value inject fake rows — as do ANSI/cursor escapes
// like \x1b, the original reason this exists.
func needsEscape(r rune) bool {
	return unicode.IsControl(r) && r != '\t'
}

// newResultStreamer picks the page-by-page streamer for the requested output
// format, all writing to stdout.
func newResultStreamer(format output.Format, raw bool, name string) resultStreamer {
	switch format {
	case output.JSON:
		if raw {
			return newRawJSONArrayStreamer(os.Stdout)
		}
		return newJSONEnvelopeStreamer(os.Stdout)
	case output.YAML:
		return newYAMLStreamer(os.Stdout, raw)
	default:
		return newHumanTableStreamer(os.Stdout, name, false)
	}
}

// resultStreamer prints a query result incrementally, one page at a time, so the
// whole result never has to be held in memory. setColumns runs once (after the
// statement leaves PENDING, before any rows); writeRows runs per fetched page;
// close finishes with the metadata known only after the drain.
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
	return &jsonEnvelopeStreamer{w: bufio.NewWriter(w)}
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
	_, err := s.w.WriteString("rows:\n")
	return err
}

func (s *yamlStreamer) writeRows(rows []types.StatementResultRow) error {
	if !s.opened {
		if !s.raw {
			if err := s.writeColumnsBlock(); err != nil {
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

// ---- Human table ------------------------------------------------------------

// humanTableStreamer renders each page as its own aligned table block, so a large
// result never buffers. Column widths are computed per page, so they can differ
// between pages (the Snowflake/psql trade-off for streaming a table).
type humanTableStreamer struct {
	w             io.Writer
	name          string
	headers       []string
	showOperation bool
	wroteAny      bool
}

func newHumanTableStreamer(w io.Writer, name string, showOperation bool) *humanTableStreamer {
	return &humanTableStreamer{w: w, name: name, showOperation: showOperation}
}

func (s *humanTableStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = columnNames(columns)
	return nil
}

func (s *humanTableStreamer) writeRows(rows []types.StatementResultRow) error {
	if len(rows) == 0 {
		return nil
	}
	headers := s.headers
	if s.showOperation {
		headers = append([]string{"Operation"}, headers...)
	}
	escaped := make([]string, len(headers))
	for i, h := range headers {
		escaped[i] = escapeControlChars(h)
	}

	out := make([][]string, len(rows))
	for i, row := range rows {
		fields := make([]string, 0, len(headers))
		if s.showOperation {
			fields = append(fields, row.Operation.String())
		}
		for _, field := range row.GetFields() {
			fields = append(fields, escapeControlChars(field.ToString()))
		}
		out[i] = fields
	}

	table := tablewriter.NewWriter(s.w)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(escaped)
	table.AppendBulk(out)
	table.Render()
	s.wroteAny = true
	return nil
}

func (s *humanTableStreamer) close(phase string, _ int, _ bool) error {
	if !s.wroteAny {
		// Keep the notice off stdout so a piped/redirected run stays clean.
		output.ErrPrintf(false, "The query returned no rows. Statement \"%s\" is in phase %s.\n", s.name, phase)
	}
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
