package environment

import (
	"context"
	"fmt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	account, err := c.Client.Account.Get(context.Background(), &orgv1.Account{Id: id})
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.EnvNotFoundErrorMsg, id), errors.EnvNotFoundSuggestions)
	}
	c.Context.State.Auth.Account = account

	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, errors.EnvSwitchErrorMsg)
	}

	utils.Printf(cmd, errors.UsingEnvMsg, id)
	return nil
}
