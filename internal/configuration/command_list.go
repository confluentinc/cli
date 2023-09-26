package configuration

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user-configurable fields in ~/.confluent/config.json.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	whitelist := getWhitelist(c.cfg)

	list := output.NewList(cmd)
	for field := range whitelist {
		list.Add(c.newFieldOut(field, whitelist))
	}
	return list.Print()
}
