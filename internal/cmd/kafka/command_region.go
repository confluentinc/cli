package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/go-printer"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	listFields      = []string{"Id", "Name", "ServiceProvider", "Region", "Durability", "Status"}
	listLabels      = []string{"Id", "Name", "Provider", "Region", "Durability", "Status"}
	describeFields  = []string{"Id", "Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "ApiEndpoint"}
	describeRenames = map[string]string{"NetworkIngress": "Ingress", "NetworkEgress": "Egress", "ServiceProvider": "Provider"}
)

type regionCommand struct {
	*cobra.Command
	config *config.Config
	client ccloud.Kafka
	ch     *pcmd.ConfigHelper
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewRegionCommand(config *config.Config, client ccloud.Kafka, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &clusterCommand{
		Command: &cobra.Command{
			Use:   "region",
			Short: "Kafka cloud provider and regions.",
		},
		config: config,
		client: client,
		ch:     ch,
	}
	cmd.init()
	return cmd.Command
}

func (c *regionCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})
}

func (c *regionCommand) list(cmd *cobra.Command, args []string) error {
	environment, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	req := &kafkav1.KafkaCluster{AccountId: environment}
	clusters, err := c.client.List(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	currCtx, err := c.config.Context()
	if err != nil && err != errors.ErrNoContext {
		return err
	}
	var data [][]string
	for _, cluster := range clusters {
		if cluster.Id == currCtx.Kafka {
			cluster.Id = fmt.Sprintf("* %s", cluster.Id)
		} else {
			cluster.Id = fmt.Sprintf("  %s", cluster.Id)
		}
		data = append(data, printer.ToRow(cluster, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}
