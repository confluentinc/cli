package streamgovernance

import (
	"fmt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *streamGovernanceCommand) newEnableCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "enable",
		Short:       "Enable Stream Governance for an environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.enable),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Enable Stream Governance, using AWS in us-east-2 region with 'advanced' " +
					"package for environment 'env-00000'",
				Code: fmt.Sprintf("%s stream-governance enable --cloud aws --region us-east-2 "+
					"--package advanced --environment env-00000", version.CLIName),
			},
		),
	}

	pcmd.AddStreamGovernancePackageFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "",
		`Cloud region name (use "confluent stream-governance region list" to see all).`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("region")
	_ = cmd.MarkFlagRequired("package")

	return cmd
}

func (c *streamGovernanceCommand) enable(cmd *cobra.Command, _ []string) error {
	ctx := c.V2Client.StreamGovernanceApiContext()

	// Collect the parameters
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	clouds, err := c.Client.EnvironmentMetadata.Get(ctx)
	if err != nil {
		return err
	}

	if err := checkCloudProvider(cloud, clouds); err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	packageType, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	streamGovernanceV2Region, err := c.getStreamGovernanceV2Region(cloud, region, packageType, ctx)
	if err != nil {
		return err
	}

	newClusterRequest := c.createNewStreamGovernanceClusterRequest(streamGovernanceV2Region, packageType)

	newClusterResponse, _, err := c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.
		CreateStreamGovernanceV2Cluster(ctx).StreamGovernanceV2Cluster(*newClusterRequest).Execute()

	if err != nil {
		existingCluster, getExistingErr := c.getStreamGovernanceV2ClusterForEnvironment(ctx)
		if getExistingErr != nil {
			return err
		}

		spec := existingCluster.GetSpec()
		regionSpec := spec.GetRegion()
		existingRegion, getRegionErr := c.getStreamGovernanceV2RegionFromId(regionSpec.GetId(), ctx)
		if getRegionErr != nil {
			return getRegionErr
		}
		PrintStreamGovernanceClusterOutput(cmd, *existingCluster, *existingRegion)
	} else {
		PrintStreamGovernanceClusterOutput(cmd, newClusterResponse, *streamGovernanceV2Region)
	}

	return nil
}

func (c *streamGovernanceCommand) createNewStreamGovernanceClusterRequest(
	streamGovernanceV2Region *sgsdk.StreamGovernanceV2Region, packageType string) *sgsdk.StreamGovernanceV2Cluster {

	newClusterRequest := sgsdk.NewStreamGovernanceV2ClusterWithDefaults()
	spec := sgsdk.NewStreamGovernanceV2ClusterSpecWithDefaults()

	envObjectReference := sgsdk.NewObjectReferenceWithDefaults()
	envObjectReference.SetId(c.EnvironmentId())

	regionObjectReference := sgsdk.NewObjectReferenceWithDefaults()
	regionObjectReference.SetId(streamGovernanceV2Region.GetId())

	spec.SetPackage(packageType)
	spec.SetEnvironment(*envObjectReference)
	spec.SetRegion(*regionObjectReference)
	newClusterRequest.SetSpec(*spec)

	return newClusterRequest
}

func checkCloudProvider(cloudId string, clouds []*schedv1.CloudMetadata) error {
	for _, cloud := range clouds {
		if cloudId == cloud.Id {
			return nil
		}
	}
	return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SGCloudProviderNotAvailableErrorMsg, cloudId),
		errors.SGCloudProviderNotAvailableSuggestions)
}
