package kafka

import (
	"context"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/resource"
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

	brokerIdToIntId, err := mapBrokerIdToIntId(args)
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletion(cmd, restClient, restContext, clusterId, args, brokerIdToIntId); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	opts := &kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{ShouldShutdown: optional.NewBool(true)}
	deleteFunc := func(id string) error {
		if _, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerIdToIntId[id], opts); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return nil
	}

	singleDeleteMsg := "Started deletion of broker %s. To monitor the remove-broker task run `confluent kafka broker get-tasks <id> --task-type remove-broker`.\n"
	multipleDeleteMsg := "Started deletion of brokers %s. To monitor a remove-broker task run `confluent kafka broker get-tasks <id> --task-type remove-broker`.\n"
	_, err = resource.DeleteWithCustomMessage(args, deleteFunc, singleDeleteMsg, multipleDeleteMsg)
	return err
}

func (c *brokerCommand) confirmDeletion(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, args []string, brokerIdToIntId map[string]int32) (bool, error) {
	existenceFunc := func(id string) bool {
		_, _, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerIdToIntId[id])
		return err == nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Broker, existenceFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Broker, args))
}

func mapBrokerIdToIntId(args []string) (map[string]int32, error) {
	brokerIdToIntId := make(map[string]int32)
	for _, arg := range args {
		i, err := strconv.ParseInt(arg, 10, 32)
		if err != nil {
			return nil, err
		}
		brokerIdToIntId[arg] = int32(i)
	}

	return brokerIdToIntId, nil
}
