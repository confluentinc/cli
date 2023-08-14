package configuration

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration fields in ~/.confluent/config.json.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	settableFields, _ := getSettableConfigFields(c.config)
	list := output.NewList(cmd)
	t := reflect.TypeOf(*c.config)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := reflect.ValueOf(c.config).Elem().FieldByName(field.Name)
		var values []reflect.Value
		jsonTag := getJsonFieldName(field)
		if jsonTag != "" {
			_, ok := settableFields[jsonTag]
			configOut := &configurationOut{
				Name:     jsonTag,
				Value:    fmt.Sprintf("%v", value),
				Settable: ok,
			}
			if value.Kind() == reflect.Map {
				values = value.MapKeys()
				configOut.Value = fmt.Sprintf("%v", values)
			}
			list.Add(configOut)
		}
	}
	return list.Print()
}
