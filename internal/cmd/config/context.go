package config

import (
	"sort"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type contextCommand struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
	analytics analytics.Client
}

// NewContext returns the Cobra contextCommand for `config context`.
func NewContext(config *v3.Config, prerunner pcmd.PreRunner, analytics analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "context",
			Short: "Manage config contexts.",
		},
		config, prerunner)
	cmd := &contextCommand{
		CLICommand: cliCmd,
		prerunner:  prerunner,
		analytics:  analytics,
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			c.analytics.SetCommandType(analytics.ContextUse)
			return c.prerunner.Anonymous(c.CLICommand)(cmd, args)
		},
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
	var contextNames []string
	for name := range c.Config.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)
	var data [][]string
	for _, name := range contextNames {
		current := ""
		if c.Config.CurrentContext == name {
			current = "*"
		}
		context := c.Config.Contexts[name]
		r := &row{current, name, context.PlatformName, context.CredentialName}
		data = append(data, printer.ToRow(r, []string{"Current", "Name", "Platform", "Credential"}))
	}
	printer.RenderCollectionTableOut(data, []string{"Current", "Name", "Platform", "Credential"}, cmd.OutOrStdout())
	return nil
}

func (c *contextCommand) use(cmd *cobra.Command, args []string) error {
	name := args[0]
	err := c.Config.SetContext(name)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *contextCommand) current(cmd *cobra.Command, args []string) error {
	pcmd.Println(cmd, c.Config.CurrentContext)
	return nil
}

func (c *contextCommand) get(cmd *cobra.Command, args []string) error {
	context, err := c.context(cmd, args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return printer.RenderYAMLOut(context, nil, nil, cmd.OutOrStdout())
}

func (c *contextCommand) set(cmd *cobra.Command, args []string) error {
	context, err := c.context(cmd, args)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if cmd.Flags().Changed("kafka-cluster") {
		clusterId, err := cmd.Flags().GetString("kafka-cluster")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return context.SetActiveKafkaCluster(cmd, clusterId)
	}
	return nil
}

func (c *contextCommand) delete(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	err := c.Config.DeleteContext(contextName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return c.Config.Save()
}

func (c *contextCommand) context(cmd *cobra.Command, args []string) (*pcmd.DynamicContext, error) {
	var context *pcmd.DynamicContext
	var err error
	if len(args) == 1 {
		contextName := args[0]
		context, err = c.Config.FindContext(contextName)
	} else {
		context, err = c.Config.Context(cmd)
		if context == nil {
			err = errors.ErrNoContext
		}
	}
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	return context, nil
}
