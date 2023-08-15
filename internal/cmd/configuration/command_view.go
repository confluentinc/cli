package configuration

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View user-configurable fields in ~/.confluent/config.json.",
		Args:  cobra.NoArgs,
		RunE:  c.view,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) view(cmd *cobra.Command, _ []string) error {
	list := output.NewList(cmd)
	for field := range c.jsonFieldToName {
		value := reflect.ValueOf(c.cfg).Elem().FieldByName(c.jsonFieldToName[field])
		configOut := &configurationOut{
			Name:  field,
			Value: fmt.Sprintf("%v", value),
		}
		list.Add(configOut)
	}
	return list.Print()
}
