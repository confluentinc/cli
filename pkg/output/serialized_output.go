package output

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"
)

// SerializedOutput - pretty prints an object in specified format (JSON or YAML) using tags specified in struct definition
func SerializedOutput(cmd *cobra.Command, v any) error {
	switch GetFormat(cmd) {
	default:
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		Print(false, string(pretty.Pretty(out)))
	case YAML:
		out, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		Print(false, string(out))
	}
	return nil
}
