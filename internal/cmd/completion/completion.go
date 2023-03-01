package completion

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

const longDescription = `Use this command to print the shell completion
code for the specified shell (Bash/Zsh only). The shell code must be evaluated to provide
interactive completion of ` + "`confluent`" + ` commands.

Install Bash completions on macOS:
  #. Install Homebrew (https://brew.sh/).

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
`

func New() *cobra.Command {
	return &cobra.Command{
		Use:       "completion <bash|zsh>",
		Short:     "Print shell completion code.",
		Long:      longDescription,
		ValidArgs: []string{"bash", "zsh"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := completion(cmd.Root(), args[0])
			if err != nil {
				return err
			}

			utils.Println(out)
			return nil
		},
	}
}

func completion(root *cobra.Command, shell string) (string, error) {
	buf := new(bytes.Buffer)

	if shell == "zsh" {
		err := root.GenZshCompletion(buf)
		return "#compdef confluent\n" + strings.TrimPrefix(buf.String(), "#"), err
	} else {
		err := root.GenBashCompletionV2(buf, true)
		return buf.String(), err
	}
}
