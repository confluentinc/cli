package version

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	c.RunE = pcmd.NewCLIRunE(c.version)

	return c.Command
}

func (c *command) version(cmd *cobra.Command, _ []string) error {
	utils.Println(cmd, c.ver)
	return nil
}
