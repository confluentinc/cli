package output

import "github.com/spf13/cobra"

type Format int

const (
	Human Format = iota
	JSON
	YAML
)

const FlagName = "output"

var ValidFlagValues = []string{"human", "json", "yaml"}

func GetFormat(cmd *cobra.Command) Format {
	format, _ := cmd.Flags().GetString(FlagName)

	switch format {
	default:
		return Human
	case "json":
		return JSON
	case "yaml":
		return YAML
	}
}

func (o Format) String() string {
	return ValidFlagValues[o]
}

func (o Format) IsSerialized() bool {
	return o == JSON || o == YAML
}
