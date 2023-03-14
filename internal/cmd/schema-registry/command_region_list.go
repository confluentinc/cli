package schemaregistry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type region struct {
	CloudId    string `human:"Cloud ID" serialized:"cloud_id"`
	CloudName  string `human:"Cloud Name" serialized:"cloud_name"`
	RegionId   string `human:"Region ID" serialized:"region_id"`
	RegionName string `human:"Region Name" serialized:"region_name"`
}

func (c *command) newRegionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List regions for a certain location and cloud provider.",
		Args:        cobra.NoArgs,
		RunE:        c.regionList,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("geo", "", fmt.Sprintf("Specify the geo as %s.", utils.ArrayToCommaDelimitedString(availableGeos)))
	addPackageFlag(cmd, essentialsPackage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) regionList(cmd *cobra.Command, _ []string) error {
	metadataList, err := c.Client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return err
	}

	var regions []*region

	for _, metadata := range metadataList {
		if cloud != "" && cloud != metadata.GetId() {
			continue
		}

		for _, r := range metadata.GetRegions() {
			if !r.GetIsSchedulable() {
				continue
			}

			regions = append(regions, &region{
				CloudId:    metadata.GetId(),
				CloudName:  metadata.GetName(),
				RegionId:   r.GetId(),
				RegionName: r.GetName(),
			})
		}
	}
	return nil
}
