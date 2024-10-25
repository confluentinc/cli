package iam

import (
	"slices"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type ipGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipGroupOut struct {
	ID         string   `human:"ID" serialized:"id"`
	Name       string   `human:"Name" serialized:"name"`
	CidrBlocks []string `human:"CIDR blocks" serialized:"cidr_blocks"`
}

func newIpGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ip-group",
		Short:       "Manage IP groups.",
		Long:        "Manage IP groups and their permissions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ipGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printIpGroup(cmd *cobra.Command, ipGroup iamipfilteringv2.IamV2IpGroup) error {
	cidrBlocks := ipGroup.GetCidrBlocks()
	slices.Sort(cidrBlocks)

	table := output.NewTable(cmd)
	table.Add(&ipGroupOut{
		ID:         ipGroup.GetId(),
		Name:       ipGroup.GetGroupName(),
		CidrBlocks: cidrBlocks,
	})
	return table.Print()
}

func (c *ipGroupCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteIpGroups(c.V2Client)
}
