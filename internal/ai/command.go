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
	cmd.Run = c.ai

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

func (c *command) ai(cmd *cobra.Command, _ []string) {
	output.Println(c.Config.EnableColor, message)

	shell := newShell(c.V2Client)
	prompt.New(shell.executor, shell.completer, prompt.OptionPrefix("> ")).Run()
}
