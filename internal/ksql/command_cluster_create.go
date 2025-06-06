package ksql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *ksqlCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a ksqlDB cluster.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create ksqlDB cluster "my-cluster" associated with user "u-123456".`,
				Code: "confluent ksql cluster create my-cluster --credential-identity u-123456",
			},
		),
	}

	c.addCredentialIdentityFlag(cmd)
	cmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	cmd.Flags().Bool("log-exclude-rows", false, "Exclude row data in the processing log.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("credential-identity"))

	return cmd
}

func (c *ksqlCommand) addCredentialIdentityFlag(cmd *cobra.Command) {
	cmd.Flags().String("credential-identity", "", "User account ID or service account ID to be associated with this cluster. An API key associated with this identity will be created and used to authenticate the ksqlDB cluster with Kafka.")
	pcmd.RegisterFlagCompletionFunc(cmd, "credential-identity", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		users, err := c.V2Client.ListIamUsers()
		if err != nil {
			return nil
		}

		serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(users)+len(serviceAccounts))

		for i, user := range users {
			suggestions[i] = fmt.Sprintf("%s\t%s", user.GetId(), user.GetFullName())
		}

		for i, serviceAccount := range serviceAccounts {
			description := fmt.Sprintf("%s: %s", serviceAccount.GetDisplayName(), serviceAccount.GetDescription())
			suggestions[len(users)+i] = fmt.Sprintf("%s\t%s", serviceAccount.GetId(), description)
		}

		return suggestions
	})
}

func (c *ksqlCommand) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.CreateKsqlCluster(name, environmentId, kafkaCluster.ID, credentialIdentity, csu, !logExcludeRows)
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
		res, err := c.V2Client.DescribeKsqlCluster(cluster.GetId(), environmentId)
		if err != nil {
			return err
		}
		endpoint = res.Status.GetHttpEndpoint()
	}
	ticker.Stop()
	if endpoint == "" {
		output.ErrPrintln(c.Config.EnableColor, "Endpoint not yet populated. To obtain the endpoint, use `confluent ksql cluster describe`.")
	}

	if clusters, _ := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId); len(clusters) > 0 {
		if _, ok := clusters[0].GetIdOk(); ok {
			output.ErrPrintln(c.Config.EnableColor, "[WARN] Confirm that the users or service accounts that will interact with this cluster have the required privileges to access Schema Registry.")
		}
	}

	table := output.NewTable(cmd)
	table.Add(c.formatClusterForDisplayAndList(&cluster))
	return table.Print()
}
