package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newPairListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List switchover pairs.",
		Args:  cobra.NoArgs,
		RunE:  c.pairList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) pairList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pairs, err := c.V2Client.ListSwitchoverPairs(environmentId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, pair := range pairs {
		if pair.Spec == nil {
			continue
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
		list.Add(out)
	}
	return list.Print()
}
