package cmd

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
)

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous() func(cmd *cobra.Command, args []string) error
	Authenticated() func(cmd *cobra.Command, args []string) error
	AuthenticatedAPIKey() func(cmd *cobra.Command, args []string) error
}

// PreRun is the standard PreRunner implementation
type PreRun struct {
	UpdateClient update.Client
	CLIName      string
	Version      string
	Logger       *log.Logger
	Config       *config.Config
	ConfigHelper *ConfigHelper
}

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := log.SetLoggingVerbosity(cmd, r.Logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.notifyIfUpdateAvailable(cmd, r.CLIName, r.Version); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}

// Authenticated provides PreRun operations for commands that require a logged-in user
func (r *PreRun) Authenticated() func(cmd *cobra.Command, args []string) error {
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

// AuthenticatedAPIKey provides PreRun operations for commands that require a logged-in user with an API key
func (r *PreRun) AuthenticatedAPIKey() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Authenticated()(cmd, args); err != nil {
			return err
		}
		cluster, err := GetKafkaCluster(cmd, r.ConfigHelper)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		err = r.Config.CheckHasAPIKey(cluster.Id)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, name string, currentVersion string) error {
	updateAvailable, _, err := r.UpdateClient.CheckForUpdates(name, currentVersion, false)
	if err != nil {
		return err
	}
	if updateAvailable {
		msg := "Updates are available for %s. To install them, please run:\n$ %s update\n\n"
		ErrPrintf(cmd, msg, name, name)
	}
	return nil
}
