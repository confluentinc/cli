package lint_cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func FullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Use}
	cmd.VisitParents(func(command *cobra.Command) {
		use = append([]string{command.Use}, use...)
	})
	return strings.Join(use, " ")
}
