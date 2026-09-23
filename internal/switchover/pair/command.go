package pair

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

// out is the human-readable shape shared by every pair command (list rows and
// create/describe/update/failover tables). Machine-readable output (`-o json` /
// `-o yaml`) is emitted straight from the SDK object so it matches the API
// response verbatim.
type out struct {
	Id             string `human:"ID"`
	DisplayName    string `human:"Display Name"`
	EnvironmentCrn string `human:"Environment CRN"`
	ActiveMember   string `human:"Active Member"`
	FirstActive    string `human:"First Active,omitempty"`
	FailoverType   string `human:"Failover Type,omitempty"`
	Phase          string `human:"Phase"`
	Members        string `human:"Members,omitempty"`
	Conditions     string `human:"Conditions,omitempty"`
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
	cmd.AddCommand(c.newFailoverCommand())

	return cmd
}

func newPairOut(pair switchoverv1.SwitchoverV1SwitchoverPair) *out {
	return &out{
		Id:             pair.GetId(),
		DisplayName:    pair.Spec.GetDisplayName(),
		EnvironmentCrn: pair.Spec.GetEnvironmentCrn(),
		ActiveMember:   pair.Spec.GetActiveMember(),
		FirstActive:    pair.Spec.GetFirstActive(),
		FailoverType:   pair.Spec.GetFailoverType(),
		Phase:          pair.Status.GetPhase(),
		Members:        formatMembers(pair.Spec.GetMembers()),
		Conditions:     formatConditions(pair.Status.GetConditions()),
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
