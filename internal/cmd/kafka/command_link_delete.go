package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link-1> [link-2] ... [link-n]",
		Short: "Delete cluster links.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)
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

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	if validArgs, err := c.validateArgs(cmd, kafkaREST, clusterId, args); err != nil {
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
		if r, err := kafkaREST.CloudClient.DeleteKafkaLink(clusterId, id); err != nil {
			errs = errors.Join(errs, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, r))
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.ClusterLink)

	return errs
}

func (c *linkCommand) validateArgs(cmd *cobra.Command, kafkaREST *pcmd.KafkaREST, clusterId string, args []string) ([]string, error) {
	describeFunc := func(id string) error {
		_, _, err := kafkaREST.CloudClient.ListKafkaLinkConfigs(clusterId, id)
		return err
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, resource.ClusterLink, describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, resource.ClusterLink, "kafka link"))

	return validArgs, err
}
