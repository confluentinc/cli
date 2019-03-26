package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
)

type contextCommand struct {
	*cobra.Command
	config *config.Config
}

// NewContext returns the Cobra contextCommand for `config context`.
func NewContext(config *config.Config) *cobra.Command {
	cmd := &contextCommand{
		Command: &cobra.Command{
			Use:   "context",
			Short: "Manage config contexts.",
		},
		config: config,
	}
	cmd.init()
	return cmd.Command
}

func (c *contextCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all config contexts.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})
	c.AddCommand(&cobra.Command{
		Use:   "use ID",
		Short: "Use a config context.",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "current",
		Short: "Show the current config context.",
		RunE:  c.current,
		Args:  cobra.NoArgs,
	})
	c.AddCommand(&cobra.Command{
		Use:   "get [ID]",
		Short: "Get a config context parameter.",
		RunE:  c.get,
		Args:  cobra.RangeArgs(0, 1),
	})

	setCmd := &cobra.Command{
		Use:   "set [ID]",
		Short: "Set a config context parameter.",
		RunE:  c.set,
		Args:  cobra.RangeArgs(0, 1),
	}
	setCmd.Flags().String("kafka-cluster", "", "Set the current Kafka cluster context")
	c.AddCommand(setCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete a config context.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *contextCommand) list(cmd *cobra.Command, args []string) error {
	type row struct {
		Current    string
		Name       string
		Platform   string
		Credential string
	}
	var data [][]string
	for name, context := range c.config.Contexts {
		current := ""
		if c.config.CurrentContext == name {
			current = "*"
		}
		r := &row{current, name, context.Platform, context.Credential}
		data = append(data, printer.ToRow(r, []string{"Current", "Name", "Platform", "Credential"}))
	}
	printer.RenderCollectionTableOut(data, []string{"Current", "Name", "Platform", "Credential"}, cmd.OutOrStdout())
	return nil
}

func (c *contextCommand) use(cmd *cobra.Command, args []string) error {
	c.config.CurrentContext = args[0]
	return c.config.Save()
}

func (c *contextCommand) current(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), c.config.CurrentContext)
	return nil
}

func (c *contextCommand) get(cmd *cobra.Command, args []string) error {
	context, err := c.context(args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return printer.RenderYAMLOut(context, nil, nil, cmd.OutOrStdout())
}

func (c *contextCommand) set(cmd *cobra.Command, args []string) error {
	context, err := c.context(args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if cmd.Flags().Changed("kafka-cluster") {
		k, err := cmd.Flags().GetString("kafka-cluster")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		context.Kafka = k
	}

	return c.config.Save()
}

func (c *contextCommand) delete(cmd *cobra.Command, args []string) error {
	delete(c.config.Contexts, args[0])
	return c.config.Save()
}

//
// HELPERS
//

func (c *contextCommand) context(args []string) (*config.Context, error) {
	if len(args) == 1 {
		context, ok := c.config.Contexts[args[0]]
		if !ok {
			context = &config.Context{}
			c.config.Contexts[args[0]] = context
			return context, nil
		}
		return context, nil
	}
	return c.config.Context()
}
