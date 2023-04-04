package kafka

import (
	"context"
	"fmt"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *brokerCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete Kafka brokers.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *brokerCommand) delete(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	brokerIdToNumId, validArgs, err := c.validateArgs(cmd, restClient, restContext, clusterId, args)
	if err != nil {
		return err
	}
	args = validArgs

	if ok, err := form.ConfirmDeletionYesNo(cmd, "broker", args); err != nil || !ok {
		return err
	}

	var errs error
	var deleted []string
	opts := kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{ShouldShutdown: optional.NewBool(true)}
	for _, id := range args {
		if _, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerIdToNumId[id], &opts); err != nil {
			errs = errors.Join(errs, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp))
		} else {
			deleted = append(deleted, id)
		}
	}
	if len(deleted) == 1 {
		output.Printf("Started deletion of broker %[1]s. To monitor the remove-broker task run `confluent kafka broker get-tasks %[1]s --task-type remove-broker`.\n", deleted[0])
	} else if len(deleted) > 1 {
		output.Printf("Started deletion of brokers %s. To monitor a remove-broker task run `confluent kafka broker get-tasks <id> --task-type remove-broker`.\n", utils.ArrayToCommaDelimitedString(deleted, "and"))
	}

	return nil
}

func (c *brokerCommand) validateArgs(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, args []string) (map[string]int32, []string, error) {
	brokerIdToNumId := make(map[string]int32)
	describeFunc := func(id string) error {
		i, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
		numId := int32(i)

		if _, _, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, numId); err != nil {
			return err
		} else {
			brokerIdToNumId[id] = numId
		}
		return nil
	}

	validArgs, err := deletion.ValidateArgsForDeletion(cmd, args, "broker", describeFunc)
	err = errors.NewWrapAdditionalSuggestions(err, fmt.Sprintf(errors.ListResourceSuggestions, "broker", "kafka broker"))

	return brokerIdToNumId, validArgs, err
}
