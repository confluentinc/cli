package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	regionListFields           = []string{"CloudId", "CloudName", "RegionId", "RegionName"}
	regionListHumanLabels      = []string{"Cloud ID", "Cloud Name", "Region ID", "Region Name"}
	regionListStructuredLabels = []string{"cloud_id", "cloud_name", "region_id", "region_name"}
)

func (c *regionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *regionCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, _ := cmd.Flags().GetString("cloud")

	regions, err := kafka.ListRegions(c.Client, cloud)
	if err != nil {
		return err
	}

	w, err := output.NewListOutputWriter(cmd, regionListFields, regionListHumanLabels, regionListStructuredLabels)
	if err != nil {
		return err
	}

	for _, region := range regions {
		w.AddElement(region)
	}

	return w.Out()
}
