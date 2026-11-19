package tableflow

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"
)

// newTableFormatFlagsTestCommand declares --table-formats with the default of enable ("ICEBERG") or update ("").
func newTableFormatFlagsTestCommand(t *testing.T, tableFormatsDefault string, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("table-formats", tableFormatsDefault, "")
	addTableFormatPropertiesFlags(cmd, tableFormatsDefault == "")
	require.NoError(t, cmd.ParseFlags(args))
	return cmd
}

var tableFormatPropertiesTests = []struct {
	name    string
	args    []string
	want    string
	wantErr string
}{
	{name: "no flags", want: "null"},
	{name: "format version", args: []string{"--iceberg-format-version", "3"}, want: `[{"format":"ICEBERG","enabled":true,"format_version":3}]`},
	{name: "format version with lowercase iceberg", args: []string{"--iceberg-format-version", "3", "--table-formats", "iceberg"}, wantErr: "`--iceberg-format-version` requires `--table-formats ICEBERG`"},
	{name: "format version of zero is still sent", args: []string{"--iceberg-format-version", "0"}, want: `[{"format":"ICEBERG","enabled":true,"format_version":0}]`},
}

func TestTableFormatPropertiesForCreate(t *testing.T) {
	for _, test := range tableFormatPropertiesTests {
		t.Run(test.name, func(t *testing.T) {
			entries, err := tableFormatPropertiesForCreate(newTableFormatFlagsTestCommand(t, icebergFormat, test.args...))
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			body, err := json.Marshal(entries)
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(body))
		})
	}
}

func TestTableFormatPropertiesForUpdate(t *testing.T) {
	for _, test := range tableFormatPropertiesTests {
		t.Run(test.name, func(t *testing.T) {
			entries, err := tableFormatPropertiesForUpdate(newTableFormatFlagsTestCommand(t, "", test.args...))
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			body, err := json.Marshal(entries)
			require.NoError(t, err)
			require.JSONEq(t, test.want, string(body))
		})
	}
}

func TestGetIcebergFormatVersion(t *testing.T) {
	iceberg := func(enabled bool, version *int32) tableflowv1.TableflowV1TableflowTopicSpecTableFormatPropertiesOneOf {
		return tableflowv1.TableflowV1IcebergTableFormatPropertiesAsTableflowV1TableflowTopicSpecTableFormatPropertiesOneOf(&tableflowv1.TableflowV1IcebergTableFormatProperties{
			Format:        "ICEBERG",
			Enabled:       enabled,
			FormatVersion: version,
		})
	}
	delta := tableflowv1.TableflowV1DeltaTableFormatPropertiesAsTableflowV1TableflowTopicSpecTableFormatPropertiesOneOf(&tableflowv1.TableflowV1DeltaTableFormatProperties{
		Format:  "DELTA",
		Enabled: true,
	})
	topic := func(properties ...tableflowv1.TableflowV1TableflowTopicSpecTableFormatPropertiesOneOf) tableflowv1.TableflowV1TableflowTopic {
		return tableflowv1.TableflowV1TableflowTopic{Spec: &tableflowv1.TableflowV1TableflowTopicSpec{TableFormatProperties: &properties}}
	}

	tests := []struct {
		name  string
		topic tableflowv1.TableflowV1TableflowTopic
		want  int32
	}{
		{name: "ICEBERG enabled at 3", topic: topic(delta, iceberg(true, tableflowv1.PtrInt32(3))), want: 3},
		{name: "ICEBERG disabled with a stored version", topic: topic(delta, iceberg(false, tableflowv1.PtrInt32(3))), want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, getIcebergFormatVersion(test.topic))
		})
	}
}
