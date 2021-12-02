package ksql

import (
	"context"
	"fmt"
	"os"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *appCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a ksqlDB app.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
	}

	cmd.Flags().String("api-key", "", "Kafka API key for the ksqlDB cluster to use.")
	cmd.Flags().String("api-secret", "", "Secret for the Kafka API key.")
	cmd.Flags().String("image", "", "Image to run (internal).")
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

	_ = cmd.MarkFlagRequired("api-key")
	_ = cmd.MarkFlagRequired("api-secret")
	_ = cmd.Flags().MarkHidden("image")

	return cmd
}

func (c *appCommand) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
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

	if kafkaApiKey != "" && kafkaApiKeySecret != "" {
		cfg.KafkaApiKey = &schedv1.ApiKey{
			Key:    kafkaApiKey,
			Secret: kafkaApiKeySecret,
		}
	} else {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLCreateDeprecateWarning)
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
		count += 1
	}

	if cluster.Endpoint == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, cluster.Id)
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}
