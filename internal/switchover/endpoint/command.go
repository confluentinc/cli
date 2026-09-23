package endpoint

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

// out is the human-readable shape shared by every endpoint command (list rows
// and create/describe/update tables). Machine-readable output (`-o json` /
// `-o yaml`) is emitted straight from the SDK object so it matches the API
// response verbatim.
type out struct {
	Id                string `human:"ID"`
	DisplayName       string `human:"Display Name"`
	ParentResourceCrn string `human:"Parent Resource CRN"`
	EnvironmentCrn    string `human:"Environment CRN"`
	Target            string `human:"Target,omitempty"`
	Phase             string `human:"Phase"`
	Endpoints         string `human:"Endpoints,omitempty"`
	Conditions        string `human:"Conditions,omitempty"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage switchover endpoints.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func newEndpointOut(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) *out {
	return &out{
		Id:                endpoint.GetId(),
		DisplayName:       endpoint.Spec.GetDisplayName(),
		ParentResourceCrn: endpoint.Spec.GetParentResourceCrn(),
		EnvironmentCrn:    endpoint.Spec.GetEnvironmentCrn(),
		Target:            endpoint.Spec.GetTarget(),
		Phase:             endpoint.Status.GetPhase(),
		Endpoints:         formatEndpoints(endpoint.Spec.GetEndpoints()),
		Conditions:        formatConditions(endpoint.Status.GetConditions()),
	}
}

func formatEndpoints(endpoints []switchoverv1.SwitchoverV1EndpointConfig) string {
	lines := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		filter := endpoint.EndpointFilter
		parts := []string{endpoint.GetName(), filter.GetType()}
		if networkCrn := filter.GetNetworkCrn(); networkCrn != "" {
			parts = append(parts, "network="+networkCrn)
		}
		if accessPointCrn := filter.GetAccessPointCrn(); accessPointCrn != "" {
			parts = append(parts, "access-point="+accessPointCrn)
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

func formatConditions(conditions []switchoverv1.SwitchoverV1SwitchoverEndpointCondition) string {
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
