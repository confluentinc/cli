package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *schemaCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete",
		Short:       "Delete one or more schemas.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.onPremDelete),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete one or more topics. This command should only be used in extreme circumstances.",
				Code: fmt.Sprintf("%s schema-registry schema delete --subject payments --version latest %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version, 'all', or 'latest'.")
	cmd.Flags().BoolP("permanent", "P", false, "Permanently delete the schema.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())

	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}

func (c *schemaCommand) onPremDelete(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return deleteSchema(cmd, srClient, ctx)
}
