package flink

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type statementCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// printStatementWarnings renders warnings below the table, on stderr so that stdout stays the
// command's data. Serialized output already carries them in the warnings field.
func printStatementWarnings(cmd *cobra.Command, warnings []types.StatementWarning) {
	if output.GetFormat(cmd) != output.Human {
		return
	}

	if block := types.FormatStatementWarnings(warnings); block != "" {
		output.ErrPrintln(false, "")
		output.ErrPrintln(false, block)
		output.ErrPrintln(false, "")
	}
}

type statementOut struct {
	CreationDate           time.Time                `human:"Creation Date" serialized:"creation_date"`
	Name                   string                   `human:"Name" serialized:"name"`
	Statement              string                   `human:"Statement" serialized:"statement"`
	ComputePool            string                   `human:"Compute Pool,omitempty" serialized:"compute_pool,omitempty"`
	Status                 string                   `human:"Status" serialized:"status"`
	StatusDetail           string                   `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
	Warnings               []types.StatementWarning `human:"-" serialized:"warnings,omitempty"`
	LatestOffsets          map[string]string        `human:"Latest Offsets" serialized:"latest_offsets"`
	LatestOffsetsTimestamp *time.Time               `human:"Latest Offsets Timestamp" serialized:"latest_offsets_timestamp"`
}

func newStatementCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command { //nolint:unparam
	cmd := &cobra.Command{
		Use:         "statement",
		Short:       "Manage Flink SQL statements in Confluent Cloud.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &statementCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
	}

	cmd.AddCommand(
		c.newCreateCommand(),
		c.newDeleteCommand(),
		c.newDescribeCommand(),
		c.newStatementExceptionCommand(),
		c.newListCommand(),
		c.newStatementResumeCommand(),
		c.newStatementStopCommand(),
		c.newUpdateCommand(),
	)

	return cmd
}

func (c *statementCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *statementCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteStatements()
}

func (c *statementCommand) autocompleteStatements() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return nil
	}

	statements, err := client.ListStatements(environmentId, c.Context.GetCurrentOrganization(), c.Context.GetCurrentFlinkComputePool())
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(statements))
	for i, statement := range statements {
		suggestions[i] = statement.GetName()
	}
	return suggestions
}

// addDatabaseFlag mirrors the shared (*command).addDatabaseFlag; statementCommand embeds
// *pcmd.AuthenticatedCLICommand rather than *command, so it carries its own copy.
func (c *statementCommand) addDatabaseFlag(cmd *cobra.Command) {
	cmd.Flags().String("database", "", "The database which will be used as the default database. When using Kafka, this is the cluster ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "database", c.autocompleteDatabases)
}

func (c *statementCommand) autocompleteDatabases(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}
