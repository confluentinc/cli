package switchover

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newPairCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pair",
		Short: "Manage switchover pairs.",
		Long:  "Manage cluster-level Disaster Recovery switchover pairs for Kafka.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newPairCreateCommand())
	cmd.AddCommand(c.newPairDeleteCommand())
	cmd.AddCommand(c.newPairDescribeCommand())
	cmd.AddCommand(c.newPairFailoverCommand())
	cmd.AddCommand(c.newPairListCommand())
	cmd.AddCommand(c.newPairUpdateCommand())

	return cmd
}

// parseMember parses a `--member` flag value of the form "name=member-id" into
// a SwitchoverPairMember.
func parseMember(raw string) (switchoverv1.SwitchoverV1SwitchoverPairMember, error) {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return switchoverv1.SwitchoverV1SwitchoverPairMember{}, fmt.Errorf(`invalid --member value %q: expected format "name=member-id" (for example, "west=lkc-12345")`, raw)
	}
	return switchoverv1.SwitchoverV1SwitchoverPairMember{
		Name:     strings.TrimSpace(parts[0]),
		MemberId: strings.TrimSpace(parts[1]),
	}, nil
}

func (c *command) validPairArgs(cmd *cobra.Command, args []string) []string {
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
	return autocompletePairs(c.V2Client, environmentId)
}

func autocompletePairs(client *ccloudv2.Client, environmentId string) []string {
	pairs, err := client.ListSwitchoverPairs(environmentId)
	if err != nil {
		return nil
	}
	suggestions := make([]string, len(pairs))
	for i, pair := range pairs {
		suggestions[i] = fmt.Sprintf("%s\t%s", pair.GetId(), pair.Spec.GetDisplayName())
	}
	return suggestions
}

func memberNames(pair switchoverv1.SwitchoverV1SwitchoverPair) []string {
	if pair.Spec == nil {
		return nil
	}
	names := make([]string, 0, len(pair.Spec.GetMembers()))
	for _, member := range pair.Spec.GetMembers() {
		names = append(names, fmt.Sprintf("%s=%s", member.GetName(), member.GetMemberId()))
	}
	return names
}

func printPairTable(cmd *cobra.Command, pair switchoverv1.SwitchoverV1SwitchoverPair) error {
	if pair.Spec == nil {
		return fmt.Errorf("switchover pair response is missing its spec")
	}

	out := &pairOut{
		Id:           pair.GetId(),
		Name:         pair.Spec.GetDisplayName(),
		Environment:  pair.Spec.Environment.GetId(),
		Members:      memberNames(pair),
		ActiveMember: pair.Spec.GetActiveMember(),
	}
	if pair.Status != nil {
		out.Phase = pair.Status.GetPhase()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
