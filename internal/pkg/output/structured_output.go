package output

import (
	"encoding/json"

	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

// StructuredOutput - pretty prints an object in specified format (JSON or YAML) using tags specified in struct definition
func StructuredOutput(cmd *cobra.Command, obj interface{}) error {
	switch GetFormat(cmd) {
	default:
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		utils.Print(cmd, string(pretty.Pretty(b)))
	case YAML:
		b, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		utils.Print(cmd, string(b))
	}
	return nil
}
