package endpoint

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

// out is the detailed human-readable shape used by create/describe/update.
// Machine-readable output (`-o json` / `-o yaml`) is emitted straight from the
// SDK object so it matches the API response verbatim.
type out struct {
	Id             string `human:"ID"`
	DisplayName    string `human:"Display Name"`
	SwitchoverPair string `human:"Switchover Pair"`
	Environment    string `human:"Environment"`
	Target         string `human:"Target,omitempty"`
	Phase          string `human:"Phase"`
	Endpoints      string `human:"Endpoints,omitempty"`
	Conditions     string `human:"Conditions,omitempty"`
}

// listOut is the compact per-row shape for `list` (human output only).
type listOut struct {
	Id             string `human:"ID"`
	DisplayName    string `human:"Display Name"`
	SwitchoverPair string `human:"Switchover Pair"`
	Environment    string `human:"Environment"`
	Phase          string `human:"Phase"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage switchover endpoints.",
		Long:  "Manage switchover endpoints. This API is not yet implemented on the backend; commands will fail against a live Confluent Cloud environment.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

// printSerialized emits v as JSON or YAML using the SDK object's `json` tags,
// so the output matches the API response verbatim. (output.SerializedOutput's
// YAML path uses yaml.v3, which ignores `json` tags; route YAML through the
// JSON representation to preserve the API field names.)
func printSerialized(cmd *cobra.Command, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.YAML {
		var generic any
		if err := json.Unmarshal(data, &generic); err != nil {
			return err
		}
		out, err := yaml.Marshal(generic)
		if err != nil {
			return err
		}
		output.Print(false, string(out))
		return nil
	}

	output.Print(false, string(pretty.Pretty(data)))
	return nil
}

func newEndpointOut(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) *out {
	return &out{
		Id:             endpoint.GetId(),
		DisplayName:    endpoint.Spec.GetDisplayName(),
		SwitchoverPair: endpoint.Spec.GetParentResourceCrn(),
		Environment:    endpoint.Spec.GetEnvironmentCrn(),
		Target:         endpoint.Spec.GetTarget(),
		Phase:          endpoint.Status.GetPhase(),
		Endpoints:      formatEndpoints(endpoint.Spec.GetEndpoints()),
		Conditions:     formatConditions(endpoint.Status.GetConditions()),
	}
}

func formatEndpoints(endpoints []switchoverv1.SwitchoverV1EndpointConfig) string {
	lines := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		filter := endpoint.EndpointFilter
		parts := []string{endpoint.GetName(), filter.GetType()}
		if networkId := filter.GetNetworkCrn(); networkId != "" {
			parts = append(parts, "network="+networkId)
		}
		if accessPoint := filter.GetAccessPointCrn(); accessPoint != "" {
			parts = append(parts, "access-point="+accessPoint)
		}
		if hostname := endpoint.GetHostname(); hostname != "" {
			parts = append(parts, "hostname="+hostname)
		}
		if cloud, region := endpoint.GetCloud(), endpoint.GetRegion(); cloud != "" || region != "" {
			parts = append(parts, strings.TrimPrefix(cloud+"/"+region, "/"))
		}
		if connectionType := endpoint.GetConnectionType(); connectionType != "" {
			parts = append(parts, connectionType)
		}
		lines[i] = strings.Join(parts, " ")
	}
	return strings.Join(lines, "\n")
}

func formatConditions(conditions []switchoverv1.SwitchoverV1Condition) string {
	lines := make([]string, len(conditions))
	for i, condition := range conditions {
		line := fmt.Sprintf("%s=%s", condition.GetType(), condition.GetStatus())
		if reason := condition.GetReason(); reason != "" {
			line += fmt.Sprintf(" (%s)", reason)
		}
		if message := condition.GetMessage(); message != "" {
			line += ": " + message
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
