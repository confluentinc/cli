package completion

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/spf13/cobra"
)

const longDescriptionTemplate = `Use this command to print the output shell completion
code for the specified shell (Bash only). The shell code must be evaluated to provide
interactive completion of {{.CLIName}} commands.

Install Bash completions on macOS:

::

  # Enable Bash completions using homebrew
  brew install bash-completion
  # Set the {{.CLIName}} completion code for Bash to a file that's sourced on login
  {{.CLIName}} completion bash > $(brew --prefix)/etc/bash_completion.d/{{.CLIName}}

Install Bash completions on Linux:

::

  # Set the {{.CLIName}} completion code for Bash to a file that's sourced on login
  {{.CLIName}} completion bash > /etc/bash_completion.d/{{.CLIName}}

  # Load the {{.CLIName}} completion code for Bash into the current shell
  source /etc/bash_completion.d/{{.CLIName}}

To update your completion scripts after updating the CLI, run "{{.CLIName}} completion bash"
again and overwrite the file initially created above.
`

type completionCommand struct {
	*cobra.Command
	rootCmd *cobra.Command
}

// NewCompletionCmd returns the Cobra command for shell completion.
func NewCompletionCmd(rootCmd *cobra.Command, cliName string) *cobra.Command {
	cmd := &completionCommand{
		rootCmd: rootCmd,
	}
	cmd.init(cliName)
	return cmd.Command
}

func (c *completionCommand) init(cliName string) {
	c.Command = &cobra.Command{
		Use:   "completion <shell>",
		Short: "Print shell completion code.",
		Long:  getLongDescription(cliName),
		RunE:  c.completion,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *completionCommand) completion(cmd *cobra.Command, args []string) error {
	var err error
	if args[0] == "bash" {
		err = c.rootCmd.GenBashCompletion(cmd.OutOrStdout())
	} else {
		err = fmt.Errorf(`unsupported shell type "%s"`, args[0])
	}
	return err
}

func getLongDescription(cliName string) string {
	t := template.Must(template.New("longDescription").Parse(longDescriptionTemplate))
	buf := new(bytes.Buffer)
	data := map[string]interface{}{"CLIName": cliName}
	if err := t.Execute(buf, data); err != nil {
		// We're okay with this since its definitely a development error; should never happen to users
		panic(err)
	}
	return buf.String()
}
