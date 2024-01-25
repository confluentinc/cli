package ai

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/go-prompt"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ai",
		Short:       "Start an interactive AI shell",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.ai

	return cmd
}

func (c *command) ai(cmd *cobra.Command, _ []string) error {
	output.Println(c.Config.EnableColor, `Welcome to the Confluent AI Assistant! Exit with "exit", or rate an answer with "+1" or "-1".`)

	s := &shell{
		client:  c.V2Client,
		session: newSession(),
	}

	prompt.New(s.executor, s.completer, prompt.OptionPrefix("> ")).Run()

	return nil
}
