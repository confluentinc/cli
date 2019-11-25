package cmd

import (
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/confluentinc/ccloud-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/analytics"
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
	Analytics    analytics.Client
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

		err := r.validateAuthToken(cmd)
		switch err.(type) {
		case *ccloud.ExpiredTokenError:
			err := r.Config.DeleteUserAuth()
			if err != nil {
				return err
			}
			ErrPrintln(cmd, "Your token has expired. You are now logged out.")
			analyticsError := r.Analytics.SessionTimedOut()
			if analyticsError != nil {
				r.Logger.Debug(analyticsError.Error())
			}
		}
		r.Analytics.TrackCommand(cmd, args)
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
			if err := r.validateAuthToken(cmd); err != nil {
				return errors.HandleCommon(err, cmd)
			}
		}
		return nil
	}
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous()(cmd, args); err != nil {
			return err
		}
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


func (r *PreRun) validateAuthToken(cmd *cobra.Command) error {
	var claims map[string]interface{}
	token, err := jwt.ParseSigned(r.Config.AuthToken)
	if err != nil {
		return &ccloud.InvalidTokenError{}
	}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}
	if exp, ok := claims["exp"].(float64); ok {
		if float64(r.Clock.Now().Unix()) > exp {
			return &ccloud.ExpiredTokenError{}
		}
	}
	return nil
}
