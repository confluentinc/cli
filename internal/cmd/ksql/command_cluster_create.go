package ksql

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a ksqlDB cluster.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	cmd.Flags().String("credential-identity", "", `User account ID or service account ID to be associated with this cluster. We will create an API key associated with this identity and use it to authenticate the ksqlDB cluster with Kafka.`)
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().Bool("log-exclude-rows", false, "Exclude row data in the processing log.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("credential-identity")

	return cmd
}

func (c *ksqlCommand) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	csu, err := cmd.Flags().GetInt32("csu")
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

	cluster, err := c.V2Client.CreateKsqlCluster(name, c.EnvironmentId(cmd), kafkaCluster.ID, credentialIdentity, csu, !logExcludeRows)
	if err != nil {
		return err
	}
	// endpoint value filled later, loop until endpoint information is not null (usually just one describe call is enough)
	endpoint := cluster.Status.GetHttpEndpoint()

	log.CliLogger.Trace("Polling ksqlDB cluster")

	ticker := time.NewTicker(100 * time.Millisecond)
	// use count to prevent the command from hanging too long waiting for the endpoint value
	for count := 0; endpoint == "" && count < 3; count++ {
		if count != 0 {
			<-ticker.C
		}
		res, err := c.V2Client.DescribeKsqlCluster(cluster.GetId(), c.EnvironmentId(cmd))
		if err != nil {
			return err
		}
		endpoint = res.Status.GetHttpEndpoint()
	}
	ticker.Stop()
	if endpoint == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}

	srCluster, _ := c.Context.FetchSchemaRegistryByEnvironmentId(context.Background(), c.EnvironmentId(cmd))
	if srCluster != nil {
		utils.ErrPrintln(cmd, errors.SchemaRegistryRoleBindingRequiredForKsqlWarning)
	}

	table := output.NewTable(cmd)
	table.Add(c.formatClusterForDisplayAndList(&cluster))
	return table.Print()
}
