package configuration

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <config-field-1=value-1> ... [config-field-n=value-n]",
		Short: "Set a configuration field's value in ~/.confluent/config.json.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.set,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Disable plugins by setting "disable_plugins" to true`,
				Code: `confluent configuration set disable_plugins=true`,
			},
		),
	}

	return cmd
}

func (c *command) set(cmd *cobra.Command, args []string) error {
	jsonFieldToType, jsonFieldToName := getSettableConfigFields(c.config)
	configMap, err := buildConfigMap(args, jsonFieldToType)
	if err != nil {
		return err
	}

	var updates []string
	for jsonField, val := range configMap {
		oldValue := reflect.ValueOf(c.config).Elem().FieldByName(jsonFieldToName[jsonField])
		switch v := val.(type) {
		case bool:
			oldValue.SetBool(v)
		case string:
			oldValue.SetString(v)
		}
		updates = append(updates, fmt.Sprintf(errors.UpdateSuccessMsg, "value", "config field", jsonField, val))
	}

	if err := c.config.Validate(); err != nil {
		return err
	}
	if err := c.config.Save(); err != nil {
		return err
	}
	for _, update := range updates {
		output.Print(update)
	}
	return nil
}

func buildConfigMap(args []string, settableFields map[string]reflect.Kind) (map[string]any, error) {
	configMap := make(map[string]any, len(args))
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			kv := strings.SplitN(arg, "=", 2)
			if kind, ok := settableFields[kv[0]]; !ok {
				return nil, fmt.Errorf(`config field "%s" either doesn't exist or is not modifiable using this command'`, kv[0])
			} else {
				switch kind {
				case reflect.Bool:
					val, err := strconv.ParseBool(kv[1])
					if err != nil {
						return nil, fmt.Errorf(`"%s" is not a valid value for config field "%s", which is of type: %s`, kv[1], kv[0], kind.String())
					}
					configMap[kv[0]] = val
				case reflect.String:
					configMap[kv[0]] = kv[1]
				}
			}
		} else {
			return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, arg)
		}
	}
	return configMap, nil
}

func getJsonFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}
	if strings.Contains(jsonTag, ",") {
		jsonTag, _, _ = strings.Cut(jsonTag, ",")
	}
	return jsonTag
}

func getSettableConfigFields(cfg *config.Config) (map[string]reflect.Kind, map[string]string) {
	jsonFieldToType := make(map[string]reflect.Kind)
	jsonFieldToName := make(map[string]string)
	t := reflect.TypeOf(*cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		kind := field.Type.Kind()
		switch kind {
		case reflect.Bool, reflect.String:
			jsonFieldName := getJsonFieldName(field)
			if jsonFieldName != "" {
				jsonFieldToType[jsonFieldName] = kind
				jsonFieldToName[jsonFieldName] = field.Name
			}
		}
	}
	return jsonFieldToType, jsonFieldToName
}
