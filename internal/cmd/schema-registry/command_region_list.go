package schemaregistry

import (
	"fmt"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	regionListFields           = []string{"ID", "Cloud", "RegionName", "DisplayName", "Packages"}
	regionListHumanLabels      = []string{"ID", "Cloud", "Region Name", "Display Name", "Packages"}
	regionListStructuredLabels = []string{"id", "cloud", "region_name", "display_name", "packages"}
)

func (c *regionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Schema Registry cloud regions.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the schema registry cloud regions in "aws" for "advanced" package`,
				Code: fmt.Sprintf("%s schema-registry region list --cloud aws --package advanced", version.CLIName),
			},
		),
	}

	addPackageFlag(cmd, "")
	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *regionCommand) list(cmd *cobra.Command, _ []string) error {
	ctx := c.V2Client.SchemaRegistryApiContext()

	// Collect the parameters
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	packageType, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	regionListRequest := c.V2Client.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2Regions(ctx)

	if cloud != "" {
		regionListRequest = regionListRequest.SpecCloud(cloud)
	}

	if packageType != "" {
		packageSpec := []string{packageType}
		regionListRequest = regionListRequest.SpecPackages(packageSpec)
	}

	var regionList []srcm.SrcmV2Region
	done := false
	pageToken := ""
	for !done {
		regionListRequest = regionListRequest.PageToken(pageToken)
		regionPage, _, err := c.V2Client.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2RegionsExecute(regionListRequest)
		if err != nil {
			return err
		}
		regionList = append(regionList, regionPage.GetData()...)

		pageToken, done, err = c.V2Client.ExtractNextPageToken(regionPage.GetMetadata().Next)
		if err != nil {
			return err
		}
	}

	return printRegionList(cmd, regionList)
}

func printRegionList(cmd *cobra.Command, regionList []srcm.SrcmV2Region) error {
	listOutputWriter, err := output.NewListOutputWriter(cmd, regionListFields, regionListHumanLabels, regionListStructuredLabels)
	if err != nil {
		return err
	}

	for _, region := range regionList {
		regionSpec := region.GetSpec()
		v2Region := &schemaRegistryCloudRegion{
			ID:          region.GetId(),
			Cloud:       regionSpec.GetCloud(),
			RegionName:  regionSpec.GetRegionName(),
			DisplayName: regionSpec.GetDisplayName(),
			Packages:    regionSpec.GetPackages(),
		}

		listOutputWriter.AddElement(v2Region)
	}
	return listOutputWriter.Out()
}
