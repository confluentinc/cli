package ksql

import (
	"fmt"
	"os"

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

	cmd.Flags().String("credential-identity", "", `user account ID or service account ID to be associated with this cluster. We will create an API key associated with this identity and use it to authenticate the ksqlDB cluster with kafka`)
	cmd.Flags().String("image", "", "Image to run (internal).")
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().Bool("log-exclude-rows", false, "Exclude row data in the processing log.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("credential-identity")
	_ = cmd.Flags().MarkHidden("image")

	return cmd
}

func (c *ksqlCommand) createApp(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
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

	credentialIdentity, err := cmd.Flags().GetString("credential-identity")
	if err != nil {
		return err
	}

	logExcludeRows, err := cmd.Flags().GetBool("log-exclude-rows")
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.CreateKsqlCluster(args[0], c.EnvironmentId(), kafkaCluster.ID, credentialIdentity, csus, logExcludeRows)

	// use count to prevent the command from hanging too long waiting for the endpoint value
	count := 0
	// endpoint value filled later, loop until endpoint information is not null (usually just one describe call is enough)
	for cluster.Status.GetHttpEndpoint() == "" && count < 3 {
		cluster, err = c.V2Client.DescribeKsqlCluster(*cluster.Id, c.EnvironmentId())
		if err != nil {
			return err
		}
		count++
	}

	if cluster.Status.GetHttpEndpoint() == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}

	//todo bring back formatting
	//return output.DescribeObject(cmd, c.updateKsqlClusterForDescribeAndList(cluster), describeFields, describeHumanRenames, describeStructuredRenames)
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}
