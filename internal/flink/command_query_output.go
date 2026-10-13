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
// (json/yaml). Both emit a bare array of row objects. Human output is buffered
// (see printHumanTable), not handled here.
func newResultStreamer(format output.Format) resultStreamer {
	if format == output.YAML {
		return newSerialStreamer(os.Stdout, yamlRenderer{})
	}
	return newSerialStreamer(os.Stdout, jsonRenderer{})
}

// resultStreamer prints a serialized result one page at a time. setColumns runs
// once before any rows, writeRows per page, close after the drain.
type resultStreamer interface {
	setColumns(columns []flinkgatewayv1.ColumnDetails) error
	writeRows(rows []types.StatementResultRow) error
	close() error
}

func columnNames(columns []flinkgatewayv1.ColumnDetails) []string {
	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.GetName()
	}
	return headers
}

// rowField is one column's serialized value paired with its name.
type rowField struct {
	key   string
	value any
}

// orderedRow pairs a row's serialized values with their column names in schema
// (SELECT) order. Serialized output is emitted in this order, so columns keep
// their query order instead of being alphabetized by a map.
func orderedRow(headers []string, row types.StatementResultRow) []rowField {
	fields := row.GetFields()
	out := make([]rowField, 0, len(headers))
	for i, header := range headers {
		var value any
		if i < len(fields) {
			value = fields[i].ToSerializedValue()
		}
		out = append(out, rowField{key: header, value: value})
	}
	return out
}

// serialStreamer is the shared open → rows → close state machine for the
// serialized formats. It owns the paging/first-row/flush bookkeeping; a
// rowRenderer supplies the format-specific bytes.
type serialStreamer struct {
	w        *bufio.Writer
	renderer rowRenderer
	headers  []string
	opened   bool
	wroteRow bool
}

// rowRenderer turns one serialized format into byte producers. header runs once
// before the first row; row renders one row's value; between is written before
// every row after the first; close finishes the output — with hadRows=false it
// must emit the whole empty result (e.g. "[]"), since header was never written.
type rowRenderer interface {
	header() []byte
	row(fields []rowField) ([]byte, error)
	between() []byte
	close(hadRows bool) []byte
}

func newSerialStreamer(w io.Writer, renderer rowRenderer) *serialStreamer {
	return &serialStreamer{w: bufio.NewWriter(w), renderer: renderer}
}

func (s *serialStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = columnNames(columns)
	return nil
}

func (s *serialStreamer) writeRows(rows []types.StatementResultRow) error {
	for _, row := range rows {
		if !s.opened {
			if _, err := s.w.Write(s.renderer.header()); err != nil {
				return err
			}
			s.opened = true
		}
		if s.wroteRow {
			if _, err := s.w.Write(s.renderer.between()); err != nil {
				return err
			}
		}
		object, err := s.renderer.row(orderedRow(s.headers, row))
		if err != nil {
			return err
		}
		if _, err := s.w.Write(object); err != nil {
			return err
		}
		s.wroteRow = true
	}
	return s.w.Flush()
}

// close finishes the output. It takes no phase/row_count/truncated: a bare array
// carries no envelope metadata. phase and truncation surface via exit code /
// stderr instead; row_count is just the length of the emitted array.
func (s *serialStreamer) close() error {
	if _, err := s.w.Write(s.renderer.close(s.opened)); err != nil {
		return err
	}
	return s.w.Flush()
}

// jsonRowBytes pretty-prints one row object at the given indent, keeping the
// columns in schema order (json.Marshal of a map would sort them). pretty.Pretty
// preserves key order, so only the compact object has to be built by hand.
func jsonRowBytes(fields []rowField, indent string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, field := range fields {
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
	return indentLines(bytes.TrimRight(pretty.Pretty(buf.Bytes()), "\n"), indent), nil
}

// jsonRenderer writes rows as a pretty-printed bare JSON array.
type jsonRenderer struct{}

func (jsonRenderer) header() []byte                        { return []byte("[\n") }
func (jsonRenderer) row(fields []rowField) ([]byte, error) { return jsonRowBytes(fields, "  ") }
func (jsonRenderer) between() []byte                       { return []byte(",\n") }
func (jsonRenderer) close(hadRows bool) []byte {
	if !hadRows {
		return []byte("[]\n")
	}
	return []byte("\n]\n")
}

// yamlRenderer writes rows as a bare YAML list of row objects.
type yamlRenderer struct{}

func (yamlRenderer) header() []byte                        { return nil }
func (yamlRenderer) row(fields []rowField) ([]byte, error) { return yamlListItem(fields, "") }
func (yamlRenderer) between() []byte                       { return nil }
func (yamlRenderer) close(hadRows bool) []byte {
	if !hadRows {
		return []byte("[]\n")
	}
	return nil
}

// yamlListItem renders one row as a YAML sequence item ("- ...") at the given
// indent, keeping columns in schema order. A yaml.Node mapping preserves key
// order, which yaml.Marshal of a map would not.
func yamlListItem(fields []rowField, indent string) ([]byte, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, field := range fields {
		var keyNode, valueNode yaml.Node
		if err := keyNode.Encode(field.key); err != nil {
			return nil, err
		}
		if err := valueNode.Encode(field.value); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, &keyNode, &valueNode)
	}
	b, err := yaml.Marshal(node)
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
