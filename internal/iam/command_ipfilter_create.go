package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

func (c *ipFilterCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an IP filter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an IP filter named "demo-ip-filter" with resource group "management" and IP groups "ipg-12345" and "ipg-67890":`,
				Code: `confluent iam ip-filter create "demo-ip-filter" --resource-group "management" --ip-groups "ipg-12345,ipg-67890"`,
			},
		),
	}

	cmd.Flags().String("resource-group", "management", "Name of resource group. Currently, only \"management\" is supported.")
	cmd.Flags().StringSlice("ip-groups", []string{}, "Comma-separated list of IP group IDs")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("resource-group"))
	cobra.CheckErr(cmd.MarkFlagRequired("ip-groups"))

	return cmd
}

func (c *ipFilterCommand) create(cmd *cobra.Command, args []string) error {
	resourceGroup, err := cmd.Flags().GetString("resource-group")
	if err != nil {
		return err
	}

	ipGroups, err := cmd.Flags().GetStringSlice("ip-groups")
	if err != nil {
		return err
	}

	// Convert the IP group IDs into IP group objects
	var IpGroupIdObjects []iamv2.GlobalObjectReference
	for _, ipGroupId := range ipGroups {
		// The empty string fields will get filled in automatically by the cc-policy-service
		IpGroupIdObjects = append(IpGroupIdObjects, *iamv2.NewGlobalObjectReference(ipGroupId, "", ""))
	}

	createIPFilter := iamv2.IamV2IpFilter{
		FilterName:    &args[0],
		ResourceGroup: &resourceGroup,
		IpGroups:      &IpGroupIdObjects,
	}

	filter, err := c.V2Client.CreateIamIpFilter(createIPFilter)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanIpFilter(cmd, filter)
	}
	return printSerializedIpFilter(cmd, filter)
}
