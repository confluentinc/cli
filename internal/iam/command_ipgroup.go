package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

type ipGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipGroupOut struct {
	ID         string   `human:"ID" serialized:"id"`
	GroupName  string   `human:"Group Name" serialized:"group_name"`
	CidrBlocks []string `human:"CIDR Blocks" serialized:"cidr_blocks"`
}

func newIPGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip-group",
		Short: "Manage IP Groups",
		Long:  "Manage IP Groups and their permissions",
	}

	c := &ipGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	return cmd
}

func printIPGroup(cmd *cobra.Command, ipGroup iamv2.IamV2IpGroup) error {
	table := output.NewTable(cmd)
	table.Add(&ipGroupOut{
		ID:         ipGroup.GetId(),
		GroupName:  ipGroup.GetGroupName(),
		CidrBlocks: ipGroup.GetCidrBlocks(),
	})
	return table.Print()
}
