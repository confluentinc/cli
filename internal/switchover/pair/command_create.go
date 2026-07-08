package pair

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <display-name>",
		Short: "Create a switchover pair.",
		Long:  "Create a switchover pair between two Kafka clusters for disaster recovery.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create switchover pair "prod-kafka-dr" between clusters "lkc-111111" (west) and "lkc-222222" (east), active on west.`,
				Code: `confluent switchover pair create prod-kafka-dr --member name=west,id=lkc-111111 --member name=east,id=lkc-222222 --active-member west`,
			},
		),
	}

	cmd.Flags().StringArray("member", nil, `A member of the pair, in the form "name=<name>,id=<cluster-id>". Must be specified exactly twice.`)
	cmd.Flags().String("active-member", "", "The name of the member that starts as active; must match one of the --member names.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("member"))
	cobra.CheckErr(cmd.MarkFlagRequired("active-member"))

	return cmd
}

func parseMemberFlag(raw string) (switchoverv1.SwitchoverV1SwitchoverPairMember, error) {
	member := switchoverv1.SwitchoverV1SwitchoverPairMember{}
	for _, part := range strings.Split(raw, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return member, fmt.Errorf(`invalid --member value %q: expected "name=<name>,id=<cluster-id>"`, raw)
		}
		switch key {
		case "name":
			member.Name = value
		case "id":
			member.MemberId = value
		default:
			return member, fmt.Errorf(`invalid --member key %q: expected "name" or "id"`, key)
		}
	}
	if member.Name == "" || member.MemberId == "" {
		return member, fmt.Errorf(`invalid --member value %q: both "name" and "id" are required`, raw)
	}
	return member, nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	displayName := args[0]

	rawMembers, err := cmd.Flags().GetStringArray("member")
	if err != nil {
		return err
	}
	if len(rawMembers) != 2 {
		return fmt.Errorf(`exactly two --member flags are required, got %d`, len(rawMembers))
	}

	members := make([]switchoverv1.SwitchoverV1SwitchoverPairMember, len(rawMembers))
	for i, raw := range rawMembers {
		member, err := parseMemberFlag(raw)
		if err != nil {
			return err
		}
		members[i] = member
	}

	activeMember, err := cmd.Flags().GetString("active-member")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pair := switchoverv1.SwitchoverV1SwitchoverPair{
		Spec: &switchoverv1.SwitchoverV1SwitchoverPairSpec{
			DisplayName:  switchoverv1.PtrString(displayName),
			Members:      &members,
			ActiveMember: switchoverv1.PtrString(activeMember),
			Environment:  &switchoverv1.EnvScopedObjectReference{Id: environmentId},
		},
	}

	result, err := c.V2Client.CreateSwitchoverPair(pair)
	if err != nil {
		return err
	}

	return printSwitchoverPair(cmd, result)
}

func printSwitchoverPair(cmd *cobra.Command, pair switchoverv1.SwitchoverV1SwitchoverPair) error {
	table := output.NewTable(cmd)
	table.Add(&out{
		Id:           pair.GetId(),
		DisplayName:  pair.Spec.GetDisplayName(),
		ActiveMember: pair.Spec.GetActiveMember(),
		Environment:  pair.Spec.Environment.GetId(),
		Phase:        pair.Status.GetPhase(),
	})
	return table.Print()
}
