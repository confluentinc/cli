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
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, _ []string, prompt form.Prompt) error {
	ctx := context.Background()

	cluster, err := c.Context.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}

	isDeleteConfirmed, err := confirmDeletion(cmd, c.EnvironmentId(), prompt)
	if err != nil {
		return err
	}

	if !isDeleteConfirmed {
		utils.Println(cmd, "Operation terminated.")
		return nil
	}

	err = c.Client.SchemaRegistry.DeleteSchemaRegistryCluster(ctx, cluster)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.SchemaRegistryClusterDeletedMsg, c.EnvironmentId())
	return nil
}

func confirmDeletion(cmd *cobra.Command, environmentId string, prompt form.Prompt) (bool, error) {
	f := form.New(
		form.Field{
			ID:        "confirmation",
			Prompt:    fmt.Sprintf(`Are you sure you want to delete the Schema Registry cluster for environment "%s"?`, environmentId),
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(cmd, prompt); err != nil {
		return false, errors.New(errors.FailedToReadDeletionConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
