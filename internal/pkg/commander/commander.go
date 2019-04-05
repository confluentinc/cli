package commander

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/terminal"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
)

type Commander interface {
	Anonymous() func(cmd *cobra.Command, args []string) error
	Authenticated() func(cmd *cobra.Command, args []string) error
}

// PreRunner is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner struct {
	UpdateClient update.Client
	CLIName      string
	Version      string
	Logger       *log.Logger
	Config       *config.Config
	Prompt       terminal.Prompt
}

func (r *PreRunner) Anonymous() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r.Prompt.SetOutput(cmd.OutOrStdout())
		if err := log.SetLoggingVerbosity(cmd, r.Logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.notifyIfUpdateAvailable(r.CLIName, r.Version); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}

func (r *PreRunner) Authenticated() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous()(cmd, args); err != nil {
			return err
		}
		if err := r.Config.CheckLogin(); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRunner) notifyIfUpdateAvailable(name string, currentVersion string) error {
	updateAvailable, _, err := r.UpdateClient.CheckForUpdates(name, currentVersion, false)
	if err != nil {
		return err
	}
	if updateAvailable {
		msg := "Updates are available for %s. To install them, please run:\n$ %s update\n\n"
		fmt.Fprintf(os.Stderr, msg, name, name)
	}
	return nil
}
