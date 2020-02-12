package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	humanString   = "human"
	jsonString    = "json"
	yamlString    = "yaml"
	FlagName      = "output"
	ShortHandFlag = "o"
	Usage         = `Specify the output format as "human", "json" or "yaml".`
	DefaultValue  = humanString
)

var (
	InvalidFormatError = fmt.Errorf("invalid output format type")
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

type ListOutputWriter interface {
	AddElement(e interface{})
	Out()   error
	GetOutputFormat() Format
	StableSort()
}

func NewListOutputWriter(cmd *cobra.Command, listFields []string, humanLabels []string, structuredLabels []string) (ListOutputWriter, error) {
	return NewListOutputCustomizableWriter(cmd, listFields, humanLabels, structuredLabels, os.Stdout)
}

func NewListOutputCustomizableWriter(cmd *cobra.Command, listFields []string, humanLabels []string, structuredLabels []string, writer io.Writer) (ListOutputWriter, error) {
	format, err := cmd.Flags().GetString(FlagName)
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	if format == JSON.String() {
		return &StructuredListWriter{
			outputFormat: JSON,
			listFields:   listFields,
			listLabels:   structuredLabels,
			writer:       writer,
		}, nil
	} else if format == YAML.String() {
		return &StructuredListWriter{
			outputFormat: YAML,
			listFields:   listFields,
			listLabels:   structuredLabels,
			writer:       writer,
		}, nil
	} else if format == Human.String() {
		return &HumanListWriter{
			outputFormat: Human,
			listFields:   listFields,
			listLabels:   humanLabels,
			writer:       writer,
		}, nil
	}
	return nil, InvalidFormatError
}

func StructuredOutput(format string, obj interface{}) error {
	var b []byte
	if format == JSON.String() {
		j, _ := json.Marshal(obj)
		b = pretty.Pretty(j)
	} else if format == YAML.String() {
		b, _ = yaml.Marshal(obj)
	} else {
		return InvalidFormatError
	}
	_, err := fmt.Fprintf(os.Stdout, string(b))
	return err
}
