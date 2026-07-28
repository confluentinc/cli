package pair

import (
	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List switchover pairs.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List switchover pairs in the current environment.",
				Code: "confluent switchover pair list",
			},
			examples.Example{
				Text: `List switchover pairs in environment "env-123456".`,
				Code: "confluent switchover pair list --environment env-123456",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pairs, err := c.V2Client.ListSwitchoverPairs(environmentId)
	if err != nil {
		return err
	}

	// Serialized output mirrors the API response verbatim (full spec/status).
	if output.GetFormat(cmd).IsSerialized() {
		if pairs == nil {
			pairs = []switchoverv1.SwitchoverV1SwitchoverPair{}
		}
		return printSerialized(cmd, pairs)
	}

	list := output.NewList(cmd)
	for _, pair := range pairs {
		list.Add(&listOut{
			Id:           pair.GetId(),
			DisplayName:  pair.Spec.GetDisplayName(),
			Environment:  pair.Spec.GetEnvironment(),
			ActiveMember: pair.Spec.GetActiveMember(),
			FailoverType: pair.Spec.GetFailoverType(),
			Phase:        pair.Status.GetPhase(),
		})
	}
	return list.Print()
}
