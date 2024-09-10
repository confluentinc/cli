package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/flink"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
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
	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
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

	connections, err := client.ListConnections(environmentId, c.Context.GetCurrentOrganization(), strings.ToUpper(connectionType))
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, connection := range connections {
		list.Add(&connectionOut{
			CreationDate: connection.Metadata.GetCreatedAt(),
			Name:         connection.GetName(),
			Type:         connection.Spec.GetConnectionType(),
			Endpoint:     connection.Spec.GetEndpoint(),
			Data:         connection.Spec.AuthData.SqlV1PlaintextProvider.GetData(),
			Status:       connection.Status.GetPhase(),
			StatusDetail: connection.Status.GetDetail(),
		})
	}
	return list.Print()
}
