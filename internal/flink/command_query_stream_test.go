package flink

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/pretty"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

func column(name string) flinkgatewayv1.ColumnDetails {
	return flinkgatewayv1.ColumnDetails{Name: name}
}

func atomicRow(fields ...types.StatementResultField) types.StatementResultRow {
	return types.StatementResultRow{Fields: fields}
}

func intField(v string) types.StatementResultField {
	return types.AtomicStatementResultField{Type: types.Integer, Value: v}
}

func varcharField(v string) types.StatementResultField {
	return types.AtomicStatementResultField{Type: types.Varchar, Value: v}
}

// bufferedRawJSON reproduces the pre-streaming code path: build the whole slice
// of row maps and hand it to the same pretty-printer output.SerializedOutput uses
// for -o json. The streamer must produce these exact bytes.
func bufferedRawJSON(t *testing.T, headers []string, rows []types.StatementResultRow) []byte {
	t.Helper()
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		fields := make(map[string]any, len(headers))
		for j, field := range row.GetFields() {
			fields[headers[j]] = field.ToSerializedValue()
		}
		out[i] = fields
	}
	encoded, err := json.Marshal(out)
	require.NoError(t, err)
	return pretty.Pretty(encoded)
}

func TestRawJSONArrayStreamerMatchesBuffered(t *testing.T) {
	tests := []struct {
		name    string
		columns []flinkgatewayv1.ColumnDetails
		// pages is the row set split into the pages the streamer receives; the
		// buffered comparison flattens them, so page boundaries must not affect the
		// bytes.
		pages [][]types.StatementResultRow
	}{
		{
			name:    "empty",
			columns: []flinkgatewayv1.ColumnDetails{column("order_id"), column("status")},
			pages:   nil,
		},
		{
			name:    "single row",
			columns: []flinkgatewayv1.ColumnDetails{column("order_id"), column("status")},
			pages: [][]types.StatementResultRow{
				{atomicRow(intField("1021"), varcharField("SHIPPED"))},
			},
		},
		{
			name:    "many rows across pages",
			columns: []flinkgatewayv1.ColumnDetails{column("order_id"), column("status")},
			pages: [][]types.StatementResultRow{
				{
					atomicRow(intField("1021"), varcharField("SHIPPED")),
					atomicRow(intField("1044"), varcharField("PENDING")),
				},
				{
					atomicRow(intField("1077"), varcharField("CANCELLED")),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var flat []types.StatementResultRow
			for _, page := range test.pages {
				flat = append(flat, page...)
			}
			headers := make([]string, len(test.columns))
			for i, c := range test.columns {
				headers[i] = c.GetName()
			}

			var buf bytes.Buffer
			streamer := newRawJSONArrayStreamer(&buf)
			require.NoError(t, streamer.setColumns(test.columns))
			for _, page := range test.pages {
				require.NoError(t, streamer.writeRows(page))
			}
			require.NoError(t, streamer.close())

			require.Equal(t, string(bufferedRawJSON(t, headers, flat)), buf.String())
		})
	}
}
