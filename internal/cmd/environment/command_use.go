package environment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Switch to the specified Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	id := args[0]

	environment, err := c.Client.Account.Get(context.Background(), &ccloudv1.Account{Id: id})
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.EnvNotFoundErrorMsg, id), errors.EnvNotFoundSuggestions)
	}
	c.Context.SetEnvironment(environment)

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, errors.EnvSwitchErrorMsg)
	}

	output.Printf(errors.UsingEnvMsg, id)
	return nil
}
