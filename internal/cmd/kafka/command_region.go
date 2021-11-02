package kafka

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

type region struct {
	CloudId    string
	CloudName  string
	RegionId   string
	RegionName string
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

			regions, err := ListRegions(c.Client, cloud)
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

	cmd.Flags().String("cloud", "", "The cloud ID to filter by.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false

	return cmd
}

func ListRegions(client *ccloud.Client, cloud string) ([]*region, error) {
	metadataList, err := client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return nil, err
	}

	var regions []*region

	for _, metadata := range metadataList {
		if cloud != "" && cloud != metadata.Id {
			continue
		}

		for _, r := range metadata.Regions {
			if !r.IsSchedulable {
				continue
			}

			regions = append(regions, &region{
				CloudId:    metadata.Id,
				CloudName:  metadata.Name,
				RegionId:   r.Id,
				RegionName: r.Name,
			})
		}
	}

	return regions, nil
}
