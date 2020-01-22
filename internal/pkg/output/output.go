package output

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"os"
	"reflect"

	"github.com/confluentinc/go-printer"
	"github.com/go-yaml/yaml"
)

const (
	humanString        = "human"
	jsonString         = "json"
	yamlString         = "yaml"
	FlagName           = "output"
	ShortHandFlag      = "o"
	Usage              = `Specify the output format as "human", "json" or "yaml".`
	DefaultValue       = humanString
	InvalidFormatError = "invalid output format type"
)

type Format int

const (
	Human Format = iota
	JSON
	YAML
)

func (o Format) String() string {
	return [...]string{humanString, jsonString, yamlString}[o]
}

func NewListOutputWriter(cmd *cobra.Command, listFields []string, listLabels []string) (ListOutputWriter, error) {
	format, err := cmd.Flags().GetString(FlagName)
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
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
	} else if format == Human.String() {
		return &TableListWriter{
			outputFormat: Human,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	}
	return nil, fmt.Errorf(InvalidFormatError)
}

type ListOutputWriter interface {
	AddElement(e interface{})
	Out()   error
	GetOutputFormat() Format
}

type TableListWriter struct {
	outputFormat Format
	data         [][]string
	listFields   []string
	listLabels   []string
}

func (o *TableListWriter) AddElement(e interface{}) {
	o.data = append(o.data, printer.ToRow(e, o.listFields))
}


func (o *TableListWriter) Out() error {
	printer.RenderCollectionTable(o.data, o.listLabels)
	return nil
}

func (o *TableListWriter) GetOutputFormat() Format {
	return o.outputFormat
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

func (o *JSONYAMLListWriter) GetOutputFormat() Format {
	return o.outputFormat
}

