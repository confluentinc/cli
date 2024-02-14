package ai

import (
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"

	"github.com/confluentinc/go-prompt"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
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

const message = `Hi, I’m the Confluent AI Assistant. I can help you with questions and tasks related to Confluent Cloud. Ask me things like:
- How many Kafka clusters do I have?
- What is an enterprise Kafka cluster?
- How do I get started with Flink on Confluent Cloud?
- Can I chat with support?

The Confluent AI Assistant does not have access to data inside topics and cannot execute code. Visit the Confluent AI Assistant documentation to learn more.
https://docs.confluent.io/cloud/current/release-notes/cflt-assistant.html

I'm currently in preview and am learning more everyday. Help me improve by rating answers with "+1" or "-1".
You can exit the shell at any time with "exit".`

func (c *command) ai(cmd *cobra.Command, _ []string) error {
	availability, err := c.V2Client.GetAvailability()
	if err != nil {
		return err
	}

	if !availability.GetAiAssistantEnabled() {
		return errors.NewErrorWithSuggestions(
			"the AI assistant is not enabled for your organization",
			"See here for more information: https://docs.confluent.io/cloud/current/release-notes/cflt-assistant.html",
		)
	}

	output.Println(c.Config.EnableColor, message)

	stdinState := getStdin()
	defer restoreStdin(stdinState)

	shell := newShell(c.V2Client).AddCleanupFunction(func() { restoreStdin(stdinState) })
	prompt.New(shell.executor, shell.completer, prompt.OptionPrefix("> ")).Run()

	return nil
}

func getStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState: %v", err)
		return nil
	}
	return state
}

func restoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}
