package flink

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newConnectionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update [name]",
		Short:             "Update a Flink connection. Only secret can be updated.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgs),
		RunE:              c.connectionUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update API key of Flink connection "my-connection".`,
				Code: `confluent flink connection update my-connection --cloud aws --region us-west-2 --api-key new-key`,
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	AddConnectionSecretFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) connectionUpdate(cmd *cobra.Command, args []string) error {
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

	connectionType := strings.ToLower(connection.Spec.GetConnectionType())

	if err = validateConnectionType(connectionType); err != nil {
		return err
	}

	secretMap, err := validateConnectionSecrets(cmd, connectionType)
	if err != nil {
		return err
	}

	secretData, err := json.Marshal(secretMap)
	if err != nil {
		return err
	}

	newConnection := flinkgatewayv1.SqlV1Connection{
		Name: flinkgatewayv1.PtrString(args[0]),
		Spec: &flinkgatewayv1.SqlV1ConnectionSpec{
			ConnectionType: flinkgatewayv1.PtrString(strings.ToUpper(connectionType)),
			Endpoint:       flinkgatewayv1.PtrString(connection.Spec.GetEndpoint()),
			AuthData: &flinkgatewayv1.SqlV1ConnectionSpecAuthDataOneOf{
				SqlV1PlaintextProvider: &flinkgatewayv1.SqlV1PlaintextProvider{
					Kind: flinkgatewayv1.PtrString("PlaintextProvider"),
					Data: flinkgatewayv1.PtrString(string(secretData[:])),
				},
			},
		},
	}

	if err := client.UpdateConnection(environmentId, args[0], c.Context.GetCurrentOrganization(), newConnection); err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&connectionOut{
		CreationDate: connection.Metadata.GetCreatedAt(),
		Name:         connection.GetName(),
		Type:         connection.Spec.GetConnectionType(),
		Endpoint:     connection.Spec.GetEndpoint(),
		Data:         connection.Spec.AuthData.SqlV1PlaintextProvider.GetData(),
		Status:       connection.Status.GetPhase(),
		StatusDetail: connection.Status.GetDetail(),
	})
	return table.Print()
}
