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

type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewRegionCommand returns the Cobra command for Kafka region.
func NewRegionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage Confluent Cloud regions.",
		Long:        "Use this command to manage Confluent Cloud regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &regionCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.AddCommand(c.newListCommand())
	return c.Command
}

func (c *regionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		Args:  cobra.NoArgs,
		RunE: pcmd.NewCLIRunE(func(cmd *cobra.Command, _ []string) error {
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
		}),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}
