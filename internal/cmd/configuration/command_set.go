package configuration

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

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
	keys, values, err := buildUpdates(args, jsonFieldToType)
	if err != nil {
		return err
	}

	var updates []string

	for i := range keys {
		oldValue := reflect.ValueOf(c.config).Elem().FieldByName(jsonFieldToName[keys[i]])
		switch v := values[i].(type) {
		case bool:
			oldValue.SetBool(v)
		case string:
			oldValue.SetString(v)
		}
		updates = append(updates, fmt.Sprintf(errors.UpdateSuccessMsg, "value", "config field", keys[i], values[i]))
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

func buildUpdates(args []string, settableFields map[string]reflect.Kind) ([]string, []any, error) {
	keys := make([]string, len(args))
	values := make([]any, len(args))
	index := 0
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			kv := strings.SplitN(arg, "=", 2)
			if kind, ok := settableFields[kv[0]]; !ok {
				return nil, nil, fmt.Errorf(`config field "%s" either doesn't exist or is not modifiable using this command'`, kv[0])
			} else {
				switch kind {
				case reflect.Bool:
					val, err := strconv.ParseBool(kv[1])
					if err != nil {
						return nil, nil, fmt.Errorf(`"%s" is not a valid value for config field "%s", which is of type: %s`, kv[1], kv[0], kind.String())
					}
					values[index] = val
				case reflect.String:
					values[index] = kv[1]
				}
				keys[index] = kv[0]
				index++
			}
		} else {
			return nil, nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, arg)
		}
	}
	return keys, values, nil
}
