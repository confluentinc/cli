package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <subject>",
		Short:       "Describe subject versions.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremDescribe,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Retrieve all versions registered under subject "payments" and its compatibility level.`,
				Code: fmt.Sprintf("%s schema-registry subject describe payments %s", version.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted schema.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *subjectCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return listSubjectVersions(cmd, args[0], srClient, ctx)
}
