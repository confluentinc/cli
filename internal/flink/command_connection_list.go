package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newConnectionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink connections.",
		Args:  cobra.NoArgs,
		RunE:  c.connectionList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("type", "", fmt.Sprintf("Specify the connection type as %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionTypes, "or")))
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(envNotFoundErrorMsg, environmentId))
	}

	connectionType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	if connectionType != "" {
		if err = validateConnectionType(connectionType); err != nil {
			return err
		}
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	connections, err := client.ListConnections(environmentId, c.Context.GetCurrentOrganization(), strings.ToUpper(connectionType))
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, connection := range connections {
		list.Add(&connectionOut{
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
	}
	return list.Print()
}
