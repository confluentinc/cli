package pair

import (
	"fmt"

	"github.com/spf13/cobra"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newTriggerSwitchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-switch <id>",
		Short: "Trigger a failover or switchback on a switchover pair.",
		Long:  "Trigger a failover (or switchback) on a switchover pair. This redirects live traffic between the pair's members.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.triggerSwitch,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Fail switchover pair "sw-123456" over to its "east" member.`,
				Code: `confluent switchover pair trigger-switch sw-123456 --active-member east`,
			},
		),
	}

	cmd.Flags().String("active-member", "", "The name of the member to promote to active. If omitted, the other member is promoted.")
	cmd.Flags().String("failover-type", "PLANNED", "The failover semantics to apply: PLANNED, UNPLANNED, or RESTORE.")
	cmd.Flags().Bool("force", false, "Skip the confirmation prompt.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) triggerSwitch(cmd *cobra.Command, args []string) error {
	id := args[0]

	activeMember, err := cmd.Flags().GetString("active-member")
	if err != nil {
		return err
	}

	failoverType, err := cmd.Flags().GetString("failover-type")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// The :failover body carries the environment as a CRN (ORC-9794), unlike other operations
	// which take it as a query parameter.
	environmentCrn := fmt.Sprintf("crn://confluent.cloud/organization=%s/environment=%s", c.Context.GetCurrentOrganization(), environmentId)

	promptMsg := fmt.Sprintf(`This triggers a %s failover on switchover pair "%s", redirecting live traffic between regions. Do you want to proceed?`, failoverType, id)
	if err := deletion.ConfirmPrompt(cmd, promptMsg); err != nil {
		return err
	}

	req := switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest{
		Spec: switchoverv1.SwitchoverV1SwitchoverPairFailoverRequestSpec{
			EnvironmentCrn: environmentCrn,
			FailoverType:   failoverType,
		},
	}
	if activeMember != "" {
		req.Spec.ActiveMember = switchoverv1.PtrString(activeMember)
	}

	result, err := c.V2Client.TriggerSwitchoverPairFailover(id, req)
	if err != nil {
		return err
	}

	return printSwitchoverPair(cmd, result)
}
