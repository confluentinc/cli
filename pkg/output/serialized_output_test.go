package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type nestedJsonTagged struct {
	DisplayName string `json:"display_name"`
}

type jsonValueOut struct {
	Id           string    `human:"ID" serialized:"id"`
	Integrations JsonValue `human:"-" serialized:"integrations"`
}

func TestJsonValueKeepsNestedJsonNames(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+----+-------+",
			"| ID | int-1 |",
			"+----+-------+",
		},
		JSON.String(): {
			"{",
			`  "id": "int-1",`,
			`  "integrations": [`,
			"    {",
			`      "display_name": "Slack"`,
			"    }",
			"  ]",
			"}",
		},
		YAML.String(): {
			"id: int-1",
			"integrations:",
			"    - display_name: Slack",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(&jsonValueOut{Id: "int-1", Integrations: FromJsonTags([]nestedJsonTagged{{DisplayName: "Slack"}})})
		require.NoError(t, table.Print())

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestFromJsonTagsPrintsUnencodableValueAsNull(t *testing.T) {
	data, err := FromJsonTags(make(chan int)).MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "null", string(data))
}
