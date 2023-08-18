package kafka

import (
	"fmt"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
)

func (c *brokerCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kafka broker.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *brokerCommand) delete(cmd *cobra.Command, args []string) error {
	brokerIdStr := args[0]
	i, err := strconv.ParseInt(brokerIdStr, 10, 32)
	if err != nil {
		return err
	}
	brokerId := int32(i)

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmYesNoMsg, "broker", brokerIdStr)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	opts := &kafkarestv3.ClustersClusterIdBrokersBrokerIdDeleteOpts{ShouldShutdown: optional.NewBool(true)}
	_, resp, err := restClient.BrokerV3Api.ClustersClusterIdBrokersBrokerIdDelete(restContext, clusterId, brokerId, opts)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	fmt.Printf("Started deletion of broker %d. To monitor the remove-broker task run `confluent kafka broker get-tasks %d --task-type remove-broker`.", brokerId, brokerId)
	return nil
}
