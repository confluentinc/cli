package streamgovernance

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *streamGovernanceCommand) newDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the Stream Governance cluster for an environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.describe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe Stream Governance cluster for environment 'env-00000'",
				Code: fmt.Sprintf("%s stream-governance describe --environment env-00000", version.CLIName),
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *streamGovernanceCommand) describe(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	clusterId, err := c.getClusterIdFromEnvironment(ctx)
	if err != nil {
		return errors.NewStreamGovernanceNotEnabledError()
	}

	clusterDescription, _, err := c.V2Client.StreamGovernanceClient.ClustersStreamGovernanceV2Api.
		GetStreamGovernanceV2Cluster(ctx, clusterId).Execute()
	if err != nil {
		return err
	}

	spec := clusterDescription.GetSpec()
	regionSpec := spec.GetRegion()
	streamGovernanceV2Region, err := c.getStreamGovernanceV2RegionFromId(regionSpec.GetId())
	if err != nil {
		return err
	}

	PrintStreamGovernanceClusterOutput(cmd, clusterDescription, *streamGovernanceV2Region)
	return nil
}
