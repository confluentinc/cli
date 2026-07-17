package endpoint

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <display-name>",
		Short: "Create a switchover endpoint.",
		Long:  "Create a switchover endpoint bound to a switchover pair.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create switchover endpoint "prod-kafka-dr-endpoint" for switchover pair "sw-123456".`,
				Code: `confluent switchover endpoint create prod-kafka-dr-endpoint --switchover-pair sw-123456 --endpoint name=west-platt,type=private --endpoint name=east-platt,type=private`,
			},
		),
	}

	cmd.Flags().String("switchover-pair", "", "The ID of the switchover pair this endpoint is bound to.")
	cmd.Flags().StringArray("endpoint", nil, `An endpoint side, in the form "name=<name>,type=<private|public>[,network=<network-id>][,access-point=<access-point-id>]". Must be specified exactly twice.`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("switchover-pair"))
	cobra.CheckErr(cmd.MarkFlagRequired("endpoint"))

	return cmd
}

func parseEndpointFlag(raw string) (switchoverv1.SwitchoverV1EndpointConfig, error) {
	config := switchoverv1.SwitchoverV1EndpointConfig{}
	filter := switchoverv1.SwitchoverV1EndpointFilter{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return config, fmt.Errorf(`invalid --endpoint value %q: expected "key=value" pairs`, raw)
		}
		switch key {
		case "name":
			config.Name = value
		case "type":
			filter.Type = value
		case "network":
			filter.NetworkId = switchoverv1.PtrString(value)
		case "access-point":
			filter.AccessPoint = switchoverv1.PtrString(value)
		default:
			return config, fmt.Errorf(`invalid --endpoint key %q`, key)
		}
	}
	if config.Name == "" || filter.Type == "" {
		return config, fmt.Errorf(`invalid --endpoint value %q: "name" and "type" are required`, raw)
	}
	config.EndpointFilter = filter
	return config, nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	displayName := args[0]

	switchoverPairId, err := cmd.Flags().GetString("switchover-pair")
	if err != nil {
		return err
	}

	rawEndpoints, err := cmd.Flags().GetStringArray("endpoint")
	if err != nil {
		return err
	}
	if len(rawEndpoints) != 2 {
		return fmt.Errorf(`exactly two --endpoint flags are required, got %d`, len(rawEndpoints))
	}

	endpoints := make([]switchoverv1.SwitchoverV1EndpointConfig, len(rawEndpoints))
	for i, raw := range rawEndpoints {
		endpointConfig, err := parseEndpointFlag(raw)
		if err != nil {
			return err
		}
		endpoints[i] = endpointConfig
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	endpoint := switchoverv1.SwitchoverV1SwitchoverEndpoint{
		Spec: &switchoverv1.SwitchoverV1SwitchoverEndpointSpec{
			DisplayName:      switchoverv1.PtrString(displayName),
			Endpoints:        &endpoints,
			Environment:      switchoverv1.PtrString(environmentId),
			SwitchoverPairId: switchoverv1.PtrString(switchoverPairId),
		},
	}

	result, err := c.V2Client.CreateSwitchoverEndpoint(endpoint)
	if err != nil {
		return err
	}

	return printSwitchoverEndpoint(cmd, result)
}

func printSwitchoverEndpoint(cmd *cobra.Command, endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) error {
	table := output.NewTable(cmd)
	table.Add(&out{
		Id:             endpoint.GetId(),
		DisplayName:    endpoint.Spec.GetDisplayName(),
		SwitchoverPair: endpoint.Spec.GetSwitchoverPairId(),
		Environment:    endpoint.Spec.GetEnvironment(),
		Phase:          endpoint.Status.GetPhase(),
	})
	return table.Print()
}
