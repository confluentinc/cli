package flink

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/tidwall/pretty"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

// rawJSONArrayStreamer writes result rows as a pretty-printed JSON array
// incrementally, one page at a time, so a large `--raw -o json` result never has
// to be held in memory all at once.
//
// Its output is byte-for-byte identical to the buffered path
// (output.SerializedOutput(cmd, rows) with -o json, i.e.
// pretty.Pretty(json.Marshal([]map[string]any{...}))), so existing golden files
// stay valid. Parity is guarded by TestRawJSONArrayStreamerMatchesBuffered.
//
// It streams only the bare --raw array. The default envelope carries phase,
// row_count and truncated, which are known only after the drain finishes, so
// streaming it would mean emitting those trailing fields after the rows (a
// key-order change worth a separate decision). -o yaml and the human table are
// likewise still buffered; see command_query_output.go.
type rawJSONArrayStreamer struct {
	w        *bufio.Writer
	headers  []string
	wroteRow bool
	opened   bool
}

func newRawJSONArrayStreamer(w io.Writer) *rawJSONArrayStreamer {
	return &rawJSONArrayStreamer{w: bufio.NewWriter(w)}
}

// setColumns records the column order so each row can be keyed by name. It
// satisfies query.Options.OnSchema.
func (s *rawJSONArrayStreamer) setColumns(columns []flinkgatewayv1.ColumnDetails) error {
	s.headers = make([]string, len(columns))
	for i, column := range columns {
		s.headers[i] = column.GetName()
	}
	return nil
}

// writeRows emits one page. It satisfies query.Options.OnRows.
func (s *rawJSONArrayStreamer) writeRows(rows []types.StatementResultRow) error {
	if !s.opened {
		if _, err := s.w.WriteString("[\n"); err != nil {
			return err
		}
		s.opened = true
	}

	for _, row := range rows {
		// Every row is guaranteed len(headers) fields: query.Run hard-errors on a
		// row/schema mismatch before this ever runs.
		fields := make(map[string]any, len(s.headers))
		for j, field := range row.GetFields() {
			fields[s.headers[j]] = field.ToSerializedValue()
		}

		encoded, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		// pretty.Pretty on the whole array indents each element by two spaces; we
		// reproduce that per row so the streamed bytes match the buffered array.
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

	// Flush per page so a long-running drain shows progress and stays bounded.
	return s.w.Flush()
}

// close finishes the array. An empty result renders as "[]\n", matching
// pretty.Pretty of an empty slice.
func (s *rawJSONArrayStreamer) close() error {
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

// indentLines prefixes every line of b with indent.
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
