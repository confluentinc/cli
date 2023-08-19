package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

	existenceFunc := func(id string) bool {
		_, _, err := client.ClusterLinkingV3Api.ListKafkaLinkConfigs(ctx, clusterId, id)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.ClusterLink, args[0]); err != nil {
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

	_, err = deletion.Delete(args, deleteFunc, resource.ClusterLink)
	return err
}
