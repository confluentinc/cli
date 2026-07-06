package output

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"
)

// SerializedOutput - pretty prints an object in specified format (JSON or YAML) using tags specified in struct definition
func SerializedOutput(cmd *cobra.Command, v any) error {
	switch GetFormat(cmd) {
	default:
		out, err := marshalJSON(v)
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

// marshalJSON marshals v to JSON without HTML-escaping <, >, and & (which
// encoding/json does by default), keeping text like SQL statements readable.
func marshalJSON(v any) ([]byte, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	// Encode appends a trailing newline; trim it to match json.Marshal.
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}
