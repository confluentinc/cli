package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

func (c *configCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe top-level or subject-level schema compatibility.",
		Args:        cobra.MaximumNArgs(0),
		RunE:        c.onPremDescribe,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the configuration of a subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry config describe --subject payments %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: "Describe the top-level configuration.",
				Code: fmt.Sprintf("%s schema-registry config describe %s", pversion.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *configCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}
