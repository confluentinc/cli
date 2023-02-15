package schemaregistry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *clusterCommand) newUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       "Upgrade the Schema Registry package for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.upgrade,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Upgrade Schema Registry to the "advanced" package for environment "env-12345".`,
				Code: fmt.Sprintf("%s schema-registry cluster upgrade --package advanced --environment env-12345", version.CLIName),
			},
		),
	}

	addPackageFlag(cmd, "")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("package")

	return cmd
}

func (c *clusterCommand) upgrade(cmd *cobra.Command, _ []string) error {
	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	packageInternalName, err := getPackageInternalName(packageDisplayName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}

	if packageInternalName == cluster.Package {
		utils.ErrPrintf(cmd, errors.SRInvalidPackageUpgrade, c.EnvironmentId(), packageDisplayName)
		return nil
	}

	cluster.Package = packageInternalName
	_, err = c.Client.SchemaRegistry.UpdateSchemaRegistryCluster(ctx, cluster)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.SchemaRegistryClusterUpgradedMsg, c.EnvironmentId(), packageDisplayName)
	return nil
}
