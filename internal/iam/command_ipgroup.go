package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"strings"
)

type ipGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipGroupHumanOut struct {
	ID         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	CidrBlocks string `human:"CIDR blocks" serialized:"cidr_blocks"`
}

type ipGroupSerializedOut struct {
	ID         string   `human:"ID" serialized:"id"`
	Name       string   `human:"Name" serialized:"name"`
	CidrBlocks []string `human:"CIDR blocks" serialized:"cidr_blocks"`
}

func newIPGroupCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip-group",
		Short: "Manage IP groups",
		Long:  "Manage IP groups and their permissions",
	}

	c := &ipGroupCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	return cmd
}

func printHumanIPGroup(cmd *cobra.Command, ipGroup iamv2.IamV2IpGroup) error {
	table := output.NewTable(cmd)
	table.Add(&ipGroupHumanOut{
		ID:         ipGroup.GetId(),
		Name:       ipGroup.GetGroupName(),
		CidrBlocks: strings.Join(ipGroup.GetCidrBlocks(), ", "),
	})
	return table.Print()
}

func printSerializedIPGroup(cmd *cobra.Command, ipGroup iamv2.IamV2IpGroup) error {
	out := &ipGroupSerializedOut{
		ID:         ipGroup.GetId(),
		Name:       ipGroup.GetGroupName(),
		CidrBlocks: ipGroup.GetCidrBlocks(),
	}
	return output.SerializedOutput(cmd, out)
}
