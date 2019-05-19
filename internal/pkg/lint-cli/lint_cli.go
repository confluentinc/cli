package lint_cli

import (
	"fmt"

	"github.com/client9/gospell"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Linter struct {
	Rules     []Rule
	FlagRules []FlagRule
	Vocab     *gospell.GoSpell
	Debug     bool
}

func (l *Linter) Lint(cmd *cobra.Command) error {
	var issues *multierror.Error

	err := l.lintRules(cmd)
	if err != nil {
		issues = multierror.Append(issues, err)
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := l.Lint(c); err != nil {
			issues = multierror.Append(issues, err)
		}
	}

	return issues.ErrorOrNil()
}

func (l *Linter) lintRules(cmd *cobra.Command) *multierror.Error {
	if l.Debug {
		fmt.Println(FullCommand(cmd))
		fmt.Println(cmd.Short)
		fmt.Println()
	}

	var issues *multierror.Error

	for _, rule := range l.Rules {
		issues = multierror.Append(issues, rule(cmd))
	}

	// check that flags are consistent
	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		for _, rule := range l.FlagRules {
			issues = multierror.Append(issues, rule(pf, cmd))
		}
	})
	return issues
}
