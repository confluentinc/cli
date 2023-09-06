package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/sevlyar/retag"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type Table struct {
	isList  bool
	writer  io.Writer
	format  Format
	objects []any
	filter  []string
	sort    bool
}

// NewTable creates a table for printing a single object.
func NewTable(cmd *cobra.Command) *Table {
	return &Table{
		writer: cmd.OutOrStdout(),
		format: GetFormat(cmd),
	}
}

// NewList creates a table for printing multiple objects.
func NewList(cmd *cobra.Command) *Table {
	table := NewTable(cmd)
	table.isList = true
	table.sort = true
	return table
}

func (t *Table) Add(i any) {
	if t.isList {
		t.objects = append(t.objects, i)
	} else {
		t.objects = []any{i}
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
	return t.printCore(t.writer, auto)
}

func (t *Table) PrintString() (string, error) {
	tableStr := &strings.Builder{}
	err := t.printCore(tableStr, true)

	return tableStr.String(), err
}

func (t *Table) printCore(writer io.Writer, auto bool) error {
	if !t.isMap() {
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
				// Don't sort the IsCurrent column, if it exists.
				if reflect.TypeOf(t.objects[i]).Elem().Field(k).Name == "IsCurrent" {
					continue
				}

				vi := reflect.ValueOf(t.objects[i]).Elem().Field(k)
				vj := reflect.ValueOf(t.objects[j]).Elem().Field(k)

				si := fmt.Sprint(vi)
				sj := fmt.Sprint(vj)

				if si != sj {
					return si < sj
				}
			}
			return false
		})
	}

	if t.format.IsSerialized() {
		var v any
		if t.isList {
			v = t.objects
			if len(t.objects) == 0 {
				v = []any{}
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
			_, err = writer.Write(pretty.Pretty(out))
			return err
		case YAML:
			out, err := yaml.Marshal(v)
			if err != nil {
				return err
			}
			_, err = writer.Write(out)
			return err
		}
	}

	isEmpty := false
	if t.isList {
		isEmpty = len(t.objects) == 0
	} else if t.isMap() {
		isEmpty = len(t.objects[0].(map[string]string)) == 0
	}
	if isEmpty {
		_, err := fmt.Fprintln(writer, "None found.")
		return err
	}

	w := tablewriter.NewWriter(writer)
	w.SetAutoWrapText(auto)

	if t.isList {
		var header []string
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			tag := strings.Split(reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()), ",")
			if !slices.Contains(tag, "-") {
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
				if !slices.Contains(tag, "-") {
					val := reflect.ValueOf(object).Elem().Field(i)
					row = append(row, getValueAsString(val, tag))
				}
			}
			w.Append(row)
		}
	} else if t.isMap() {
		w.SetAlignment(tablewriter.ALIGN_LEFT)
		for k, v := range t.objects[0].(map[string]string) {
			w.Append([]string{k, v})
		}
	} else {
		w.SetAlignment(tablewriter.ALIGN_LEFT)
		for i := 0; i < reflect.TypeOf(t.objects[0]).Elem().NumField(); i++ {
			tag := strings.Split(reflect.TypeOf(t.objects[0]).Elem().Field(i).Tag.Get(t.format.String()), ",")
			val := reflect.ValueOf(t.objects[0]).Elem().Field(i)
			if !slices.Contains(tag, "-") && !(slices.Contains(tag, "omitempty") && val.IsZero()) {
				w.Append([]string{tag[0], fmt.Sprint(val)})
			}
		}
	}

	w.Render()

	return nil
}

func getValueAsString(val reflect.Value, tag []string) string {
	if slices.Contains(tag, "Current") {
		if val.Bool() {
			return "*"
		} else {
			return " "
		}
	}
	return fmt.Sprint(val)
}

func (t *Table) isMap() bool {
	if len(t.objects) == 0 {
		return false
	}

	_, ok := t.objects[0].(map[string]string)
	return ok
}
