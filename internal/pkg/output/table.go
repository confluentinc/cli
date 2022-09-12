package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type table struct {
	writer       io.Writer
	objects      []interface{}
	fields       []string
	headerFields []string
	rows         [][]string
	format       Format
	isList       bool
}

// NewTable initializes a table, capable of printing a single object with Set(), or multiple objects with Add().
func NewTable(cmd *cobra.Command, fields []string) *table {
	return &table{
		writer: cmd.OutOrStdout(),
		fields: fields,
		format: GetFormat(cmd),
	}
}

// Set to a Nx2 table with the fields and values of a single object.
func (t *table) Set(i interface{}) {
	if t.format.IsSerialized() {
		t.objects = []interface{}{i}
		return
	}

	v := reflect.ValueOf(i).Elem()
	ty := reflect.TypeOf(i).Elem()
	t.rows = make([][]string, len(t.fields))
	for i, name := range t.fields {
		field, _ := ty.FieldByName(name)
		t.rows[i] = []string{field.Tag.Get("human"), fmt.Sprintf("%v", v.FieldByName(name))}
	}
}

// Add a row to a NxM table with the fields and values of multiple objects.
func (t *table) Add(i interface{}) {
	t.isList = true

	if t.format.IsSerialized() {
		t.objects = append(t.objects, i)
		t.headerFields = nil
		return
	}

	if t.headerFields == nil {
		ty := reflect.TypeOf(i).Elem()
		t.headerFields = make([]string, len(t.fields))
		for i, name := range t.fields {
			field, _ := ty.FieldByName(name)
			t.headerFields[i] = field.Tag.Get("human")
		}
	}

	v := reflect.ValueOf(i).Elem()
	row := make([]string, len(t.fields))
	for i, field := range t.fields {
		row[i] = fmt.Sprintf("%v", v.FieldByName(field))
	}
	t.rows = append(t.rows, row)
}

func (t *table) Sort() {
	sort.Slice(t.rows, func(i, j int) bool {
		for x := range t.rows[i] {
			if t.rows[i][x] != t.rows[j][x] {
				return t.rows[i][x] < t.rows[j][x]
			}
		}
		return false
	})
}

func (t *table) Print() error {
	return t.PrintWithAutoWrap(true)
}

func (t *table) PrintWithAutoWrap(auto bool) error {
	if t.format.IsSerialized() {
		v := t.objects[0]
		if t.isList {
			v = t.objects
		}

		switch t.format {
		default:
			out, err := json.Marshal(v)
			if err != nil {
				return err
			}
			_, err = t.writer.Write(pretty.Pretty(out))
			return err
		case YAML:
			return yaml.NewEncoder(t.writer).Encode(v)
		}
	}

	w := tablewriter.NewWriter(t.writer)

	if t.isList {
		w.SetAutoFormatHeaders(false)
		w.SetBorder(false)
		w.SetHeader(t.headerFields)
	}

	w.AppendBulk(t.rows)
	w.SetAutoWrapText(auto)
	w.Render()

	return nil
}
