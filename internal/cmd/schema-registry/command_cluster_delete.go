package schemaregistry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *command) newClusterDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete",
		Short:       "Delete the Schema Registry cluster for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterDelete,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the Schema Registry cluster for environment "env-12345"`,
				Code: fmt.Sprintf("%s schema-registry cluster delete --environment env-12345", version.CLIName),
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) clusterDelete(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, environmentId)
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(`Are you sure you want to delete %s "%s" for %s "%s"?`, resource.SchemaRegistryCluster, cluster.Id, resource.Environment, environmentId)
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	err = c.Client.SchemaRegistry.DeleteSchemaRegistryCluster(ctx, cluster)
	if err != nil {
		return err
	}

	output.Printf(errors.SchemaRegistryClusterDeletedMsg, environmentId)
	return nil
}
