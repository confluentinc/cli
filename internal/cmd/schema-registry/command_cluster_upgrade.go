package schemaregistry

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newClusterUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "upgrade",
		Short:       "Upgrade the Schema Registry package for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterUpgrade,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Upgrade Schema Registry to the "advanced" package for environment "env-12345".`,
				Code: "confluent schema-registry cluster upgrade --package advanced --environment env-12345",
			},
		),
	}

	addPackageFlag(cmd, "")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("package"))

	return cmd
}

func (c *command) clusterUpgrade(cmd *cobra.Command, _ []string) error {
	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	packageInternalName, err := getPackageInternalName(packageDisplayName)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	ctx := context.Background()
	cluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, environmentId)
	if err != nil {
		return err
	}

	if packageInternalName == cluster.Package {
		output.ErrPrintf(errors.SRInvalidPackageUpgrade, environmentId, packageDisplayName)
		return nil
	}
	cluster.Package = packageInternalName

	if _, err := c.Client.SchemaRegistry.UpdateSchemaRegistryCluster(ctx, cluster); err != nil {
		return err
	}

	output.Printf(errors.SchemaRegistryClusterUpgradedMsg, environmentId, packageDisplayName)
	return nil
}
