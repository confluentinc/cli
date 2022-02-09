package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List subjects.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.onPremList),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all subjects available in a Schema Registry.",
				Code: fmt.Sprintf("%s schema-registry subject list %s", version.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted subjects.")
	cmd.Flags().String("prefix", ":*:", "Subject prefix.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *subjectCommand) onPremList(cmd *cobra.Command, _ []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return listSubjects(cmd, srClient, ctx)
}
