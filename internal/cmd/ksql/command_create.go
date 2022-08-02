package ksql

import (
	"context"
	"fmt"
	"os"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newCreateCommand(isApp bool) *cobra.Command {
	shortText := "Create a ksqlDB cluster."
	var longText string
	runCommand := c.createCluster
	if isApp {
		// DEPRECATED: this should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "DEPRECATED: Create a ksqlDB app."
		longText = "DEPRECATED: Create a ksqlDB app. " + errors.KSQLAppDeprecateWarning
		runCommand = c.createApp
	}

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: shortText,
		Long:  longText,
		Args:  cobra.ExactArgs(1),
		RunE:  runCommand,
	}
	cmd.Flags().String("api-key", "", `DEPRECATED, use credential-identity instead. Kafka API key for the ksqlDB cluster to use (use "confluent api-key create --resource lkc-123456" to create one if none exist).`)
	cmd.Flags().String("api-secret", "", "DEPRECATED, use credential-identity instead. Secret for the Kafka API key.")
	cmd.Flags().String("credential-identity", "", `User account ID or service account ID to be associated with this cluster. We will create an API key associated with this identity and use it to authenticate the ksqlDB cluster with kafka.`)
	cmd.Flags().String("image", "", "Image to run (internal).")
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().Bool("log-exclude-rows", false, "Exclude row data in the processing log.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive("credential-identity",
		"api-key")
	cmd.MarkFlagsMutuallyExclusive("credential-identity",
		"api-secret")

	cmd.MarkFlagsRequiredTogether("api-key", "api-secret")

	_ = cmd.Flags().MarkHidden("image")

	return cmd
}

func (c *ksqlCommand) createApp(cmd *cobra.Command, args []string) error {
	_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	return c.createCluster(cmd, args)
}

func (c *ksqlCommand) createCluster(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	csus, err := cmd.Flags().GetInt32("csu")
	if err != nil {
		return err
	}

	logExcludeRows, err := cmd.Flags().GetBool("log-exclude-rows")
	if err != nil {
		return err
	}

	name := args[0]
	kafkaClusterId := kafkaCluster.ID

	credentialIdentity, err := cmd.Flags().GetString("credential-identity")
	if err != nil {
		return err
	}

	if credentialIdentity != "" {
		return c.createClusterV2(cmd, name, kafkaClusterId, credentialIdentity, csus, logExcludeRows)
	} else {
		return c.createClusterDeprecated(cmd, name, csus, kafkaClusterId, logExcludeRows)
	}

}

func (c *ksqlCommand) createClusterDeprecated(cmd *cobra.Command, name string, csus int32, kafkaClusterId string, logExcludeRows bool) error {
	utils.ErrPrintln(cmd, errors.KSQLApiSecretDeprecateWarning)

	kafkaApiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}

	kafkaApiKeySecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return err
	}

	cfg := &schedv1.KSQLClusterConfig{
		AccountId:             c.EnvironmentId(),
		Name:                  name,
		TotalNumCsu:           uint32(csus),
		KafkaClusterId:        kafkaClusterId,
		DetailedProcessingLog: &types.BoolValue{Value: !logExcludeRows},
		KafkaApiKey: &schedv1.ApiKey{
			Key:    kafkaApiKey,
			Secret: kafkaApiKeySecret,
		},
	}
	cluster, err := c.Client.KSQL.Create(context.Background(), cfg)
	if err != nil {
		return err
	}

	err = c.checkClusterHasEndpoint(cmd, cluster.Endpoint, cluster.Id)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, c.formatClusterForDisplayAndList(c.convertV1ToSchedV2Subset(cluster)), describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *ksqlCommand) createClusterV2(cmd *cobra.Command, name, kafkaClusterId, credentialIdentity string, csus int32, logExcludeRows bool) error {
	if credentialIdentity == "" {
		utils.ErrPrintln(cmd, "You need to provide credential-identity when using the v2 api")
	}
	cluster, err := c.V2Client.CreateKsqlCluster(name, c.EnvironmentId(), kafkaClusterId, credentialIdentity, csus, !logExcludeRows)
	if err != nil {
		return err
	}
	// endpoint value filled later, loop until endpoint information is not null (usually just one describe call is enough)
	endpoint := cluster.Status.GetHttpEndpoint()
	clusterId := *cluster.Id

	err = c.checkClusterHasEndpoint(cmd, endpoint, clusterId)
	if err != nil {
		return err
	}

	//todo bring back formatting
	return output.DescribeObject(cmd, c.formatClusterForDisplayAndList(&cluster), describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *ksqlCommand) checkClusterHasEndpoint(cmd *cobra.Command, endpoint string, clusterId string) error {
	// use count to prevent the command from hanging too long waiting for the endpoint value
	count := 0
	for endpoint == "" && count < 3 {
		res, err := c.V2Client.DescribeKsqlCluster(clusterId, c.EnvironmentId())
		if err != nil {
			return err
		}
		endpoint = res.Status.GetHttpEndpoint()
		count++
	}
	if endpoint == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}
	return nil
}
