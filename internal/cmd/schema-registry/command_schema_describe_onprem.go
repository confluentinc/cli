package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSchemaDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe [id]",
		Short:       "Get schema either by schema ID, or by subject/version.",
		Args:        cobra.MaximumNArgs(1),
		PreRunE:     c.preDescribe,
		RunE:        c.schemaDescribeOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the schema string by schema ID.",
				Code: fmt.Sprintf("confluent schema-registry schema describe 1337 %s", OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: "Describe the schema string by both subject and version.",
				Code: fmt.Sprintf("confluent schema-registry schema describe --subject payments --version latest %s", OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().String("version", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().Bool("show-references", false, "Display the entire schema graph, including references.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) schemaDescribeOnPrem(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	showReferences, err := cmd.Flags().GetBool("show-references")
	if err != nil {
		return err
	}

	var id string
	if len(args) > 1 {
		id = args[0]
	}

	if showReferences {
		return describeGraph(cmd, id, srClient, ctx)
	}

	if id != "" {
		return describeById(id, srClient, ctx)
	}
	return describeBySubject(cmd, srClient, ctx)
}
