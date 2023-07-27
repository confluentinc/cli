package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <link-1> [link-2] ... [link-n]",
		Short:             "Delete one or more cluster links.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletion(cmd, kafkaREST, cluster.ID, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := kafkaREST.CloudClient.DeleteKafkaLink(cluster.ID, id); err != nil {
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, r)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc)
	resource.PrintDeleteSuccessMsg(deleted, resource.ClusterLink)

	return err
}

func (c *linkCommand) confirmDeletion(cmd *cobra.Command, kafkaREST *pcmd.KafkaREST, clusterId string, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, _, err := kafkaREST.CloudClient.ListKafkaLinkConfigs(clusterId, id)
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
