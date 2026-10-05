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
// (json/yaml). Both emit a bare array of row objects — no envelope. Human output
// is buffered (see printHumanTable), not handled here.
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
	close(phase string, rowCount int, truncated bool) error
}

func columnNames(columns []flinkgatewayv1.ColumnDetails) []string {
	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.GetName()
	}
	return headers
}

// rowMap keys a row's serialized values by column name.
func rowMap(headers []string, row types.StatementResultRow) map[string]any {
	fields := make(map[string]any, len(headers))
	for j, field := range row.GetFields() {
		fields[headers[j]] = field.ToSerializedValue()
	}
	return fields
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
	row(fields map[string]any) ([]byte, error)
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
		object, err := s.renderer.row(rowMap(s.headers, row))
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

// close ignores phase/row_count/truncated: a bare array carries no envelope
// metadata, so those are surfaced via exit code / stderr instead.
func (s *serialStreamer) close(_ string, _ int, _ bool) error {
	if _, err := s.w.Write(s.renderer.close(s.opened)); err != nil {
		return err
	}
	return s.w.Flush()
}

// jsonRowBytes pretty-prints one row object at the given indent, matching the
// buffered output byte-for-byte.
func jsonRowBytes(fields map[string]any, indent string) ([]byte, error) {
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	return indentLines(bytes.TrimRight(pretty.Pretty(encoded), "\n"), indent), nil
}

// jsonRenderer writes rows as a pretty-printed bare JSON array.
type jsonRenderer struct{}

func (jsonRenderer) header() []byte                            { return []byte("[\n") }
func (jsonRenderer) row(fields map[string]any) ([]byte, error) { return jsonRowBytes(fields, "  ") }
func (jsonRenderer) between() []byte                           { return []byte(",\n") }
func (jsonRenderer) close(hadRows bool) []byte {
	if !hadRows {
		return []byte("[]\n")
	}
	return []byte("\n]\n")
}

// yamlRenderer writes rows as a bare YAML list of row objects.
type yamlRenderer struct{}

func (yamlRenderer) header() []byte                            { return nil }
func (yamlRenderer) row(fields map[string]any) ([]byte, error) { return yamlListItem(fields, "") }
func (yamlRenderer) between() []byte                           { return nil }
func (yamlRenderer) close(hadRows bool) []byte {
	if !hadRows {
		return []byte("[]\n")
	}
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
