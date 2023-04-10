package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSubjectDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <subject>",
		Short:       "Describe subject versions.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.subjectDescribeOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Retrieve all versions registered under subject "payments" and its compatibility level.`,
				Code: fmt.Sprintf("confluent schema-registry subject describe payments %s", OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().Bool("deleted", false, "View the deleted schema.")
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) subjectDescribeOnPrem(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return listSubjectVersions(cmd, args[0], srClient, ctx)
}
