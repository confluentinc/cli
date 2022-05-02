package streamgovernance

import (
	"context"
	"fmt"
	sgsdk "github.com/confluentinc/ccloud-sdk-go-v2/stream-governance/v2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *streamGovernanceCommand) newUpgradeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       "Upgrade Stream Governance Package for an environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.upgrade),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Upgrade Stream Governance package to 'advanced' for environment 'env-00000'",
				Code: fmt.Sprintf("%s stream-governance upgrade --package advanced "+
					"--environment env-00000", version.CLIName),
			},
		),
	}

	pcmd.AddStreamGovernancePackageFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("package")

	return cmd
}

func (c *streamGovernanceCommand) upgrade(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	clusterId, err := c.getClusterIdFromEnvironment(ctx)
	if err != nil {
		return errors.NewStreamGovernanceNotEnabledError()
	}

	// Collect the parameter
	packageType, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	newClusterUpdateRequest := c.createNewStreamGovernanceClusterUpdateRequest(packageType)
	updatedClusterResponse, _, err := c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.
		UpdateStreamGovernanceV2Cluster(ctx, clusterId).StreamGovernanceV2ClusterUpdate(*newClusterUpdateRequest).Execute()

	if err != nil {
		return err
	}

	spec := updatedClusterResponse.GetSpec()
	regionSpec := spec.GetRegion()
	streamGovernanceV2Region, err := c.getStreamGovernanceV2RegionFromId(regionSpec.GetId())
	if err != nil {
		return err
	}

	PrintStreamGovernanceClusterOutput(cmd, updatedClusterResponse, *streamGovernanceV2Region)
	return nil
}

func (c *streamGovernanceCommand) createNewStreamGovernanceClusterUpdateRequest(packageType string) *sgsdk.StreamGovernanceV2ClusterUpdate {
	newClusterUpdateRequest := sgsdk.NewStreamGovernanceV2ClusterUpdateWithDefaults()
	spec := sgsdk.NewStreamGovernanceV2ClusterSpecUpdateWithDefaults()
	envObjectReference := sgsdk.NewObjectReferenceWithDefaults()
	envObjectReference.SetId(c.EnvironmentId())

	spec.SetPackage(packageType)
	spec.SetEnvironment(*envObjectReference)
	newClusterUpdateRequest.SetSpec(*spec)

	return newClusterUpdateRequest
}
