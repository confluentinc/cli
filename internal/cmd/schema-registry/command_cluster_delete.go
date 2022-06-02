package schemaregistry

import (
	"context"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
	"os"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Schema Registry cluster for this environment.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.delete(cmd, args, form.NewPrompt(os.Stdin))
		},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete Schema Registry cluster for environment 'env-0000'",
				Code: fmt.Sprintf("%s schema-registry cluster delete --environment env-00000", version.CLIName),
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

	ctxClient := dynamicconfig.NewContextClient(c.Context)
	cluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}

	isDeleteConfirmed, err := deleteConfirmation(cmd, c.EnvironmentId(), prompt)
	if err != nil {
		return err
	}

	if !isDeleteConfirmed {
		utils.Println(cmd, "Terminating operation ...")
		return nil
	}

	err = c.Client.SchemaRegistry.DeleteSchemaRegistryCluster(ctx, cluster)
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.SchemaRegistryClusterDeletedMsg, c.EnvironmentId())
	return nil
}

func deleteConfirmation(cmd *cobra.Command, environmentId string, prompt form.Prompt) (bool, error) {
	f := form.New(
		form.Field{ID: "confirmation", Prompt: fmt.Sprintf("Are you sure you want to delete the Schema Registry "+
			"cluster for environment %s?", environmentId), IsYesOrNo: true},
	)
	if err := f.Prompt(cmd, prompt); err != nil {
		return false, errors.New(errors.SRFailedToReadDeletionConfirmationErrorMsg)
	}
	return f.Responses["confirmation"].(bool), nil
}
