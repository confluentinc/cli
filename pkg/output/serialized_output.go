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

// SerializedOutputFromJsonTags pretty prints an object in the specified format (JSON or YAML) using
// the `json` tags in the struct definition for both formats. Use it for SDK response objects, which
// carry `json` tags only: yaml.Marshal ignores `json` tags, so the YAML path is routed through the
// JSON representation to keep the API field names.
func SerializedOutputFromJsonTags(cmd *cobra.Command, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if GetFormat(cmd) == YAML {
		var generic any
		if err := json.Unmarshal(data, &generic); err != nil {
			return err
		}
		out, err := yaml.Marshal(generic)
		if err != nil {
			return err
		}
		Print(false, string(out))
		return nil
	}

	Print(false, string(pretty.Pretty(data)))
	return nil
}

// JsonValue is a value's JSON encoding, printed under its JSON key names in both `-o json` and
// `-o yaml`. A table row carries an SDK object as a JsonValue because only the row's own tags are
// rewritten per format: yaml.Marshal ignores a nested SDK struct's `json` tags and would print its
// Go field names. The row printer also rejects interface-typed fields, so the field can't be `any`.
type JsonValue json.RawMessage

// FromJsonTags encodes v as a JsonValue. A value that can't be encoded prints as null.
func FromJsonTags(v any) JsonValue {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}

func (v JsonValue) MarshalJSON() ([]byte, error) {
	if len(v) == 0 {
		return []byte("null"), nil
	}
	return v, nil
}

func (v JsonValue) MarshalYAML() (any, error) {
	if len(v) == 0 {
		return nil, nil
	}
	var generic any
	if err := json.Unmarshal(v, &generic); err != nil {
		return nil, err
	}
	return generic, nil
}
