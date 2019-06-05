package ps1

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/ps1"
)

const longDescriptionTemplate = `Use this command to add {{.CLIName}} information in
your PS1 prompt.

For Bash, you'll want to do something like this:

::

  $ export PS1=${PS1}'$({{.CLIPath}} ps1)\n\$ '

ZSH users should be aware that they will have to set the 'PROMPT_SUBST'' option first:

::

  $ setopt prompt_subst
  $ export PS1='%n@%m:%~ $({{.CLIPath}} ps1)$ '

Formats
~~~~~~~

'{{.CLIName}} ps1' comes with a number of formatting tokens. What follows is a list of all tokens:

* '%c'

  The name of the current context in use. E.g., "dev-app1", "stag-dc1", "prod"

* '%e'

  The ID of the current environment in use. E.g., "a-4567"

* '%E'

  The name of the current environment in use. E.g., "default", "prod-team1"

* '%k'

  The ID of the current Kafka cluster in use. E.g., "lkc-abc123"

* '%K'

  The name of the current Kafka cluster in use. E.g., "prod-us-west-2-iot"

* '%a'

  The current Kafka API key in use. E.g., "ABCDEF1234567890"

* '%u'

  The current user or credentials in use. E.g., "cody@confluent.io"

Colors
~~~~~~

Unless you pass '--no-color', the output will be colored heuristically based on
the context, environment, and Kafka cluster names. E.g., if the environment contains "prod",
the output will be red; "dev" will be green; "stag" be yellow; everything else uncolored.

`

type ps1Command struct {
	*cobra.Command
	config *config.Config
	ps1    *ps1.Prompt
	logger *log.Logger
}

// NewPS1Cmd returns the Cobra command for the PS1 prompt.
func NewPS1Cmd(config *config.Config, ps1 *ps1.Prompt, logger *log.Logger) *cobra.Command {
	cmd := &ps1Command{
		config: config,
		ps1:    ps1,
		logger: logger,
	}
	cmd.init()
	return cmd.Command
}

func (c *ps1Command) init() {
	c.Command = &cobra.Command{
		Use:   "ps1",
		Short: "Print a prompt-string suitable for the PS1 prompt.",
		Long:  parseTemplate(c.config.CLIName, longDescriptionTemplate),
		RunE:  c.prompt,
		Args:  cobra.NoArgs,
	}
	// Ideally we'd default to ({{.CLIName}} %c) but contexts are implicit today with uber-verbose names like `login-cody@confluent.io-https://devel.cpdev.cloud`
	c.Command.Flags().StringP("format", "f", "(%E %K)", "The format string to use.")
	c.Command.Flags().BoolP("no-color", "g", false, "Do not colorize output based on the inferred environment (prod=red, stag=yellow, devel=green, unknown=none).")
	c.Command.Flags().StringP("timeout", "t", "200ms", "The maximum execution time in milliseconds.")
	c.Command.Flags().SortFlags = false
}

// Output context about the current CLI config suitable for a PS1 prompt.
// It allows custom user formatting the configuration by parsing format flags.
func (c *ps1Command) prompt(cmd *cobra.Command, args []string) error {
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
		prompt = parseTemplate(c.config.CLIName, prompt)

		envColor, err := c.ps1.InferEnvironmentColor()
		if err != nil {
			errCh <- err
			return
		}

		retCh <- envColor(prompt)
	}()

	// Wait for parse results, error, or timeout
	select {
	case prompt := <-retCh:
		pcmd.Println(cmd, prompt)
	case err := <-errCh:
		return err
	case <-time.After(timeout):
		// log the timeout and just print nothing
		c.logger.Warnf("timed out after %s", timeout)
		return nil
	}

	return nil
}

func parseTemplate(cliName, text string) string {
	t := template.Must(template.New("tmpl").Parse(text))
	buf := new(bytes.Buffer)
	data := map[string]interface{}{"CLIName": cliName, "CLIPath": filepath.Clean(os.Args[0])}
	if err := t.Execute(buf, data); err != nil {
		// We're okay with this since its definitely a development error; should never happen to users
		panic(err)
	}
	return buf.String()
}
