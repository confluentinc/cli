package switchover

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage switchover endpoints.",
		Long:  "Manage endpoint-level Disaster Recovery switchover endpoints for Kafka.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newEndpointActivateCommand())
	cmd.AddCommand(c.newEndpointCreateCommand())
	cmd.AddCommand(c.newEndpointDeleteCommand())
	cmd.AddCommand(c.newEndpointDescribeCommand())
	cmd.AddCommand(c.newEndpointListCommand())
	cmd.AddCommand(c.newEndpointUpdateCommand())

	return cmd
}

// parseEndpoint parses an `--endpoint` flag value into a SwitchoverV1EndpointConfig.
// Format: comma-separated key=value pairs, for example:
//
//	name=west-platt,resource-id=lkc-west,type=PRIVATE,cloud=AWS,region=us-west-2,gateway=n-ccn-west-1
//
// `name`, `resource-id`, and `type` are required; `cloud`, `region`, `gateway`,
// and `access-point` are optional.
func parseEndpoint(raw string) (switchoverv1.SwitchoverV1EndpointConfig, error) {
	fields := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		if strings.TrimSpace(pair) == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return switchoverv1.SwitchoverV1EndpointConfig{}, fmt.Errorf(`invalid --endpoint value %q: each field must be "key=value"`, raw)
		}
		fields[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	name := fields["name"]
	resourceId := fields["resource-id"]
	endpointType := strings.ToUpper(fields["type"])
	if name == "" || resourceId == "" || endpointType == "" {
		return switchoverv1.SwitchoverV1EndpointConfig{}, fmt.Errorf(`invalid --endpoint value %q: "name", "resource-id", and "type" are required`, raw)
	}

	filter := switchoverv1.SwitchoverV1EndpointFilter{
		ResourceId: resourceId,
		Type:       endpointType,
	}
	if v, ok := fields["cloud"]; ok {
		filter.SetCloud(strings.ToUpper(v))
	}
	if v, ok := fields["region"]; ok {
		filter.SetRegion(v)
	}
	if v, ok := fields["gateway"]; ok {
		filter.SetGateway(v)
	}
	if v, ok := fields["access-point"]; ok {
		filter.SetAccessPoint(v)
	}

	return switchoverv1.SwitchoverV1EndpointConfig{Name: name, EndpointFilter: filter}, nil
}

func (c *command) validEndpointArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}
	return autocompleteEndpoints(c.V2Client, environmentId)
}

func autocompleteEndpoints(client *ccloudv2.Client, environmentId string) []string {
	endpoints, err := client.ListSwitchoverEndpoints(environmentId)
	if err != nil {
		return nil
	}
	suggestions := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", endpoint.GetId(), endpoint.Spec.GetDisplayName())
	}
	return suggestions
}

func endpointNames(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) []string {
	if endpoint.Spec == nil {
		return nil
	}
	names := make([]string, 0, len(endpoint.Spec.GetEndpoints()))
	for _, config := range endpoint.Spec.GetEndpoints() {
		names = append(names, config.GetName())
	}
	return names
}

func printEndpointTable(cmd *cobra.Command, endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) error {
	if endpoint.Spec == nil {
		return fmt.Errorf("switchover endpoint response is missing its spec")
	}

	out := &endpointOut{
		Id:             endpoint.GetId(),
		Name:           endpoint.Spec.GetDisplayName(),
		Environment:    endpoint.Spec.Environment.GetId(),
		SwitchoverPair: endpoint.Spec.SwitchoverPair.GetId(),
		Endpoints:      endpointNames(endpoint),
		Target:         endpoint.Spec.GetTarget(),
		DrEndpoint:     endpoint.Spec.GetDrEndpoint(),
	}
	if endpoint.Status != nil {
		out.Phase = endpoint.Status.GetPhase()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
