package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/sevlyar/retag"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

type kind int

const (
	table kind = iota
	mapping
	list
)

type Table struct {
	kind     kind
	writer   io.Writer
	resource string
	format   Format
	objects  []interface{}
	filter   []string
	sort     bool
}

// NewTable creates a table for printing a single object.
func NewTable(cmd *cobra.Command) *Table {
	return &Table{
		writer: cmd.OutOrStdout(),
		format: GetFormat(cmd),
	}
}

// NewMapping creates a table for printing key-value pairs.
func NewMapping(cmd *cobra.Command, resource string) *Table {
	table := NewTable(cmd)
	table.kind = mapping
	table.resource = resource
	return table
}

// NewList creates a table for printing multiple objects.
func NewList(cmd *cobra.Command, resource string) *Table {
	table := NewTable(cmd)
	table.kind = list
	table.resource = resource
	table.sort = true
	return table
}

func (t *Table) Add(i interface{}) {
	switch t.kind {
	case table, mapping:
		t.objects = []interface{}{i}
	case list:
		t.objects = append(t.objects, i)
	}
}

// Filter allows for printing a specific subset or ordering of fields
func (t *Table) Filter(fields []string) {
	t.filter = fields
}

func (t *Table) Sort(sort bool) {
	t.sort = sort
}

func (t *Table) Print() error {
	return t.PrintWithAutoWrap(true)
}

func (t *Table) PrintWithAutoWrap(auto bool) error {
	if t.kind != mapping {
		if t.format.IsSerialized() {
			for i := range t.objects {
				serializer := FieldSerializer{format: t.format}
				t.objects[i] = retag.Convert(t.objects[i], serializer)
			}
		}

		for i := range t.objects {
			hider := FieldHider{
				format: t.format,
				filter: &t.filter,
			}
			t.objects[i] = retag.Convert(t.objects[i], hider)
		}
	}

	if t.sort {
		sort.Slice(t.objects, func(i, j int) bool {
			for k := 0; k < reflect.TypeOf(t.objects[i]).Elem().NumField(); k++ {
				vi := reflect.ValueOf(t.objects[i]).Elem().Field(k)
				vj := reflect.ValueOf(t.objects[j]).Elem().Field(k)

				si := fmt.Sprintf("%v", vi)
				sj := fmt.Sprintf("%v", vj)
			
				if si != sj {
					return si < sj
				}
			}
			return false
		})
	}

	if t.format.IsSerialized() {
		var v interface{}
		switch t.kind {
		case table:
			v = t.objects[0]
		case mapping:
			v = t.objects[0].(map[string]string)
		case list:
			v = t.objects
			if len(t.objects) == 0 {
				v = []interface{}{}
			}
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

	isEmpty := false
	switch t.kind {
	case list:
		isEmpty = len(t.objects) == 0
	case mapping:
		isEmpty = len(t.objects[0].(map[string]string)) == 0
	}
	if isEmpty {
		_, err := fmt.Fprintf(t.writer, "No %ss found.\n", t.resource)
		return err
	}

	w := tablewriter.NewWriter(t.writer)
	w.SetAutoWrapText(auto)

	switch t.kind {
	case table:
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			tag := strings.Split(reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()), ",")
			val := reflect.ValueOf(t.objects[0]).Elem().Field(i)
			if tag[0] != "-" && !(utils.Contains(tag, "omitempty") && val.IsZero()) {
				w.Append([]string{tag[0], fmt.Sprintf("%v", val)})
			}
		}
	case mapping:
		for k, v := range t.objects[0].(map[string]string) {
			w.Append([]string{k, v})
		}
	case list:
		var header []string
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			tag := strings.Split(reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()), ",")
			if tag[0] != "-" {
				header = append(header, tag[0])
			}
		}

		w.SetAutoFormatHeaders(false)
		w.SetBorder(false)
		w.SetHeader(header)

		for _, object := range t.objects {
			var row []string
			for i := 0; i < reflect.TypeOf(object).Elem().NumField(); i++ {
				tag := strings.Split(reflect.TypeOf(object).Elem().Field(i).Tag.Get(t.format.String()), ",")
				if tag[0] != "-" {
					row = append(row, fmt.Sprintf("%v", reflect.ValueOf(object).Elem().Field(i)))
				}
			}
			w.Append(row)
		}
	}

	w.Render()

	return nil
}
