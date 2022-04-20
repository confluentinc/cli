package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *schemaCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe [id]",
		Short:       "Get schema either by schema ID, or by subject/version.",
		Args:        cobra.MaximumNArgs(1),
		PreRunE:     pcmd.NewCLIPreRunnerE(c.preDescribe),
		RunE:        pcmd.NewCLIRunE(c.onPremDescribe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the schema string by schema ID.",
				Code: fmt.Sprintf("%s schema-registry schema describe 1337 %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: "Describe the schema string by both subject and version.",
				Code: fmt.Sprintf("%s schema-registry schema describe --subject payments --version latest %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", `Version of the schema. Can be a specific version or "latest".`)
	cmd.Flags().Bool("show-refs", false, "Display the entire schema graph, including references.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *schemaCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	showRefs, err := cmd.Flags().GetBool("show-refs")
	if err != nil {
		return err
	}

	var id string
	if len(args) > 1 {
		id = args[0]
	}

	if showRefs {
		return describeGraph(cmd, id, srClient, ctx)
	}

	if id != "" {
		return describeById(cmd, id, srClient, ctx)
	}
	return describeBySubject(cmd, srClient, ctx)
}
