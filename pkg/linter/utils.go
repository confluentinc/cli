package linter

import (
	"strings"

	"github.com/spf13/cobra"
)

func FullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Use}
	cmd.VisitParents(func(cmd *cobra.Command) {
		use = append([]string{cmd.Use}, use...)
	})
	return strings.Join(use, " ")
}
