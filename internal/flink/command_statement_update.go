package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newStatementUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <name>",
		Short:             "Update a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to update the principal of statement "my-statement" to service account "sa-123456".`,
				Code: "confluent flink statement update my-statement --principal sa-123456",
			},
			examples.Example{
				Text: `Request to move "my-statement" to compute pool "lfcp-123456".`,
				Code: "confluent flink statement update my-statement --compute-pool lfcp-123456",
			},
			examples.Example{
				Text: `Request to resume statement "my-statement".`,
				Code: "confluent flink statement update my-statement --stopped=false",
			},
			examples.Example{
				Text: `Request to stop statement "my-statement".`,
				Code: "confluent flink statement update my-statement --stopped=true",
			},
		),
	}

	c.addPrincipalFlag(cmd)
	c.addComputePoolFlag(cmd)
	cmd.Flags().Bool("stopped", false, "Request to stop the statement.")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cmd.MarkFlagsOneRequired("principal", "compute-pool", "stopped")

	return cmd
}

func (c *command) addPrincipalFlag(cmd *cobra.Command) {
	cmd.Flags().String("principal", "", "A user or service account the statement runs as.")

	pcmd.RegisterFlagCompletionFunc(cmd, "principal", func(cmd *cobra.Command, args []string) []string {
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
			suggestions[len(users)+i] = fmt.Sprintf("%s\t%s", serviceAccount.GetId(), serviceAccount.GetDescription())
		}
		return suggestions
	})
}

func (c *command) statementUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(environmentId, args[0], c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return err
	}
	if principal != "" {
		statement.Spec.SetPrincipal(principal)
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	if computePool != "" {
		statement.Spec.SetComputePoolId(computePool)
	}

	if cmd.Flags().Changed("stopped") {
		stopped, err := cmd.Flags().GetBool("stopped")
		if err != nil {
			return err
		}
		statement.Spec.SetStopped(stopped)
	}

	if err := client.UpdateStatement(environmentId, args[0], c.Context.GetCurrentOrganization(), statement); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Requested to update %s \"%s\".\n", resource.FlinkStatement, args[0])
	return nil
}
