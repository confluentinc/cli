package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/go-yaml/yaml"
	"github.com/tidwall/pretty"
)

type StructuredListWriter struct {
	outputFormat output
	data         []map[string]string
	listFields   []string
	listLabels   []string
	writer       io.Writer
}

func (o *StructuredListWriter) AddElement(e interface{}) {
	elementMap := make(map[string]string)
	c := reflect.ValueOf(e).Elem()
	for i := range o.listFields {
		elementMap[o.listLabels[i]] = fmt.Sprintf("%v", c.FieldByName(o.listFields[i]))
	}
	o.data = append(o.data, elementMap)
}

func (o *StructuredListWriter) Out() error {
	var outputBytes []byte
	var err error
	if o.outputFormat == YAML {
		outputBytes, err = yaml.Marshal(o.data)
	} else {
		outputBytes, err = json.Marshal(o.data)
		outputBytes = pretty.Pretty(outputBytes)
	}
	if err != nil {
		return err
	}
	if len(o.data) != 0 {
		_, err = fmt.Fprint(o.writer, string(outputBytes))
	} else {
		if o.outputFormat == JSON {
			_, err = fmt.Fprintln(o.writer, "[]")
		} else {
			_, err = fmt.Fprintln(o.writer, "")
		}
	}
	return err
}

func (o *StructuredListWriter) GetOutputFormat() output {
	return o.outputFormat
}

func (o *StructuredListWriter) StableSort() {
	sort.Slice(o.data, func(i, j int) bool {
		// use listLabels because map iteration order not guaranteed
		for _, x := range o.listLabels {
			if o.data[i][x] != o.data[j][x] {
				return o.data[i][x] < o.data[j][x]
			}
		}
		return false
	})
}
