package streamgovernance

import (
	"context"
	"fmt"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2-internal/stream-governance/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *streamGovernanceCommand) newEnableCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "enable",
		Short:       "Enable Stream Governance for this environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.enable),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Enable Stream Governance, using Google Cloud Platform in a region of choice with ADVANCED package",
				Code: fmt.Sprintf("%s stream-governance enable --cloud gcp --region <region_id> --package advanced", version.CLIName),
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

	newClusterRequest := c.createNewStreamGovernanceClusterRequest(cloud, region, packageType)

	//TODO: remove this line
	//PrintStreamGovernanceClusterOutput(cmd, *newClusterRequest)
	newClusterResponse, _, err := c.StreamGovernanceClient.ClustersStreamGovernanceV1Api.
		CreateStreamGovernanceV1Cluster(ctx).StreamGovernanceV1Cluster(*newClusterRequest).Execute()

	if err != nil {
		return err
	}

	PrintStreamGovernanceClusterOutput(cmd, newClusterResponse)
	return nil
}

func (c *streamGovernanceCommand) createNewStreamGovernanceClusterRequest(cloud, region, packageType string) *sgsdk.StreamGovernanceV1Cluster {
	newClusterRequest := sgsdk.NewStreamGovernanceV1ClusterWithDefaults()
	spec := sgsdk.NewStreamGovernanceV1ClusterSpecWithDefaults()

	envObjectReference := sgsdk.NewObjectReferenceWithDefaults()
	envObjectReference.SetId(c.EnvironmentId())

	spec.SetCloud(cloud)
	spec.SetPackage(packageType)
	spec.SetEnvironment(*envObjectReference)
	spec.SetRegion(region)
	newClusterRequest.SetSpec(*spec)

	return newClusterRequest
}
