package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newConnectionDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [name]",
		Short:             "Describe a Flink connection.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgs), // TODO: update this to connection
		RunE:              c.connectionDescribe,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	connection, err := client.GetConnection(environmentId, args[0], c.Context.GetCurrentOrganization())
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&connectionOut{
		CreationDate: connection.Metadata.GetCreatedAt(),
		Name:         connection.GetName(),
		Type:         connection.Spec.GetConnectionType(),
		Endpoint:     connection.Spec.GetEndpoint(),
		Data:         "<REDACTED>",
		Status:       connection.Status.GetPhase(),
		StatusDetail: connection.Status.GetDetail(),
	})
	return table.Print()
}
