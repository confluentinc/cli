package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/network"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newRegionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		Args:  cobra.NoArgs,
		RunE:  c.regionList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) regionList(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	regions, err := network.ListRegions(c.Client, cloud)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, region := range regions {
		list.Add(region)
	}
	return list.Print()
}
