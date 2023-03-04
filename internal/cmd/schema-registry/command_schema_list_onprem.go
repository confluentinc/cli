package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *command) newSchemaListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List schemas for a given subject prefix.",
		Args:        cobra.NoArgs,
		RunE:        c.schemaListOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all schemas for subjects with prefix "my-subject".`,
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix my-subject %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: `List all schemas for all subjects in context ":.mycontext:".`,
				Code: fmt.Sprintf("%s schema-registry schema list --subject-prefix :.mycontext: %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: "List all schemas in the default context.",
				Code: fmt.Sprintf("%s schema-registry schema list %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("subject-prefix", "", "List schemas for subjects with a given prefix.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) schemaListOnPrem(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return c.listSchemas(cmd, srClient, ctx)
}
