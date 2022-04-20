package streamgovernance

import (
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	enableLabels            = []string{"Id", "SchemaRegistryEndpoint"}
	enableHumanRenames      = map[string]string{"ID": "Cluster ID", "SchemaRegistryEndpoint": "Endpoint URL"}
	enableStructuredRenames = map[string]string{"ID": "cluster_id", "SchemaRegistryEndpoint": "endpoint_url"}
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
				Code: fmt.Sprintf("%s stream-governance cluster enable --cloud gcp --region <region_id> --package advanced", version.CLIName),
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddStreamGovernancePackageFlag(cmd)
	cmd.Flags().String("region", "", "Region ID")
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

//TODO: Once new SDK is available
func (c *streamGovernanceCommand) enable(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Inside enable")

	return nil
}
