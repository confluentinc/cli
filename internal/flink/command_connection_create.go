package flink

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/flink"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newConnectionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create <name>",
		Short:             "Create a Flink connection.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConnectionArgs), // TODO
		RunE:              c.connectionCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink connection "my-connection" in AWS us-west-2 for OpenAPI with endpoint and API key.`,
				Code: "confluent flink connection create my-connection --cloud aws --region us-west-2 --type openai --endpoint https://api.openai.com/v1/chat/completions --api-key 0000000000000000",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("type", "", fmt.Sprintf("Specify the connection type as %s.", utils.ArrayToCommaDelimitedString(flink.ConnectionTypes, "or")))
	cmd.Flags().String("endpoint", "", "Specify endpoint for the connection.")
	AddConnectionSecretFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("type"))
	cobra.CheckErr(cmd.MarkFlagRequired("endpoint"))

	return cmd
}

func (c *command) connectionCreate(cmd *cobra.Command, args []string) error {
	connectionType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	if err = validateConnectionType(connectionType); err != nil {
		return err
	}

	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if _, err := c.V2Client.GetOrgEnvironment(environmentId); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
	}

	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	connection := flinkgatewayv1.SqlV1Connection{
		Name: flinkgatewayv1.PtrString(args[0]),
		Spec: &flinkgatewayv1.SqlV1ConnectionSpec{
			ConnectionType: flinkgatewayv1.PtrString(strings.ToUpper(connectionType)),
			Endpoint:       flinkgatewayv1.PtrString(endpoint),
			AuthData: &flinkgatewayv1.SqlV1ConnectionSpecAuthDataOneOf{
				SqlV1PlaintextProvider: &flinkgatewayv1.SqlV1PlaintextProvider{
					Kind: lo.ToPtr("PlaintextProvider"),
					Data: lo.ToPtr(base64.StdEncoding.EncodeToString(secretData)),
				},
			},
		},
	}

	connection, err = client.CreateConnection(connection, environmentId, c.Context.LastOrgId)
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
