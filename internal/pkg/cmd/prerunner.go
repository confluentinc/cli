package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
)

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous() func(cmd *cobra.Command, args []string) error
	Authenticated() func(cmd *cobra.Command, args []string) error
	HasAPIKey() func(cmd *cobra.Command, args []string) error
}

// PreRun is the standard PreRunner implementation
type PreRun struct {
	UpdateClient update.Client
	CLIName      string
	Version      string
	Logger       *log.Logger
	Config       *config.Config
	ConfigHelper *ConfigHelper
	Clock        clockwork.Clock
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
		if r.Config.AuthToken != "" {
			// Validate token (not expired)
			var claims map[string]interface{}
			token, err := jwt.ParseSigned(r.Config.AuthToken)
			if err != nil {
				return errors.HandleCommon(&ccloud.InvalidTokenError{}, cmd)
			}
			if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
				return errors.HandleCommon(err, cmd)
			}
			if exp, ok := claims["exp"].(float64); ok {
				if float64(r.Clock.Now().Unix()) > exp {
					return errors.HandleCommon(&ccloud.ExpiredTokenError{}, cmd)
				}
			}
		}
		return nil
	}
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		context, err := r.Config.Context()
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		clusterId := context.Kafka
		return errors.HandleCommon(r.Config.CheckHasAPIKey(clusterId), cmd)
	}
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, name string, currentVersion string) error {
	updateAvailable, _, err := r.UpdateClient.CheckForUpdates(name, currentVersion, false)
	if err != nil {
		// This is a convenience helper to check-for-updates before arbitrary commands. Since the CLI supports running
		// in internet-less environments (e.g., local or on-prem deploys), swallow the error and log a warning.
		r.Logger.Warn(err)
		return nil
	}
	if updateAvailable {
		msg := "Updates are available for %s. To install them, please run:\n$ %s update\n\n"
		ErrPrintf(cmd, msg, name, name)
	}
	return nil
}
