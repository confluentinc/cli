package ksql

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newCreateCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: fmt.Sprintf("Create a ksqlDB %s.", resource),
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}
	cmd.Flags().String("credential-identity", "", `User account ID or service account ID to be associated with this cluster. We will create an API key associated with this identity and use it to authenticate the ksqlDB cluster with kafka.`)
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().Bool("log-exclude-rows", false, "Exclude row data in the processing log.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagRequired("credential-identity")

	return cmd
}

func (c *ksqlCommand) create(cmd *cobra.Command, args []string) error {
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

	credentialIdentity, err := cmd.Flags().GetString("credential-identity")
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.CreateKsqlCluster(name, c.EnvironmentId(), kafkaCluster.ID, credentialIdentity, csus, !logExcludeRows)
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

func (c *ksqlCommand) checkClusterHasEndpoint(cmd *cobra.Command, endpoint, clusterId string) error {
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
