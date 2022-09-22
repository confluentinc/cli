package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/sevlyar/retag"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type table struct {
	writer   io.Writer
	format   Format
	resource string
	objects  []interface{}
	filter   []string
	sort     bool
}

// NewTable initializes a table capable of printing a single object.
func NewTable(cmd *cobra.Command) *table {
	return &table{
		writer: cmd.OutOrStdout(),
		format: GetFormat(cmd),
	}
}

// NewList initializes a table capable of printing multiple objects.
func NewList(cmd *cobra.Command, resource string) *table {
	table := NewTable(cmd)
	table.resource = resource
	table.sort = true
	return table
}

func (t *table) Add(i interface{}) {
	if t.isList() {
		t.objects = append(t.objects, i)
	} else {
		t.objects = []interface{}{i}
	}
}

// Filter allows for printing a specific subset or ordering of fields
func (t *table) Filter(fields []string) {
	t.filter = fields
}

func (t *table) Sort(sort bool) {
	t.sort = sort
}

func (t *table) Print() error {
	return t.PrintWithAutoWrap(true)
}

func (t *table) PrintWithAutoWrap(auto bool) error {
	for i := range t.objects {
		hider := FieldHider{
			format: t.format,
			filter: &t.filter,
		}
		t.objects[i] = retag.Convert(t.objects[i], hider)
	}

	if t.sort {
		sort.Slice(t.objects, func(i, j int) bool {
			for k := 0; k < reflect.TypeOf(t.objects[i]).Elem().NumField(); k++ {
				vi := fmt.Sprintf("%v", reflect.ValueOf(t.objects[i]).Elem().Field(k))
				vj := fmt.Sprintf("%v", reflect.ValueOf(t.objects[j]).Elem().Field(k))
				if vi != vj {
					return vi < vj
				}
			}
			return false
		})
	}

	if t.format.IsSerialized() {
		var v interface{}
		if t.isList() {
			v = t.objects
			if len(t.objects) == 0 {
				v = []interface{}{}
			}
		} else {
			v = t.objects[0]
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

	if len(t.objects) == 0 {
		_, err := fmt.Fprintf(t.writer, "No %ss found.\n", t.resource)
		return err
	}

	w := tablewriter.NewWriter(t.writer)
	w.SetAutoWrapText(auto)

	if t.isList() {
		var header []string
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			if tag := reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()); tag != "-" {
				header = append(header, tag)
			}
		}

		w.SetAutoFormatHeaders(false)
		w.SetBorder(false)
		w.SetHeader(header)

		for _, object := range t.objects {
			var row []string
			for i := 0; i < reflect.TypeOf(object).Elem().NumField(); i++ {
				if tag := reflect.TypeOf(object).Elem().Field(i).Tag.Get(t.format.String()); tag != "-" {
					row = append(row, fmt.Sprintf("%v", reflect.ValueOf(object).Elem().Field(i)))
				}
			}
			w.Append(row)
		}
	} else {
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			if tag := reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()); tag != "-" {
				w.Append([]string{tag, fmt.Sprintf("%v", reflect.ValueOf(t.objects[0]).Elem().Field(i))})
			}
		}
	}

	w.Render()

	return nil
}

func (t *table) isList() bool {
	return t.resource != ""
}
