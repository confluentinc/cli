package output

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/pretty"
	"os"
	"reflect"

	"github.com/confluentinc/go-printer"
	"github.com/go-yaml/yaml"
)

const (
	FlagName = "output"
	ShortHandFlag = "o"
	Usage = "Specify the output format."
)

type Format int

// Human enum and string form is not used as we assume human list only when -o flag not used
const (
	Human Format = iota
	JSON
	YAML
)

func (o Format) String() string {
	return [...]string{"human", "json", "yaml"}[o]
}

func NewListOutputWriter(format string, listFields []string, listLabels []string) (ListOutputWriter, error) {
	if len(listLabels) != len(listFields) {
		return nil, fmt.Errorf("selected fields and ouput labels length mismatch")
	}
	if format == JSON.String() {
		return &JSONYAMLListWriter{
			outputFormat: JSON,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	} else if format == YAML.String() {
		return &JSONYAMLListWriter{
			outputFormat: YAML,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	} else if format == "" {
		return &HumanListWriter{
			listFields: listFields,
			listLabels: listLabels,
		}, nil
	}
	return nil, fmt.Errorf("invalid output type")
}

type ListOutputWriter interface {
	AddElement(e interface{})
	Out() error
}

type HumanListWriter struct {
	data       [][]string
	listFields []string
	listLabels []string
}

func (o *HumanListWriter) AddElement(e interface{}) {
	o.data = append(o.data, printer.ToRow(e, o.listFields))
}


func (o *HumanListWriter) Out() error {
	printer.RenderCollectionTable(o.data, o.listLabels)
	return nil
}


type JSONYAMLListWriter struct {
	outputFormat Format
	data         []map[string]string
	listFields   []string
	listLabels   []string
}

func (o *JSONYAMLListWriter) AddElement(e interface{}) {
	elementMap := make(map[string]string)
	c := reflect.ValueOf(e).Elem()
	for i := range o.listFields {
		elementMap[o.listLabels[i]] = fmt.Sprintf("%v", c.FieldByName(o.listFields[i]))
	}
	o.data = append(o.data, elementMap)
}

func (o *JSONYAMLListWriter) Out() error {
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
	_, err = fmt.Fprintf(os.Stdout, string(outputBytes))
	return err
}

