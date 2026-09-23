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

// formatMembers renders one member per block: a name/location line followed by the member CRN on
// its own line. The tables that show it are printed with auto-wrap disabled, so these line breaks
// are kept as-is; with auto-wrap on, tablewriter reflows all whitespace and runs adjacent members
// into one line.
func formatMembers(members []switchoverv1.SwitchoverV1SwitchoverPairMember) string {
	blocks := make([]string, len(members))
	for i, member := range members {
		header := member.GetName()
		if member.Location != nil {
			header = fmt.Sprintf("%s (%s/%s)", member.GetName(), member.Location.GetCloud(), member.Location.GetRegion())
		}
		blocks[i] = header + "\n" + member.GetMemberCrn()
	}
	return strings.Join(blocks, "\n")
}

// formatConditions renders one line per condition. Pair conditions are reported per member (each
// member's Kafka cluster writes its own), so the member name leads the line; otherwise a reader of
// a two-member pair cannot tell which side a condition describes.
func formatConditions(conditions []switchoverv1.SwitchoverV1Condition) string {
	lines := make([]string, len(conditions))
	for i, condition := range conditions {
		line := fmt.Sprintf("%s=%s", condition.GetType(), condition.GetStatus())
		if member := condition.GetMember(); member != "" {
			line = fmt.Sprintf("Member=%s %s", member, line)
		}
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
