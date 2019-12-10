package kafka

import (
	"context"
	"fmt"
	"os"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	listFields      = []string{"Id", "Name", "ServiceProvider", "Region", "Durability", "Status"}
	listLabels      = []string{"Id", "Name", "Provider", "Region", "Durability", "Status"}
	describeFields  = []string{"Id", "Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "ApiEndpoint", "PricePerHour"}
	describeRenames = map[string]string{"NetworkIngress": "Ingress", "NetworkEgress": "Egress", "ServiceProvider": "Provider"}
)

type clusterCommand struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewClusterCommand(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Kafka clusters.",
		},
		config, prerunner)
	cmd := &clusterCommand{
		CLICommand: cliCmd,
		prerunner: prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommand) init() {
	listCmd := &cobra.Command{
		Use:               "list",
		Short:             "List Kafka clusters.",
		RunE:              c.list,
		Args:              cobra.NoArgs,
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	c.AddCommand(listCmd)
	createCmd := &cobra.Command{
		Use:               "create <name>",
		Short:             "Create a Kafka cluster.",
		RunE:              c.create,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	createCmd.Flags().String("cloud", "", "Cloud provider (e.g. 'aws' or 'gcp')")
	createCmd.Flags().String("region", "", "Cloud region for cluster (e.g. 'us-west-2')")
	createCmd.Flags().Bool("multizone", false, "Use multiple zones for high availability")
	check(createCmd.MarkFlagRequired("cloud"))
	check(createCmd.MarkFlagRequired("region"))
	createCmd.Flags().SortFlags = false
	createCmd.Hidden = true
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a Kafka cluster.",
		RunE:              c.describe,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a Kafka cluster.",
		RunE:              c.update,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	updateCmd.Hidden = true
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Kafka cluster.",
		RunE:              c.delete,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	deleteCmd.Hidden = true
	c.AddCommand(deleteCmd)
	useCmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Make the Kafka cluster active for use in other commands.",
		RunE:              c.use,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: c.prerunner.Authenticated(c.Config, c.CLICommand),
	}
	c.AddCommand(useCmd)
}

func (c *clusterCommand) list(cmd *cobra.Command, args []string) error {
	state, err := c.Config.AuthenticatedState()
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	req := &kafkav1.KafkaCluster{AccountId: state.Auth.Account.Id}
	ctx := c.Config.Context()
	clusters, err := c.Client.Kafka.List(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, cluster := range clusters {
		if cluster.Id == ctx.Kafka {
			cluster.Id = fmt.Sprintf("* %s", cluster.Id)
		} else {
			cluster.Id = fmt.Sprintf("  %s", cluster.Id)
		}
		data = append(data, printer.ToRow(cluster, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	if true {
		return errors.ErrNotImplemented
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	multizone, err := cmd.Flags().GetBool("multizone")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	durability := kafkav1.Durability_LOW
	if multizone {
		durability = kafkav1.Durability_HIGH
	}
	state, err := c.Config.AuthenticatedState()
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cfg := &kafkav1.KafkaClusterConfig{
		AccountId:       state.Auth.Account.Id,
		Name:            args[0],
		ServiceProvider: cloud,
		Region:          region,
		Durability:      durability,
	}
	cluster, err := c.Client.Kafka.Create(context.Background(), cfg)
	if err != nil {
		// TODO: don't swallow validation errors (reportedly separately)
		return errors.HandleCommon(err, cmd)
	}
	return printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	state, err := c.Config.AuthenticatedState()
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	req := &kafkav1.KafkaCluster{AccountId: state.Auth.Account.Id, Id: args[0]}
	cluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	return errors.ErrNotImplemented
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	if true {
		return errors.ErrNotImplemented
	}
	state, err := c.Config.AuthenticatedState()
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	req := &kafkav1.KafkaCluster{AccountId: state.Auth.Account.Id, Id: args[0]}
	err = c.Client.Kafka.Delete(context.Background(), req)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Printf(cmd, "The Kafka cluster %s has been deleted.\n", args[0])
	return nil
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterID := args[0]

	ctx := c.Config.Context()
	if ctx == nil {
		return errors.HandleCommon(errors.ErrNoContext, cmd)
	}
	_, err := ctx.FindKafkaCluster(clusterID, c.Client)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return ctx.SetActiveKafkaCluster(clusterID, c.Client)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
