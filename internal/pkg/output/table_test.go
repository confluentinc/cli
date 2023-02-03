package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/pretty"
)

type out struct {
	Id          int    `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Description string `human:"Description,omitempty" serialized:"description,omitempty"`
}

func TestTable(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+-------------+-----------------+",
			"| ID          |               1 |",
			"| Name        | lkc-123456      |",
			"| Description | Example Cluster |",
			"+-------------+-----------------+",
		},
		JSON.String(): {
			"{",
			`  "id": 1,`,
			`  "name": "lkc-123456",`,
			`  "description": "Example Cluster"`,
			"}",
		},
		YAML.String(): {
			"id: 1",
			"name: lkc-123456",
			"description: Example Cluster",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(&out{
			Id:          1,
			Name:        "lkc-123456",
			Description: "Example Cluster",
		})

		err := table.Print()
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestTable_NoAutoWrap(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+-------------+------------+",
			"| ID          |          1 |",
			"| Name        | lkc-123456 |",
			"| Description | {          |",
			`|             |   "A": 1,  |`,
			`|             |   "B": 2   |`,
			"|             | }          |",
			"+-------------+------------+",
		},
		JSON.String(): {
			"{",
			`  "id": 1,`,
			`  "name": "lkc-123456",`,
			`  "description": "{\n  \"A\": 1,\n  \"B\": 2\n}"`,
			"}",
		},
		YAML.String(): {
			"id: 1",
			"name: lkc-123456",
			"description: |-",
			"  {",
			`    "A": 1,`,
			`    "B": 2`,
			"  }",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)

		j, _ := json.Marshal(&struct {
			A int
			B int
		}{
			A: 1,
			B: 2,
		})

		table.Add(&out{
			Id:          1,
			Name:        "lkc-123456",
			Description: strings.TrimSpace(string(pretty.Pretty(j))),
		})

		err := table.PrintWithAutoWrap(false)
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestTable_Filter(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+-------------+-----------------+",
			"| Name        | lkc-123456      |",
			"| Description | Example Cluster |",
			"+-------------+-----------------+",
		},
		JSON.String(): {
			"{",
			`  "name": "lkc-123456",`,
			`  "description": "Example Cluster"`,
			"}",
		},
		YAML.String(): {
			"name: lkc-123456",
			"description: Example Cluster",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(&out{
			Id:          1,
			Name:        "lkc-123456",
			Description: "Example Cluster",
		})

		table.Filter([]string{"Name", "Description"})
		err := table.Print()
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestTable_OmitEmpty(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+------+------------+",
			"| ID   |          1 |",
			"| Name | lkc-123456 |",
			"+------+------------+",
		},
		JSON.String(): {
			"{",
			`  "id": 1,`,
			`  "name": "lkc-123456"`,
			"}",
		},
		YAML.String(): {
			"id: 1",
			"name: lkc-123456",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(&out{
			Id:   1,
			Name: "lkc-123456",
		})

		err := table.Print()
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestTable_Map(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"+---+-------+",
			"| A | apple |",
			"+---+-------+",
		},
		JSON.String(): {
			"{",
			`  "A": "apple"`,
			"}",
		},
		YAML.String(): {
			"A: apple",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(map[string]string{"A": "apple"})

		err := table.Print()
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestTable_EmptyMap(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"None found.",
		},
		JSON.String(): {
			"{}",
		},
		YAML.String(): {
			"{}",
		},
	}

	for format, expected := range tests {
		buf := new(bytes.Buffer)
		cmd := &cobra.Command{}
		cmd.Flags().String("output", format, "")
		cmd.SetOut(buf)

		table := NewTable(cmd)
		table.Add(map[string]string{})

		err := table.Print()
		require.NoError(t, err)

		require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
	}
}

func TestList(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"  ID |    Name    | Description  ",
			"-----+------------+--------------",
			"   1 | lkc-111111 | Cluster 1    ",
			"   2 | lkc-222222 | Cluster 2    ",
		},
		JSON.String(): {
			"[",
			"  {",
			`    "id": 1,`,
			`    "name": "lkc-111111",`,
			`    "description": "Cluster 1"`,
			"  },",
			"  {",
			`    "id": 2,`,
			`    "name": "lkc-222222",`,
			`    "description": "Cluster 2"`,
			"  }",
			"]",
		},
		YAML.String(): {
			"- id: 1",
			"  name: lkc-111111",
			"  description: Cluster 1",
			"- id: 2",
			"  name: lkc-222222",
			"  description: Cluster 2",
		},
	}

	objects := []any{
		&out{
			Id:          2,
			Name:        "lkc-222222",
			Description: "Cluster 2",
		},
		&out{
			Id:          1,
			Name:        "lkc-111111",
			Description: "Cluster 1",
		},
	}

	for format, expected := range tests {
		testList(t, format, objects, expected)
	}
}

func TestList_Empty(t *testing.T) {
	tests := map[string][]string{
		Human.String(): {
			"None found.",
		},
		JSON.String(): {
			"[]",
		},
		YAML.String(): {
			"[]",
		},
	}

	for format, expected := range tests {
		testList(t, format, []any{}, expected)
	}
}

func testList(t *testing.T, format string, objects []any, expected []string) {
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.Flags().String("output", format, "")
	cmd.SetOut(buf)

	list := NewList(cmd)
	for _, object := range objects {
		list.Add(object)
	}

	err := list.Print()
	require.NoError(t, err)

	require.Equal(t, strings.Join(expected, "\n")+"\n", buf.String(), format)
}
