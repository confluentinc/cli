package switchover

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newPairCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a switchover pair.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.pairCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a switchover pair "prod-kafka-dr" between two Kafka clusters, with "west" active.`,
				Code: "confluent switchover pair create prod-kafka-dr --member west=lkc-111111 --member east=lkc-222222 --active-member west",
			},
		),
	}

	cmd.Flags().StringArray("member", nil, `A member of the pair in the format "name=member-id" (for example, "west=lkc-12345"). Specify exactly twice.`)
	cmd.Flags().String("active-member", "", "Name of the member that should start as active.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("member"))
	cobra.CheckErr(cmd.MarkFlagRequired("active-member"))

	return cmd
}

func (c *command) pairCreate(cmd *cobra.Command, args []string) error {
	displayName := args[0]

	memberFlags, err := cmd.Flags().GetStringArray("member")
	if err != nil {
		return err
	}
	if len(memberFlags) != 2 {
		return fmt.Errorf("exactly two `--member` flags must be specified, but received %d", len(memberFlags))
	}

	members := make([]switchoverv1.SwitchoverV1SwitchoverPairMember, len(memberFlags))
	for i, raw := range memberFlags {
		member, err := parseMember(raw)
		if err != nil {
			return err
		}
		members[i] = member
	}

	activeMember, err := cmd.Flags().GetString("active-member")
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(members, func(m switchoverv1.SwitchoverV1SwitchoverPairMember) bool { return m.GetName() == activeMember }) {
		return fmt.Errorf("`--active-member` %q must match one of the member names", activeMember)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createPair := switchoverv1.SwitchoverV1SwitchoverPair{
		Spec: &switchoverv1.SwitchoverV1SwitchoverPairSpec{
			DisplayName:  switchoverv1.PtrString(displayName),
			Members:      &members,
			ActiveMember: switchoverv1.PtrString(activeMember),
			Environment:  &switchoverv1.EnvScopedObjectReference{Id: environmentId},
		},
	}

	pair, err := c.V2Client.CreateSwitchoverPair(createPair)
	if err != nil {
		return err
	}

	return printPairTable(cmd, pair)
}
