package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe subject versions and compatibility.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDescribe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Retrieve all versions registered under a given subject and its compatibility level.",
				Code: fmt.Sprintf("%s schema-registry subject describe <subject-name> %s", version.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().BoolP("deleted", "D", false, "View the deleted schema.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *subjectCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return listSubjectVersions(cmd, args[0], srClient, ctx)
}
