package flink

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type describeStatementOut struct {
	CreationDate time.Time         `human:"Creation Date" serialized:"creation_date"`
	Name         string            `human:"Name" serialized:"name"`
	Statement    string            `human:"Statement" serialized:"statement"`
	ComputePool  string            `human:"Compute Pool" serialized:"compute_pool"`
	Status       string            `human:"Status" serialized:"status"`
	StatusDetail string            `human:"Status Detail,omitempty" serialized:"status_detail,omitempty"`
	Properties   map[string]string `human:"Properties" serialized:"properties"`
	Principal    string            `human:"Principal" serialized:"principal"`
}

func (c *command) newStatementDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <name>",
		Short:             "Describe a Flink SQL statement.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validStatementArgs),
		RunE:              c.statementDescribe,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementDescribe(cmd *cobra.Command, args []string) error {
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

	table := output.NewTable(cmd)
	table.Add(&describeStatementOut{
		CreationDate: statement.Metadata.GetCreatedAt(),
		Name:         statement.GetName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
		Properties:   statement.Spec.GetProperties(),
		Principal:    statement.Spec.GetPrincipal(),
	})
	return table.Print()
}
