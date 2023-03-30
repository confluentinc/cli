package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Schema Registry cloud regions.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the Schema Registry cloud regions for AWS in the "advanced" package.`,
				Code: "confluent schema-registry region list --cloud aws --package advanced",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	addPackageFlag(cmd, "")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	packageType, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	regionList, err := c.V2Client.ListSchemaRegistryCloudRegions(cloud, packageType)
	if err != nil {
		return err
	}

	return printRegionList(cmd, regionList)
}

func printRegionList(cmd *cobra.Command, regionList []srcm.SrcmV2Region) error {
	outputList := output.NewList(cmd)

	for _, region := range regionList {
		regionSpec := region.GetSpec()
		if output.GetFormat(cmd) == output.Human {
			outputList.Add(&schemaRegistryCloudRegionHumanOut{
				ID:         region.GetId(),
				Name:       regionSpec.GetDisplayName(),
				Cloud:      regionSpec.GetCloud(),
				RegionName: regionSpec.GetRegionName(),
				Packages:   strings.Join(regionSpec.GetPackages(), ", "),
			})
		} else {
			outputList.Add(&schemaRegistryCloudRegionSerializedOut{
				ID:         region.GetId(),
				Name:       regionSpec.GetDisplayName(),
				Cloud:      regionSpec.GetCloud(),
				RegionName: regionSpec.GetRegionName(),
				Packages:   regionSpec.GetPackages(),
			})
		}
	}

	outputList.Sort(false)
	return outputList.Print()
}
