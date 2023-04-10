package environment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Choose a Confluent Cloud environment to be used in subsequent commands.",
		Long:              "Choose a Confluent Cloud environment to be used in subsequent commands which support passing an environment with the `--environment` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	environment, err := c.Client.Account.Get(context.Background(), &ccloudv1.Account{Id: args[0]})
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.EnvNotFoundErrorMsg, args[0]), fmt.Sprintf(errors.OrgResourceNotFoundSuggestions, resource.Environment))
	}
	c.Context.UseEnvironment(environment)

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, errors.EnvSwitchErrorMsg)
	}

	output.Printf("Using environment \"%s\".\n", args[0])
	return nil
}
