package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *ipGroupCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <ip-group-name>",
		Short: "Create an IP Group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an IP Group named "demo-ip-group" with CIDR Blocks '["168.150.200.0/24", "147.150.200.0/24"]'':`,
				Code: `confluent iam ip-group create "demo-ip-group" --cidr-blocks "168.150.200.0/24,147.150.200.0/24"`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("cidr-blocks", []string{}, "Array of CIDR Blocks to be associated with the IP Group")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddFilterFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cidr-blocks"))

	return cmd
}

func (c *ipGroupCommand) create(cmd *cobra.Command, args []string) error {

	cidrBlocks, err := cmd.Flags().GetStringSlice("cidr-blocks")
	if err != nil {
		return err
	}

	createIPGroup := iamv2.IamV2IpGroup{
		GroupName:  &args[0],
		CidrBlocks: &cidrBlocks,
	}

	group, err := c.V2Client.CreateIamIPGroup(createIPGroup)

	return printIPGroup(cmd, group)
}
