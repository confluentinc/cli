package ps1

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// NewCobraCommand returns a Cobra CLI command named `prompt`, which can be inserted into a user's `$PS1` variable.
func (p *ps1) NewCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "prompt",
		Short:        p.short(),
		Long:         p.long(),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}

			timeout, err := cmd.Flags().GetInt("timeout")
			if err != nil {
				return err
			}

			cmd.Print(p.prompt(format, timeout))
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", fmt.Sprintf("(%s)", p.cliName), "The format string to use. See the help for details.")
	cmd.Flags().IntP("timeout", "t", 200, "The maximum execution time in milliseconds.")

	cmd.SetOut(os.Stdout)

	return cmd
}

func (p *ps1) short() string {
	return fmt.Sprintf("Add %s context to your terminal prompt.", p.cliName)
}

func (p *ps1) long() string {
	rows := make([]string, len(p.tokens))
	for i, token := range p.tokens {
		rows[i] = fmt.Sprintf("* %v - %s", token, token.Desc)
	}

	tokens := strings.Join(rows, "\n")
	return fmt.Sprintf(fmtLongDescription, p.cliName, p.cliName, p.cliName, p.cliName, tokens, p.cliName)
}

func (p *ps1) prompt(format string, timeoutMs int) string {
	timeout := time.Duration(timeoutMs) * time.Millisecond

	vals := make(chan string)
	errs := make(chan error)

	go func() {
		if val, err := p.write(format); err != nil {
			errs <- err
		} else {
			vals <- val
		}
	}()

	select {
	case val := <-vals:
		return val
	case err := <-errs:
		x := strings.Split(err.Error(), ": ")
		return fmt.Sprintf("(%s|%s)", p.cliName, x[len(x)-1])
	case <-time.After(timeout):
		return fmt.Sprintf("(%s)", p.cliName)
	}
}

func (p *ps1) write(format string) (string, error) {
	for _, token := range p.tokens {
		format = strings.ReplaceAll(format, token.String(), token.Func())
	}

	tmpl, err := template.New("ps1").Funcs(colorFuncs).Parse(format)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, nil)
	return buf.String(), err
}
