package context

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/kafka"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update [context]",
		Short:             "Update a context field.",
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}

	cmd.Flags().String("name", "", "Set the name of the context.")
	cmd.Flags().String("kafka-cluster", "", "Set the active Kafka cluster for the context.")
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "kafka-cluster")

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

	if name != "" && name != ctx.Name {
		if _, ok := ctx.Config.Contexts[name]; ok {
			return fmt.Errorf(errors.ContextAlreadyExistsErrorMsg, name)
		}

		if err := ctx.Config.ContextStates[ctx.Name].DecryptAuthToken(ctx.Name); err != nil {
			return err
		}

		if err := ctx.Config.ContextStates[ctx.Name].DecryptAuthRefreshToken(ctx.Name); err != nil {
			return err
		}

		delete(ctx.Config.Contexts, ctx.Name)
		delete(ctx.Config.ContextStates, ctx.Name)

		if ctx.Name == ctx.Config.CurrentContext {
			ctx.Config.CurrentContext = name
		}
		ctx.Name = name

		ctx.Config.Contexts[ctx.Name] = ctx
		ctx.Config.ContextStates[ctx.Name] = ctx.State

		if err := ctx.Config.Save(); err != nil {
			return err
		}
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return err
	}

	if kafkaCluster != "" {
		if _, err := kafka.FindCluster(nil, ctx, kafkaCluster); err != nil {
			return err
		}

		ctx.KafkaClusterContext.SetActiveKafkaCluster(kafkaCluster)
		if err := ctx.Save(); err != nil {
			return err
		}
	}

	return describeContext(cmd, ctx)
}
