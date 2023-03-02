package schemaregistry

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the Schema Registry cluster for this environment.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.delete(cmd, args, form.NewPrompt(os.Stdin))
		},
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

	_ = cmd.MarkFlagRequired("environment")

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, _ []string, prompt form.Prompt) error {
	ctx := context.Background()

	cluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(`Are you sure you want to delete %s "%s" for %s "%s"?`, resource.SchemaRegistryCluster, cluster.Id, resource.Environment, c.EnvironmentId())
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	err = c.Client.SchemaRegistry.DeleteSchemaRegistryCluster(ctx, cluster)
	if err != nil {
		return err
	}

	output.Printf(errors.SchemaRegistryClusterDeletedMsg, c.EnvironmentId())
	return nil
}
