package rtce

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *rtceTopicCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <topic-name-1> [topic-name-2] ... [topic-name-n]",
		Short:             "Delete one or more rtce topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *rtceTopicCommand) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	kafkaClusterConfig, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}
	kafkaClusterId := kafkaClusterConfig.GetId()
	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.GetRtceTopic(id, environmentId, kafkaClusterId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, "rtce topic"); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteRtceTopic(id, environmentId, kafkaClusterId)
	}

	deletedIds, err := deletion.DeleteWithoutMessage(cmd, args, deleteFunc)
	deleteMsg := "Requested to delete %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, "rtce topic", fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, plural.Plural("rtce topic"), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return err
}
