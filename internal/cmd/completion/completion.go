package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const longDescription = `Use this command to print the output Shell completion
code for the specified shell (Bash/Zsh only). The shell code must be evaluated to provide
interactive completion of ` + "`confluent`" + ` commands.

Install Bash completions on macOS:
  #.  Install Homebrew (https://brew.sh/).

  #. Install Bash completions using the ` + "`homebrew`" + ` command:
  
     ::
     
        brew install bash-completion
  
  #. Update your Bash profile:
  
     ::
     
       echo '[[ -r "$(brew --prefix)/etc/profile.d/bash_completion.sh" ]] && . "$(brew --prefix)/etc/profile.d/bash_completion.sh"' >> ~/.bash_profile
  
  #. Run the following command to install auto completion:
  
     ::
     
       confluent completion bash > $(brew --prefix)/etc/bash_completion.d/confluent

Install Bash completions on Linux:
  #.  Install Bash completion:

      ::

        sudo apt-get install bash-completion

  #. Set the ` + "`confluent completion`" + ` code for Bash to a file that's sourced on login:
  
     ::
     
        confluent completion bash > /etc/bash_completion.d/confluent

  #. Load the ` + "`confluent completion`" + ` code for Bash into the current shell:
  
     ::
  
        source /etc/bash_completion.d/confluent

  #. Add the source command above to your ` + "`~/.bashrc`" + ` or ` + "`~/.bash_profile`" + ` to enable completions for new terminals.

Install Zsh completions:
  Zsh looks for completion functions in the directories listed in the ` + "`fpath`" + ` shell variable.

  #. Put the ` + "`confluent completion`" + ` code for Zsh into a file in one the ` + "`fpath`" + ` directories, preferably one of the functions directories. For example:

     ::

        confluent completion zsh > ${fpath[1]}/_confluent

  #. Enable Zsh completions:
  
     ::
     
        autoload -U compinit && compinit

  #. Add the autoload command in your ` + "`~/.zshrc`" + ` to enable completions for new terminals. If you encounter error messages about insecure files, you can resolve by running the ` + "`chown`" + ` command to change the ` + "`_confluent`" + ` file to the same ` + "`user:group`" + ` as the other files in ` + "`${fpath[1]}/`" + `.

  #. To update your completion scripts after updating the CLI, run ` + "`confluent completion <bash|zsh>`" + ` again and overwrite the file initially created above.
`

type completionCommand struct {
	*cobra.Command
	rootCmd *cobra.Command
}

// New returns the Cobra command for shell completion.
func New(rootCmd *cobra.Command) *cobra.Command {
	cmd := &completionCommand{
		rootCmd: rootCmd,
	}
	cmd.init()
	return cmd.Command
}

func (c *completionCommand) init() {
	c.Command = &cobra.Command{
		Use:   "completion <shell>",
		Short: "Print shell completion code.",
		Long:  longDescription,
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.completion),
	}
}

func (c *completionCommand) completion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return c.rootCmd.GenBashCompletion(cmd.OutOrStdout())
	case "zsh":
		return c.rootCmd.GenZshCompletion(cmd.OutOrStdout())
	default:
		return fmt.Errorf(errors.UnsupportedShellErrorMsg, args[0])
	}
}
