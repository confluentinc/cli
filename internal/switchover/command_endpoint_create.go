package switchover

import (
	"fmt"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a switchover endpoint.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.endpointCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a switchover endpoint "prod-kafka-dr-endpoint" bound to switchover pair "sw-123456".`,
				Code: "confluent switchover endpoint create prod-kafka-dr-endpoint --switchover-pair sw-123456 " +
					"--endpoint name=west-platt,resource-id=lkc-111111,type=PRIVATE,cloud=AWS,region=us-west-2,gateway=n-ccn-west-1 " +
					"--endpoint name=east-platt,resource-id=lkc-222222,type=PRIVATE,cloud=AWS,region=us-east-1,gateway=n-ccn-east-1",
			},
		),
	}

	cmd.Flags().String("switchover-pair", "", "ID of the switchover pair this endpoint is bound to.")
	cmd.Flags().StringArray("endpoint", nil, `An endpoint definition as comma-separated "key=value" fields (keys: name, resource-id, type, cloud, region, gateway, access-point). Specify exactly twice.`)
	cmd.Flags().String("target", "", "Name of the endpoint that should start as active.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("switchover-pair"))
	cobra.CheckErr(cmd.MarkFlagRequired("endpoint"))

	return cmd
}

func (c *command) endpointCreate(cmd *cobra.Command, args []string) error {
	displayName := args[0]

	switchoverPairId, err := cmd.Flags().GetString("switchover-pair")
	if err != nil {
		return err
	}

	endpointFlags, err := cmd.Flags().GetStringArray("endpoint")
	if err != nil {
		return err
	}
	if len(endpointFlags) != 2 {
		return fmt.Errorf("exactly two `--endpoint` flags must be specified, but received %d", len(endpointFlags))
	}

	endpoints := make([]switchoverv1.SwitchoverV1EndpointConfig, len(endpointFlags))
	for i, raw := range endpointFlags {
		config, err := parseEndpoint(raw)
		if err != nil {
			return err
		}
		endpoints[i] = config
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	spec := &switchoverv1.SwitchoverV1SwitchoverEndpointSpec{
		DisplayName:    switchoverv1.PtrString(displayName),
		Endpoints:      &endpoints,
		Environment:    &switchoverv1.EnvScopedObjectReference{Id: environmentId},
		SwitchoverPair: &switchoverv1.EnvScopedObjectReference{Id: switchoverPairId},
	}

	if cmd.Flags().Changed("target") {
		target, err := cmd.Flags().GetString("target")
		if err != nil {
			return err
		}
		spec.SetTarget(target)
	}

	endpoint, err := c.V2Client.CreateSwitchoverEndpoint(switchoverv1.SwitchoverV1SwitchoverEndpoint{Spec: spec})
	if err != nil {
		return err
	}

	return printEndpointTable(cmd, endpoint)
}
