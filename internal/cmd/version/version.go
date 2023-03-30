package version

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*pcmd.CLICommand
	ver *pversion.Version
}

func New(prerunner pcmd.PreRunner, ver *pversion.Version) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Show version of the %s.", pversion.FullCLIName),
		Args:  cobra.NoArgs,
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		ver:        ver,
	}
	cmd.RunE = c.version

	return cmd
}

func (c *command) version(_ *cobra.Command, _ []string) error {
	output.Println(c.ver)
	return nil
}
