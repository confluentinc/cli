package flink

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

type statementOut struct {
	CreationDate           time.Time         `human:"Creation Date" serialized:"creation_date"`
	Name                   string            `human:"Name" serialized:"name"`
	Statement              string            `human:"Statement" serialized:"statement"`
	ComputePool            string            `human:"Compute Pool" serialized:"compute_pool"`
	Status                 string            `human:"Status" serialized:"status"`
	StatusDetail           string            `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
	LatestOffsets          map[string]string `human:"Latest Offsets" serialized:"latest_offsets"`
	LatestOffsetsTimestamp *time.Time        `human:"Latest Offsets Timestamp" serialized:"latest_offsets_timestamp"`
}

type statementOutOnPrem struct {
	CreationDate string `human:"Creation Date" serialized:"creation_date"`
	Name         string `human:"Name" serialized:"name"`
	Statement    string `human:"Statement" serialized:"statement"`
	ComputePool  string `human:"Compute Pool" serialized:"compute_pool"`
	Status       string `human:"Status" serialized:"status"`
	StatusDetail string `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
	Parallelism  int32  `human:"Parallelism" serialized:"parallelism"`
	Stopped      bool   `human:"Stopped" serialized:"stopped"`
	SqlKind      string `human:"SQL Kind,omitempty" serialized:"sql_kind,omitempty"`
	AppendOnly   bool   `human:"Append Only,omitempty" serialized:"append_only,omitempty"`
	Bounded      bool   `human:"Bounded,omitempty" serialized:"bounded,omitempty"`
}

func (c *command) newStatementCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statement",
		Short: "Manage Flink SQL statements.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newStatementCreateCommand())
		cmd.AddCommand(c.newStatementDeleteCommand())
		cmd.AddCommand(c.newStatementDescribeCommand())
		cmd.AddCommand(c.newStatementListCommand())
		cmd.AddCommand(c.newStatementResumeCommand())
		cmd.AddCommand(c.newStatementStopCommand())
		cmd.AddCommand(c.newStatementUpdateCommand())
	} else {
		cmd.AddCommand(c.newStatementCreateCommandOnPrem())
		cmd.AddCommand(c.newStatementDeleteCommandOnPrem())
		cmd.AddCommand(c.newStatementDescribeCommandOnPrem())
		cmd.AddCommand(c.newStatementListCommandOnPrem())
		cmd.AddCommand(c.newStatementRescaleCommandOnPrem())
		cmd.AddCommand(c.newStatementResumeCommandOnPrem())
		cmd.AddCommand(c.newStatementStopCommandOnPrem())
		cmd.AddCommand(c.newStatementWebUiForwardCommand())
	}
	cmd.AddCommand(c.newStatementExceptionCommand(cfg))

	return cmd
}

func (c *command) validStatementArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validStatementArgsMultiple(cmd, args)
}

func (c *command) validStatementArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

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
