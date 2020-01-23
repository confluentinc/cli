package output

import (
	"fmt"
	"github.com/confluentinc/go-printer"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
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

type ListOutputWriter interface {
	AddElement(e interface{})
	Out()   error
	GetOutputFormat() Format
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
		return &StructuredListWriter{
			outputFormat: JSON,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	} else if format == YAML.String() {
		return &StructuredListWriter{
			outputFormat: YAML,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	} else if format == Human.String() {
		return &HumanListWriter{
			outputFormat: Human,
			listFields:   listFields,
			listLabels:   listLabels,
		}, nil
	}
	return nil, fmt.Errorf(InvalidFormatError)
}

func DescribeObject(cmd *cobra.Command, obj interface{}, fields []string, humanRenames, structuredRenames map[string]string) error {
	format, err := cmd.Flags().GetString(FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if !(format == Human.String() || format == JSON.String() || format == YAML.String()) {
		return fmt.Errorf(InvalidFormatError)
	}
	return printer.RenderOut(obj, fields, humanRenames, structuredRenames, format, os.Stdout)
}

