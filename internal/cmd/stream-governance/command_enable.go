package streamgovernance

import (
	"context"
	"fmt"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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
				Text: "Enable Stream Governance, using Google Cloud Platform in a region of choice with 'advanced' " +
					"package for environment 'env-00000'",
				Code: fmt.Sprintf("%s stream-governance enable --cloud gcp --region <region_id> "+
					"--package advanced --environment env-00000", version.CLIName),
			},
		),
	}

	pcmd.AddStreamGovernancePackageFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "",
		`Cloud region ID for cluster (use "confluent stream-governance region list" to see all).`)
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
	ctx := context.Background()

	// Collect the parameters
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
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

	streamGovernanceV2Region, err := c.getRegionObject(cloud, region, packageType)
	if err != nil {
		return nil
	}

	newClusterRequest, err := c.createNewStreamGovernanceClusterRequest(streamGovernanceV2Region, packageType)
	if err != nil {
		return err
	}

	//TODO: remove this line
	//PrintStreamGovernanceClusterOutput(cmd, *newClusterRequest, *streamGovernanceV2Region)
	newClusterResponse, _, err := c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.
		CreateStreamGovernanceV2Cluster(ctx).StreamGovernanceV2Cluster(*newClusterRequest).Execute()

	if err != nil {
		return err
	}

	PrintStreamGovernanceClusterOutput(cmd, newClusterResponse, *streamGovernanceV2Region)
	return nil
}

func (c *streamGovernanceCommand) createNewStreamGovernanceClusterRequest(
	streamGovernanceV2Region *sgsdk.StreamGovernanceV2Region, packageType string) (*sgsdk.StreamGovernanceV2Cluster, error) {

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

	return newClusterRequest, nil
}
