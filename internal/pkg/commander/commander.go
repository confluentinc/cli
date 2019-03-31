package prerunner

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
)

// PreRunner is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner struct {
	updateClient update.Client
	cliName string
	version string
	logger *log.Logger
	cfg *config.Config
}

func (r *PreRunner) Anonymous() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := log.SetLoggingVerbosity(cmd, r.logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.updateClient.NotifyIfUpdateAvailable(r.cliName, r.version); err != nil {
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
		if err := r.cfg.CheckLogin(); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}
