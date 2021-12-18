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
		Use:     "describe <id>",
		Short:   "Get schema either by schema ID, or by subject/version.",
		Args:    cobra.MaximumNArgs(1),
		PreRunE: pcmd.NewCLIPreRunnerE(c.preDescribe),
		RunE:    pcmd.NewCLIRunE(c.onPremDescribe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the schema string by schema ID.",
				Code: fmt.Sprintf("%s schema-registry schema describe 1337", pversion.CLIName),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().StringP("version", "V", "", "Version of the schema. Can be a specific version or 'latest'.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())

	return cmd
}

func (c *schemaCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return c.describeById(cmd, args, srClient, ctx)
	}
	return c.describeBySubject(cmd, srClient, ctx)
}
