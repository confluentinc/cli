package query

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

// queryOut mirrors the PRD's canonical envelope shape (phase/columns/rows/
// row_count/truncated). statement_name, append_only, and engine were all
// dropped: the first two were Snapshot-only additions with no PRD equivalent;
// engine was Lightning-only, which this CLI doesn't support (per Florian).
type queryOut struct {
	Phase   string           `json:"phase" yaml:"phase"`
	Columns []queryColumnOut `json:"columns" yaml:"columns"`
	// Values carry their SQL type (number, null, etc); see
	// types.StatementResultField.ToSerializedValue.
	Rows      []map[string]any `json:"rows" yaml:"rows"`
	RowCount  int              `json:"row_count" yaml:"row_count"`
	Truncated bool             `json:"truncated" yaml:"truncated"`
}

func (c *command) printQueryResult(cmd *cobra.Command, name string, result *query.Result, isAppendOnly, appendOnlyKnown, raw bool) error {
	columns := make([]queryColumnOut, len(result.Columns))
	headers := make([]string, len(result.Columns))
	for i, column := range result.Columns {
		columnType := column.GetType()
		columns[i] = queryColumnOut{Name: column.GetName(), Type: columnType.GetType()}
		headers[i] = column.GetName()
	}

	showOperation := appendOnlyKnown && !isAppendOnly

	if output.GetFormat(cmd).IsSerialized() {
		rows := make([]map[string]any, len(result.Rows))
		for i, row := range result.Rows {
			// Every row is guaranteed len(headers) fields: Run() hard-errors on a
			// row/schema mismatch before this function ever sees a result.
			fields := make(map[string]any, len(headers))
			for j, field := range row.GetFields() {
				fields[headers[j]] = field.ToSerializedValue()
			}
			rows[i] = fields
		}

		// A bare array has nowhere to put the schema, so the envelope is the
		// default and --raw opts into the bare array.
		if raw {
			return output.SerializedOutput(cmd, rows)
		}

		return output.SerializedOutput(cmd, &queryOut{
			Phase:     string(result.Phase()),
			Columns:   columns,
			Rows:      rows,
			RowCount:  len(rows),
			Truncated: result.Truncated,
		})
	}

	if len(headers) == 0 || len(result.Rows) == 0 {
		output.ErrPrintf(false, "The query returned no rows. Statement \"%s\" is in phase %s.\n", name, result.Phase())
		return nil
	}

	if showOperation {
		headers = append([]string{"Operation"}, headers...)
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
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	table.AppendBulk(rows)
	table.Render()

	return nil
}

// escapeControlChars neutralizes control characters (e.g. raw ANSI escape codes)
// before a field value reaches the terminal. JSON/YAML already escape these via
// their own serializers; this is the equivalent for the plain-table renderer,
// which otherwise lets the terminal execute them (recoloring, moving the cursor).
func escapeControlChars(s string) string {
	if !strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}

	var sb strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			fmt.Fprintf(&sb, "\\x%02x", r)
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
