package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link-1> [link-2] ... [link-n]",
		Short: "Delete one or more cluster links.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteOnPrem,
	}

	pcmd.AddForceFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *linkCommand) deleteOnPrem(cmd *cobra.Command, args []string) error {
	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletionOnPrem(cmd, client, ctx, clusterId, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := client.ClusterLinkingV3Api.DeleteKafkaLink(ctx, clusterId, id, nil); err != nil {
			return handleOpenApiError(r, err, client)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc)
	resource.PrintDeleteSuccessMsg(deleted, resource.ClusterLink)

	return err
}

func (c *linkCommand) confirmDeletionOnPrem(cmd *cobra.Command, client *kafkarestv3.APIClient, ctx context.Context, clusterId string, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, _, err := client.ClusterLinkingV3Api.ListKafkaLinkConfigs(ctx, clusterId, id)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.ClusterLink, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.ClusterLink, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.ClusterLink, args[0], args[0]), args[0]); err != nil {
		return false, err
	}

	return true, nil
}
