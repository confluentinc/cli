package schemaregistry

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *clusterCommand) newUpgradeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       "Upgrade Schema Registry for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.upgrade,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Upgrade Schema Registry to "advanced" package for environment "env-12345"`,
				Code: fmt.Sprintf("%s schema-registry cluster upgrade --package advanced --environment env-12345", version.CLIName),
			},
		),
	}

	pcmd.AddStreamGovernancePackageFlag(cmd, getAllPackageDisplayNames())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("package")

	return cmd
}

func (c *clusterCommand) upgrade(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	packageInternalName, isValid := getPackageInternalName(packageDisplayName)
	if !isValid {
		return errors.New(fmt.Sprintf(errors.SRInvalidPackageType, packageDisplayName))
	}

	cluster, err := c.Context.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}

	cluster.Package = packageInternalName
	_, err = c.Client.SchemaRegistry.UpdateSchemaRegistryCluster(ctx, cluster)

	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.SchemaRegistryClusterUpgradedMsg, c.EnvironmentId(), packageDisplayName)
	return nil
}
