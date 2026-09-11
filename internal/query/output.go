package query

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/flink/query"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// engineSnapshot is the only value Engine ever takes today; see queryOut.Engine.
const engineSnapshot = "snapshot"

type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

type queryOut struct {
	StatementName string `json:"statement_name" yaml:"statement_name"`
	// Engine is always "snapshot" until M2 (Lightning routing) ships; emitting it
	// now means a script switching on this field today won't need to change later.
	Engine  string           `json:"engine" yaml:"engine"`
	Phase   string           `json:"phase" yaml:"phase"`
	Columns []queryColumnOut `json:"columns" yaml:"columns"`
	// Values carry their SQL type (number, null, etc); see
	// types.StatementResultField.ToSerializedValue.
	Rows      []map[string]any `json:"rows" yaml:"rows"`
	RowCount  int              `json:"row_count" yaml:"row_count"`
	Truncated bool             `json:"truncated" yaml:"truncated"`
	// AppendOnly is nil until traits are known; false means Rows is a changelog, not a materialized table.
	AppendOnly *bool `json:"append_only,omitempty" yaml:"append_only,omitempty"`
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

		var appendOnly *bool
		if appendOnlyKnown {
			appendOnly = &isAppendOnly
		}

		return output.SerializedOutput(cmd, &queryOut{
			StatementName: name,
			Engine:        engineSnapshot,
			Phase:         string(result.Phase()),
			Columns:       columns,
			Rows:          rows,
			RowCount:      len(rows),
			Truncated:     result.Truncated,
			AppendOnly:    appendOnly,
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
			fields = append(fields, field.ToString())
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
