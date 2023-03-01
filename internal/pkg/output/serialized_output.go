package output

import (
	"encoding/json"

	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

// SerializedOutput - pretty prints an object in specified format (JSON or YAML) using tags specified in struct definition
func SerializedOutput(cmd *cobra.Command, v any) error {
	switch GetFormat(cmd) {
	default:
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		utils.Print(string(pretty.Pretty(out)))
	case YAML:
		out, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		utils.Print(string(out))
	}
	return nil
}
