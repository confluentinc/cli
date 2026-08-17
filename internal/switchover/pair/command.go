package pair

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

// out is the detailed human-readable shape used by create/describe/update/
// trigger-switch. Machine-readable output (`-o json` / `-o yaml`) is emitted
// straight from the SDK object so it matches the API response verbatim.
type out struct {
	Id           string `human:"ID"`
	DisplayName  string `human:"Display Name"`
	Environment  string `human:"Environment"`
	ActiveMember string `human:"Active Member"`
	FirstActive  string `human:"First Active,omitempty"`
	FailoverType string `human:"Failover Type,omitempty"`
	Phase        string `human:"Phase"`
	Members      string `human:"Members,omitempty"`
	Conditions   string `human:"Conditions,omitempty"`
}

// listOut is the compact per-row shape for `list` (human output only).
type listOut struct {
	Id           string `human:"ID"`
	DisplayName  string `human:"Display Name"`
	Environment  string `human:"Environment"`
	ActiveMember string `human:"Active Member"`
	FailoverType string `human:"Failover Type"`
	Phase        string `human:"Phase"`
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Manage switchover pairs.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newTriggerSwitchCommand())

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

func newPairOut(pair switchoverv1.SwitchoverV1SwitchoverPair) *out {
	return &out{
		Id:           pair.GetId(),
		DisplayName:  pair.Spec.GetDisplayName(),
		Environment:  pair.Spec.GetEnvironmentCrn(),
		ActiveMember: pair.Spec.GetActiveMember(),
		FirstActive:  pair.Spec.GetFirstActive(),
		FailoverType: pair.Spec.GetFailoverType(),
		Phase:        pair.Status.GetPhase(),
		Members:      formatMembers(pair.Spec.GetMembers()),
		Conditions:   formatConditions(pair.Status.GetConditions()),
	}
}

func formatMembers(members []switchoverv1.SwitchoverV1SwitchoverPairMember) string {
	lines := make([]string, len(members))
	for i, member := range members {
		location := ""
		if member.Location != nil {
			location = fmt.Sprintf(", %s/%s", member.Location.GetCloud(), member.Location.GetRegion())
		}
		lines[i] = fmt.Sprintf("%s (%s%s)", member.GetName(), member.GetMemberCrn(), location)
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
