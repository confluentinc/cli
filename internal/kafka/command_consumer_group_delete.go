package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *consumerCommand) newGroupDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <group-1> [group-2] ... [group-n]",
		Short:             "Delete one or more empty Kafka consumer groups.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgsMultiple),
		RunE:              c.groupDelete,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the consumer group "my-group". The group must have no active members.`,
				Code: "confluent kafka consumer group delete my-group",
			},
		),
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *consumerCommand) groupDelete(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	// Drop repeated arguments so a group is not deleted twice, which would fail the second request with a 404.
	seen := make(map[string]bool, len(args))
	uniqueArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if !seen[arg] {
			seen[arg] = true
			uniqueArgs = append(uniqueArgs, arg)
		}
	}
	args = uniqueArgs

	// Only a 404 means the group is missing. If the lookup fails for another reason, such as no Describe permission,
	// let the delete request run so the user sees its error instead of a false "not found".
	existenceFunc := func(id string) bool {
		exists, err := kafkaREST.CloudClient.KafkaConsumerGroupExists(id)
		return err != nil || exists
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.ConsumerGroup); err != nil {
		return err
	}

	_, err = deletion.Delete(cmd, args, kafkaREST.CloudClient.DeleteKafkaConsumerGroup, resource.ConsumerGroup)
	return err
}
