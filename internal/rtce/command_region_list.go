package rtce

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *regionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rtce regions.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}
	pcmd.AddCloudAwsFlag(cmd)
	cmd.Flags().String("region", "", "Filter the results by exact match for region.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *regionCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	regions, err := c.V2Client.ListRtceRegions(cloud, region)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, region := range regions {
		out := &regionOut{
			ID:          region.GetId(),
			Cloud:       region.GetCloud(),
			Region:      region.GetRegion(),
			DisplayName: region.GetDisplayName(),
		}
		list.Add(out)
	}
	return list.Print()
}
