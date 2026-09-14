package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newConnectionDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <name>",
		Short:             "Describe a Flink connection.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgs),
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

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
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
		Environment:  environmentId,
		Cloud:        c.Context.GetCurrentFlinkCloudProvider(),
		Region:       c.Context.GetCurrentFlinkRegion(),
		Type:         connection.Spec.GetConnectionType(),
		Endpoint:     connection.Spec.GetEndpoint(),
		Data:         connection.Spec.AuthData.SqlV1PlaintextProvider.GetData(),
		Status:       connection.Status.GetPhase(),
		StatusDetail: connection.Status.GetDetail(),
	})
	return table.Print()
}
