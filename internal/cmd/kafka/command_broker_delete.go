package kafka

import (
	"context"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *brokerCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more Kafka brokers.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddForceFlag(cmd)

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

	brokerIdToNumId := make(map[string]int32)
	if confirm, err := c.confirmDeletion(cmd, restClient, restContext, clusterId, args, brokerIdToNumId); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	opts := kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{ShouldShutdown: optional.NewBool(true)}
	deleteFunc := func(id string) error {
		if _, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerIdToNumId[id], &opts); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	if len(deleted) == 1 {
		output.Printf("Started deletion of broker %[1]s. To monitor the remove-broker task run `confluent kafka broker get-tasks %[1]s --task-type remove-broker`.\n", deleted[0])
	} else if len(deleted) > 1 {
		output.Printf("Started deletion of brokers %s. To monitor a remove-broker task run `confluent kafka broker get-tasks <id> --task-type remove-broker`.\n", utils.ArrayToCommaDelimitedString(deleted, "and"))
	}

	return err
}

func (c *brokerCommand) confirmDeletion(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, args []string, brokerIdToNumId map[string]int32) (bool, error) {
	describeFunc := func(id string) error {
		i, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
		numId := int32(i)

		if _, _, err =: restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, numId); err != nil {
			return err
		}
		brokerIdToNumId[id] = numId

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, "broker", describeFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString("broker", args))
}
