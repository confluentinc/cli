package auditlog

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type migrateCmd struct {
	*pcmd.CLICommand
}

func newMigrateCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "migrate",
		Short:       "Migrate legacy audit log configurations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &migrateCmd{pcmd.NewCLICommand(cmd, prerunner)}

	c.AddCommand(c.newConfigCommand())

	return c.Command
}
