package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *configCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <subject>",
		Short:       "Describe the config of a subject, or at global level.",
		Args:        cobra.MaximumNArgs(1),
		RunE:        pcmd.NewCLIRunE(c.onPremDescribe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the config of a given-name subject.",
				Code: fmt.Sprintf("%s schema-registry config describe <subject-name> %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *configCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}
