package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *schemaCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List all schemas for given subject prefix",
		Args:        cobra.NoArgs,
		RunE:        c.listOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all schemas for given subject under default context.",
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix my-subject", pversion.CLIName),
			},
			examples.Example{
				Text: "List all schemas under given context.",
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix :.mycontext:", pversion.CLIName),
			},
			examples.Example{
				Text: "List all schemas under default context.",
				Code: fmt.Sprintf("%s schema-registry schema list", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject-prefix", "S", "", "Subject prefix to list the schemas from.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *schemaCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return c.listSchemas(cmd, srClient, ctx)
}
