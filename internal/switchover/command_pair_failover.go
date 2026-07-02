package switchover

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var failoverTypes = []string{"CLEAN", "UNCLEAN", "RESTORE"}

func (c *command) newPairFailoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "failover <id>",
		Short:             "Trigger a failover on a switchover pair.",
		Long:              "Trigger a failover (or switchback) on a switchover pair, promoting the passive member to active.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPairArgs),
		RunE:              c.pairFailover,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Trigger a clean failover of switchover pair "sw-123456", promoting member "east".`,
				Code: "confluent switchover pair failover sw-123456 --member east --type CLEAN",
			},
		),
	}

	cmd.Flags().String("member", "", "Name of the member to promote to active. If omitted, the other member is promoted.")
	cmd.Flags().String("type", "", fmt.Sprintf("Specify the failover type as %s.", utils.ArrayToCommaDelimitedString(failoverTypes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return failoverTypes })
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) pairFailover(cmd *cobra.Command, args []string) error {
	request := switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest{}

	if cmd.Flags().Changed("member") {
		member, err := cmd.Flags().GetString("member")
		if err != nil {
			return err
		}
		request.SetActiveMember(member)
	}

	if cmd.Flags().Changed("type") {
		failoverType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}
		request.SetFailoverType(strings.ToUpper(failoverType))
	}

	pair, err := c.V2Client.FailoverSwitchoverPair(args[0], request)
	if err != nil {
		return err
	}

	return printPairTable(cmd, pair)
}
