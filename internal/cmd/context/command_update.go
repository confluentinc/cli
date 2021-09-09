package context

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [context]",
		Short: "Update a context field.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
	}

	cmd.Flags().String("name", "", "Set the name of the context.")
	cmd.Flags().String("kafka-cluster", "", "Set the active Kafka cluster for the context.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	ctx, err := c.context(args)
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return err
	}

	if name == "" && kafkaCluster == "" {
		return fmt.Errorf(errors.FlagRequiredErrorMsg, "--name, --kafka-cluster")
	}

	if name != "" {
		delete(ctx.Config.Contexts, ctx.Name)
		delete(ctx.Config.ContextStates, ctx.Name)

		if ctx.Name == ctx.Config.CurrentContext {
			ctx.Config.CurrentContext = name
		}
		ctx.Name = name

		ctx.Config.Contexts[ctx.Name] = ctx.Context
		ctx.Config.ContextStates[ctx.Name] = ctx.Context.State

		if err := ctx.Config.Save(); err != nil {
			return err
		}
	}

	if kafkaCluster != "" {
		if err := ctx.SetActiveKafkaCluster(cmd, kafkaCluster); err != nil {
			return err
		}
	}

	return describeContext(cmd, ctx)
}
