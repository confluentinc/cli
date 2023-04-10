package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSchemaDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete",
		Short:       "Delete one or more schemas.",
		Long:        "Delete one or more schemas. This command should only be used if absolutely necessary.",
		Args:        cobra.NoArgs,
		RunE:        c.schemaDeleteOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Soft delete the latest version of subject "payments".`,
				Code: fmt.Sprintf("confluent schema-registry schema delete --subject payments --version latest %s", OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version, "all", or "latest".`)
	cmd.Flags().Bool("permanent", false, "Permanently delete the schema.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("subject"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *command) schemaDeleteOnPrem(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return deleteSchema(cmd, srClient, ctx)
}
