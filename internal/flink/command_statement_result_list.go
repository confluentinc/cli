package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/retry"
)

func (c *command) newStatementResultListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <statement-name>",
		Short:             "List results for a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementResultList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	cmd.Flags().Bool("wait", false, "Block until the statement is no longer pending before fetching results.")
	cmd.Flags().Int("max-rows", 100, "Maximum number of result rows to fetch.")

	return cmd
}

func (c *command) statementResultList(cmd *cobra.Command, args []string) error {
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

	phase := statement.Status.GetPhase()

	if phase == "FAILED" {
		return fmt.Errorf("statement %q has failed: %s", args[0], statement.Status.GetDetail())
	}

	if phase == "PENDING" {
		wait, err := cmd.Flags().GetBool("wait")
		if err != nil {
			return err
		}
		if !wait {
			return fmt.Errorf("statement %q is still pending, use --wait to wait for it to complete", args[0])
		}

		err = retry.Retry(time.Second, 2*time.Minute, func() error {
			statement, err = client.GetStatement(environmentId, args[0], c.Context.GetCurrentOrganization())
			if err != nil {
				return err
			}
			if statement.Status.GetPhase() == "PENDING" {
				return fmt.Errorf(`statement phase is "%s"`, statement.Status.GetPhase())
			}
			return nil
		})
		if err != nil {
			return err
		}

		if statement.Status.GetPhase() == "FAILED" {
			return fmt.Errorf("statement %q has failed: %s", args[0], statement.Status.GetDetail())
		}
	}

	traits := statement.Status.GetTraits()
	schema := traits.GetSchema()
	columns := schema.GetColumns()
	if len(columns) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Statement has no results to display.")
		return nil
	}

	maxRows, err := cmd.Flags().GetInt("max-rows")
	if err != nil {
		return err
	}

	statementResults, err := fetchAllResults(client, environmentId, args[0], c.Context.GetCurrentOrganization(), schema, maxRows)
	if err != nil {
		return err
	}

	if err := printStatementResults(cmd, statementResults); err != nil {
		return err
	}

	if maxRows > 0 && len(statementResults.Rows) >= maxRows {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: results truncated at %d rows. Use --max-rows to adjust the limit.\n", maxRows)
	}

	return nil
}
