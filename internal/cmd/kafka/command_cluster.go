package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go"
	"os"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields                = []string{"Id", "Name", "ServiceProvider", "Region", "Durability", "Status"}
	listHumanLabels           = []string{"Id", "Name", "Provider", "Region", "Durability", "Status"}
	listStructuredLabels      = []string{"id", "name", "provider", "region", "durability", "status"}
	describeFields            = []string{"Id", "Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "ApiEndpoint"}
	describeHumanRenames      = map[string]string{"NetworkIngress": "Ingress", "NetworkEgress": "Egress", "ServiceProvider": "Provider"}
	describeStructuredRenames = map[string]string{"Id": "id", "Name": "name", "NetworkIngress": "ingress", "NetworkEgress": "egress", "Storage": "storage",
		"ServiceProvider": "provider", "Region": "region", "Status": "status", "Endpoint": "endpoint", "ApiEndpoint": "api_endpoint"}
)

type clusterCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewClusterCommand(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Kafka clusters.",
		},
		config, prerunner)
	cmd := &clusterCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Kafka cluster.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().String("cloud", "", "Cloud provider ID (e.g. 'aws' or 'gcp').")
	createCmd.Flags().String("region", "", "Cloud region ID for cluster (e.g. 'us-west-2').")
	check(createCmd.MarkFlagRequired("cloud"))
	check(createCmd.MarkFlagRequired("region"))
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka cluster.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Kafka cluster.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Hidden = true
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kafka cluster.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)
	c.AddCommand(&cobra.Command{
		Use:   "use <id>",
		Short: "Make the Kafka cluster active for use in other commands.",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *clusterCommand) list(cmd *cobra.Command, args []string) error {
	req := &kafkav1.KafkaCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.Kafka.List(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, cluster := range clusters {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if cluster.Id == c.Context.KafkaClusterContext.GetActiveKafkaClusterId() {
				cluster.Id = fmt.Sprintf("* %s", cluster.Id)
			} else {
				cluster.Id = fmt.Sprintf("  %s", cluster.Id)
			}
		}
		outputWriter.AddElement(cluster)
	}
	return outputWriter.Out()
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = checkCloudAndRegion(cloud, region, c.Client)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cfg := &kafkav1.KafkaClusterConfig{
		AccountId:       c.EnvironmentId(),
		Name:            args[0],
		ServiceProvider: cloud,
		Region:          region,
		Durability:      kafkav1.Durability_LOW,
		// TODO: remove this once it's no longer required (MCM-130)
		Storage: 5000,
	}
	cluster, err := c.Client.Kafka.Create(context.Background(), cfg)
	if err != nil {
		// TODO: don't swallow validation errors (reportedly separately)
		return errors.HandleCommon(err, cmd)
	}
	return printer.RenderTableOut(cluster, describeFields, describeHumanRenames, os.Stdout)
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &kafkav1.KafkaCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// go-printer has trouble marshaling kafkav1.KafkaCluster struct, creating another struct to fix for now
	type describeStruct struct {
		Id              string
		Name            string
		NetworkIngress  int32
		NetworkEgress   int32
		Storage         int32
		ServiceProvider string
		Region          string
		Status          string
		Endpoint        string
		ApiEndpoint     string
	}
	describeObject := &describeStruct{
		Id:              cluster.Id,
		Name:            cluster.Name,
		NetworkIngress:  cluster.NetworkIngress,
		NetworkEgress:   cluster.NetworkEgress,
		Storage:         cluster.Storage,
		ServiceProvider: cluster.ServiceProvider,
		Region:          cluster.Region,
		Status:          cluster.Status.String(),
		Endpoint:        cluster.Endpoint,
		ApiEndpoint:     cluster.ApiEndpoint,
	}
	return output.DescribeObject(cmd, describeObject, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	return errors.ErrNotImplemented
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &kafkav1.KafkaCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	err := c.Client.Kafka.Delete(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Printf(cmd, "The Kafka cluster %s has been deleted.\n", args[0])
	return nil
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterID := args[0]

	_, err := c.Context.FindKafkaCluster(cmd, clusterID)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return c.Context.SetActiveKafkaCluster(cmd, clusterID)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkCloudAndRegion(cloudId string, regionId string, client *ccloud.Client) error {
	clouds, err := client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return err
	}
	for _, cloud := range clouds {
		if cloudId == cloud.Id {
			for _, region := range cloud.Regions {
				if regionId == region.Id {
					if region.IsSchedulable {
						return nil
					} else {
						break
					}
				}
			}
			return fmt.Errorf("'%s' is not an available region for '%s'. You can view a list of available regions for '%s' with 'kafka region list --cloud %s' command.", regionId, cloudId, cloudId, cloudId)
		}
	}
	return fmt.Errorf("'%s' cloud provider does not exist. You can view a list of available cloud providers and regions with the 'kafka region list' command.", cloudId)
}
