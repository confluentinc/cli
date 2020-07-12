package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/confluentinc/go-printer"
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
	Usage         = `Specify the output format as "human", "json", or "yaml".`
	DefaultValue  = humanString
)

var (
	allFormatStrings = []string{humanString, jsonString, yamlString}
)

type Format int

const (
	Human Format = iota
	JSON
	YAML
)

func (o Format) String() string {
	return allFormatStrings[o]
}

type ListOutputWriter interface {
	AddElement(e interface{})
	Out() error
	GetOutputFormat() Format
	StableSort()
}

func NewListOutputWriter(cmd *cobra.Command, listFields []string, humanLabels []string, structuredLabels []string) (ListOutputWriter, error) {
	return NewListOutputCustomizableWriter(cmd, listFields, humanLabels, structuredLabels, os.Stdout)
}

func NewListOutputCustomizableWriter(cmd *cobra.Command, listFields []string, humanLabels []string, structuredLabels []string, writer io.Writer) (ListOutputWriter, error) {
	format, err := cmd.Flags().GetString(FlagName)
	if err != nil {
		return nil, err
	}
	switch format {
	case JSON.String():
		return &StructuredListWriter{
			outputFormat: JSON,
			listFields:   listFields,
			listLabels:   structuredLabels,
			writer:       writer,
		}, nil
	case YAML.String():
		return &StructuredListWriter{
			outputFormat: YAML,
			listFields:   listFields,
			listLabels:   structuredLabels,
			writer:       writer,
		}, nil
	case Human.String():
		return &HumanListWriter{
			outputFormat: Human,
			listFields:   listFields,
			listLabels:   humanLabels,
			writer:       writer,
		}, nil
	default:
		return nil, newInvalidOutputFormatFlagError(format)
	}
}

func DescribeObject(cmd *cobra.Command, obj interface{}, fields []string, humanRenames, structuredRenames map[string]string) error {
	format, err := cmd.Flags().GetString(FlagName)
	if err != nil {
		return err
	}
	if !(format == Human.String() || format == JSON.String() || format == YAML.String()) {
		return newInvalidOutputFormatFlagError(format)
	}
	return printer.RenderOut(obj, fields, humanRenames, structuredRenames, format, os.Stdout)
}

func StructuredOutput(format string, obj interface{}) error {
	var b []byte
	if format == JSON.String() {
		j, _ := json.Marshal(obj)
		b = pretty.Pretty(j)
	} else if format == YAML.String() {
		b, _ = yaml.Marshal(obj)
	} else {
		return newInvalidOutputFormatFlagError(format)
	}
	_, err := fmt.Fprintf(os.Stdout, string(b))
	return err
}

func newInvalidOutputFormatFlagError(format string) error {
	errorMsg := fmt.Sprintf(errors.InvalidFlagValueErrorMsg, format, FlagName)
	suggestionsMsg := fmt.Sprintf(errors.InvalidFlagValueSuggestions, FlagName, strings.Join(allFormatStrings, ", "))
	return errors.NewErrorWithSuggestions(errorMsg, suggestionsMsg)
}
