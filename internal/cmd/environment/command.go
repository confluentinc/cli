package environment

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "environment",
		Aliases:     []string{"env"},
		Short:       "Manage and select Confluent Cloud environments.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newUseCommand())

	return cmd
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteEnvironments(c.Client, c.V2Client, c.Context)
}

func convertNameToId(input string, v2Client *ccloudv2.Client) (string, error) {
	envs, err := v2Client.ListOrgEnvironments()
	if err != nil {
		return "", err
	}
	envPtrs := resource.ConvertToPtrSlice(envs)
	id, err := resource.ConvertV2NameToId(input, envPtrs)
	if err != nil {
		return "", err
	}
	return id, nil
}
