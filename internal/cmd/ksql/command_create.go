package ksql

import (
	"context"
	"fmt"
	"os"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newCreateCommand(isApp bool) *cobra.Command {
	shortText := "Create a ksqlDB cluster."
	runCommand := c.createCluster
	if isApp {
		// DEPRECATED: this line should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "Create a ksqlDB app. " + errors.KSQLAppDeprecateWarning
		runCommand = c.createApp
	}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: shortText,
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(runCommand),
	}

	cmd.Flags().String("api-key", "", "Kafka API key for the ksqlDB cluster to use.")
	cmd.Flags().String("api-secret", "", "Secret for the Kafka API key.")
	cmd.Flags().String("image", "", "Image to run (internal).")
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("api-key")
	_ = cmd.MarkFlagRequired("api-secret")
	_ = cmd.Flags().MarkHidden("image")

	return cmd
}

func (c *ksqlCommand) createCluster(cmd *cobra.Command, args []string) error {
	return c.create(cmd, args, false)
}

func (c *ksqlCommand) createApp(cmd *cobra.Command, args []string) error {
	return c.create(cmd, args, true)
}

func (c *ksqlCommand) create(cmd *cobra.Command, args []string, isApp bool) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	csus, err := cmd.Flags().GetInt32("csu")
	if err != nil {
		return err
	}

	cfg := &schedv1.KSQLClusterConfig{
		AccountId:      c.EnvironmentId(),
		Name:           args[0],
		TotalNumCsu:    uint32(csus),
		KafkaClusterId: kafkaCluster.ID,
	}

	kafkaApiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}

	kafkaApiKeySecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return err
	}

	cfg.KafkaApiKey = &schedv1.ApiKey{
		Key:    kafkaApiKey,
		Secret: kafkaApiKeySecret,
	}

	image, err := cmd.Flags().GetString("image")
	if err == nil && len(image) > 0 {
		cfg.Image = image
	}

	cluster, err := c.Client.KSQL.Create(context.Background(), cfg)
	if err != nil {
		return err
	}

	// use count to prevent the command from hanging too long waiting for the endpoint value
	count := 0
	// endpoint value filled later, loop until endpoint information is not null (usually just one describe call is enough)
	for cluster.Endpoint == "" && count < 3 {
		req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: cluster.Id}
		cluster, err = c.Client.KSQL.Describe(context.Background(), req)
		if err != nil {
			return err
		}
		count++
	}

	if cluster.Endpoint == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}

	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}
