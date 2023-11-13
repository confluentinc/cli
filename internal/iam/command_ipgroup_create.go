package iam

import (
	"fmt"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *ipGroupCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an IP Group.",
		Args:  cobra.ExactArgs(0),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an IP Group named "demo-ip-group" with CIDR Blocks '["168.150.200.0/24", "147.150.200.0/24"]'':`,
				Code: `confluent iam ip-group create --group_name "demo-ip-group" --cidr_blocks "168.150.200.0/24,147.150.200.0/24"'`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("group_name", "", "Name of the IP Group being created")
	cmd.Flags().StringSlice("cidr_blocks", []string{}, "Array of CIDR Blocks to be associated with the IP Group")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddFilterFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group_name"))
	cobra.CheckErr(cmd.MarkFlagRequired("cidr_blocks"))

	return cmd
}

func (c *ipGroupCommand) create(cmd *cobra.Command, args []string) error {
	groupName, err := cmd.Flags().GetString("group_name")
	if err != nil {
		return err
	}

	fmt.Println("Group Name:", groupName)

	cidrBlocks, err := cmd.Flags().GetStringSlice("cidr_blocks")
	if err != nil {
		return err
	}

	fmt.Println("Cidr blocks 0:", cidrBlocks[0])
	fmt.Println("Cidr blocks 1:", cidrBlocks[1])
	fmt.Println("Cidr blocks:", cidrBlocks)

	createIPGroup := iamv2.IamV2IpGroup{
		GroupName:  &groupName,
		CidrBlocks: &cidrBlocks,
	}

	group, err := c.V2Client.CreateIamIPGroup(createIPGroup)

	fmt.Println("Created group:", group)

	return printIPGroup(cmd, group)

}
