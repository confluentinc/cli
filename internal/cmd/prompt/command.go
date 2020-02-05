package prompt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/ps1"
)

const longDescriptionTemplate = `Use this command to add {{.CLIName}} information in your terminal prompt.

For Bash, you'll want to do something like this:

::

  $ export PS1="$({{.CLIName}} prompt) $PS1"

ZSH users should be aware that they will have to set the 'PROMPT_SUBST' option first:

::

  $ setopt prompt_subst
  $ export PS1="$({{.CLIName}} prompt) $PS1"

You can customize the prompt by calling passing a '--format' flag, such as '-f "{{.CLIName}}|%E:%K"'.
If you want to create a more sophisticated prompt (such as using the built-in color functions),
it'll be easiest for you if you use an environment variable rather than try to escape the quotes.

::

  $ export {{.CLIName | ToUpper}}_PROMPT_FMT='({{"{{"}}color "blue" "{{.CLIName}}"{{"}}"}}|{{"{{"}}color "red" "%E"{{"}}"}}:{{"{{"}}color "cyan" "%K"{{"}}"}})'
  $ export PS1="$({{.CLIName}} prompt -f "${{.CLIName | ToUpper}}_PROMPT_FMT") $PS1"

To make this permanent, you must add it to your bash or zsh profile.

Formats
~~~~~~~

'{{.CLIName}} prompt' comes with a number of formatting tokens. What follows is a list of all tokens:

* '%C' or {{"{{"}}.ContextName{{"}}"}}

  The name of the current context in use. E.g., "dev-app1", "stag-dc1", "prod"

* '%e' or {{"{{"}}.EnvironmentId{{"}}"}}

  The ID of the current environment in use. E.g., "a-4567"

* '%E' or {{"{{"}}.EnvironmentName{{"}}"}}

  The name of the current environment in use. E.g., "default", "prod-team1"

* '%k' or {{"{{"}}.KafkaClusterId{{"}}"}}

  The ID of the current Kafka cluster in use. E.g., "lkc-abc123"

* '%K' or {{"{{"}}.KafkaClusterName{{"}}"}}

  The name of the current Kafka cluster in use. E.g., "prod-us-west-2-iot"

* '%a' or {{"{{"}}.KafkaAPIKey{{"}}"}}

  The current Kafka API key in use. E.g., "ABCDEF1234567890"

* '%u' or {{"{{"}}.UserName{{"}}"}}

  The current user or credentials in use. E.g., "joe@montana.com"

Colors
~~~~~~

There are special functions used for controlling colors.

* {{"{{"}}color "<color>" "some text"{{"}}"}}
* {{"{{"}}fgcolor "<color>" "some text"{{"}}"}}
* {{"{{"}}bgcolor "<color>" "some text"{{"}}"}}
* {{"{{"}}colorattr "<attr>" "some text"{{"}}"}}

Available colors: black, red, green, yellow, blue, magenta, cyan, white
Available attributes: bold, underline, invert (swaps the fg/bg colors)

Examples:

* {{"{{"}}color "red" "some text" | colorattr "bold" | bgcolor "blue"{{"}}"}}
* {{"{{"}}color "red"{{"}}"}} some text here {{"{{"}}resetcolor{{"}}"}}

You can also mix format tokens and/or data in the same line
* {{"{{"}}color "cyan" "%E"{{"}}"}} {{"{{"}}color "blue" .KafkaClusterId{{"}}"}}

Notes:

* 'color' is just an alias of 'fgcolor'
* calling 'resetcolor' will reset all color attributes, not just the most recently set

You can disable color output by passing the flag '--no-color'.

`

// UX inspired by https://github.com/djl/vcprompt

type promptCommand struct {
	*cobra.Command
	config *v2.Config
	ps1    *ps1.Prompt
	logger *log.Logger
}

// NewPromptCmd returns the Cobra command for the PS1 prompt.
func NewPromptCmd(config *v2.Config, ps1 *ps1.Prompt, logger *log.Logger) *cobra.Command {
	cmd := &promptCommand{
		config: config,
		ps1:    ps1,
		logger: logger,
	}
	cmd.init()
	return cmd.Command
}

func (c *promptCommand) init() {
	c.Command = &cobra.Command{
		Use:   "prompt",
		Short: c.mustParseTemplate("Print {{.CLIName}} CLI context for your terminal prompt."),
		Long:  c.mustParseTemplate(longDescriptionTemplate),
		RunE:  c.prompt,
		Args:  cobra.NoArgs,
	}
	// Ideally we'd default to %c but contexts are implicit today with uber-verbose names like `login-cody@confluent.io-https://devel.cpdev.cloud`
	defaultFormat := `({{color "blue" "ccloud"}}|{{color "red" "%E"}}:{{color "cyan" "%K"}})`
	if c.config.CLIName == "confluent" {
		defaultFormat = `({{color "blue" "confluent"}}|{{color "cyan" "%K"}})`
	}
	c.Command.Flags().StringP("format", "f", defaultFormat, "The format string to use. See the help for details.")
	c.Command.Flags().BoolP("no-color", "g", false, "Do not include ANSI color codes in the output.")
	c.Command.Flags().StringP("timeout", "t", "200ms", "The maximum execution time in milliseconds.")
	c.Command.Flags().SortFlags = false
}

// Output context about the current CLI config suitable for a PS1 prompt.
// It allows custom user formatting the configuration by parsing format flags.
func (c *promptCommand) prompt(cmd *cobra.Command, args []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return err
	}
	color.NoColor = noColor // we must set this, otherwise prints colors only to terminals (i.e., not for a PS1 prompt)

	t, err := cmd.Flags().GetString("timeout")
	if err != nil {
		return err
	}
	timeout, err := time.ParseDuration(t)
	if err != nil {
		di, err := strconv.Atoi(t)
		if err != nil {
			return fmt.Errorf(`invalid argument "%s" for "-t, --timeout" flag: unable to parse %s as duration or milliseconds`, t, t)
		}
		timeout = time.Duration(di) * time.Millisecond
	}

	// Parse in a background goroutine so we can set a timeout
	retCh := make(chan string)
	errCh := make(chan error)
	go func() {
		prompt, err := c.ps1.Get(format)
		if err != nil {
			errCh <- err
			return
		}
		retCh <- prompt
	}()

	// Wait for parse results, error, or timeout
	select {
	case prompt := <-retCh:
		pcmd.Println(cmd, prompt)
	case err := <-errCh:
		c.Command.SilenceUsage = true
		return errors.Wrapf(err, `error parsing prompt format string "%s"`, format)
	case <-time.After(timeout):
		// log the timeout and just print nothing
		c.logger.Warnf("timed out after %s", timeout)
		return nil
	}

	return nil
}

// mustParseTemplate will panic if text can't be parsed or executed
// don't call with user-provided text!
func (c *promptCommand) mustParseTemplate(text string) string {
	t, err := c.ps1.ParseTemplate(text)
	if err != nil {
		panic(err)
	}
	return t
}
