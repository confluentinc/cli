package iam

import (
	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *ipGroupCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an IP group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an IP group named "demo-ip-group" with CIDR blocks "168.150.200.0/24" and "147.150.200.0/24":`,
				Code: "confluent iam ip-group create demo-ip-group --cidr-blocks 168.150.200.0/24,147.150.200.0/24",
			},
		),
	}

	cmd.Flags().StringSlice("cidr-blocks", []string{}, "A comma-separated list of CIDR blocks in IP group.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cidr-blocks"))

	return cmd
}

func (c *ipGroupCommand) create(cmd *cobra.Command, args []string) error {
	cidrBlocks, err := cmd.Flags().GetStringSlice("cidr-blocks")
	if err != nil {
		return err
	}

	createIPGroup := iamipfilteringv2.IamV2IpGroup{
		GroupName:  &args[0],
		CidrBlocks: &cidrBlocks,
	}

	group, err := c.V2Client.CreateIamIpGroup(createIPGroup)
	if err != nil {
		return err
	}

	return printIpGroup(cmd, group)
}
