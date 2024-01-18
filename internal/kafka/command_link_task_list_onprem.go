package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *linkCommand) newTaskCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manager a cluster link's tasks.",
	}

	cmd.AddCommand(c.newLinkTaskListCommandOnPrem())

	return cmd
}

func (c *linkCommand) newLinkTaskListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <link>",
		Short:             "List a cluster link's tasks.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.taskListOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) taskListOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	getKafkaLinkOpts := kafkarestv3.GetKafkaLinkOpts{
		IncludeTasks: optional.NewBool(true),
	}
	link, httpResp, err := client.ClusterLinkingV3Api.GetKafkaLink(ctx, clusterId, linkName, &getKafkaLinkOpts)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	tasks := make([]task, len(*link.Tasks))
	for i, t := range *link.Tasks {
		errs := make([]taskErr, len(t.Errors))
		for j, e := range t.Errors {
			errs[j] = taskErr{
				ErrorCode:    e.ErrorCode,
				ErrorMessage: e.ErrorMessage,
			}
		}
		tasks[i] = task{
			TaskName: t.TaskName,
			State: t.State,
			Errors:   errs,
		}
	}
	isSerialized := output.GetFormat(cmd).IsSerialized()
	if isSerialized {
		return writeSerialized(cmd, tasks)
	} else {
		return writeHuman(cmd, tasks)
	}
}
