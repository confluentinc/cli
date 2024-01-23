package kafka

import (
	"fmt"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/utils"
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
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	brokerIdToIntId, err := mapBrokerIdToIntId(args)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, _, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerIdToIntId[id])
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Broker); err != nil {
		return err
	}

	opts := &kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{ShouldShutdown: optional.NewBool(true)}
	deleteFunc := func(id string) error {
		if _, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerIdToIntId[id], opts); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return nil
	}

	deletedIds, err := deletion.DeleteWithoutMessage(args, deleteFunc)
	deleteMsg := "Started deletion of %s %s. To monitor a remove-broker task run `confluent kafka broker get-tasks <id> --task-type remove-broker`.\n"
	if len(deletedIds) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.Broker, fmt.Sprintf("\"%s\"", deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.Plural(resource.Broker), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return err
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
