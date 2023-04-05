package kafka

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link-1> [link-2] ... [link-n]",
		Short: "Delete cluster links.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteOnPrem,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)
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

	if validArgs, err := c.validateArgsOnPrem(cmd, client, ctx, clusterId, args); err != nil {
		return err
	} else {
		args = validArgs
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.ClusterLink, args[0], args[0]); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.ClusterLink, args); err != nil || !ok {
			return err
		}
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if r, err := client.ClusterLinkingV3Api.DeleteKafkaLink(ctx, clusterId, id, nil); err != nil {
			errs = errors.Join(errs, handleOpenApiError(r, err, client))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ClusterLink)

	return errs
}

func (c *linkCommand) validateArgsOnPrem(cmd *cobra.Command, client *kafkarestv3.APIClient, ctx context.Context, clusterId string, args []string) ([]string, error) {
	describeFunc := func(id string) error {
		if _, _, err := client.ClusterLinkingV3Api.ListKafkaLinkConfigs(ctx, clusterId, id); err != nil {
			return err
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.ClusterLink, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.ClusterLink, "kafka link"))

	return validArgs, err
}
