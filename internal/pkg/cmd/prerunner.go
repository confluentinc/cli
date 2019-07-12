package cmd

import (
	"context"
	"fmt"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"gopkg.in/square/go-jose.v2/jwt"
	"os"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/config"
	config_pkg "github.com/confluentinc/cli/internal/pkg/config"
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

func getSrCredentials() (key string, secret string, err error) {
	prompt := NewPrompt(os.Stdin)
	fmt.Println("Enter your Schema Registry API Key:")
	key, err = prompt.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Println("Enter your Schema Registry API Secret:")
	secret, err = prompt.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	// Validate before returning
	_, _, err = srsdk.APIClient{}.DefaultApi.Get(context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: key,
		Password: secret,
	}))
	if err != nil {
		return "", "", errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}
	return strings.TrimSpace(key), strings.TrimSpace(secret), nil
}

func SrContext(config *config.Config) (context.Context, error) {
	if config.SrCredentials == nil || len(config.SrCredentials.Key) == 0 || len(config.SrCredentials.Secret) == 0 {
		key, secret, err := getSrCredentials()
		if err != nil {
			return nil, err
		}
		config.SrCredentials = &config_pkg.APIKeyPair{
			Key:    key,
			Secret: secret,
		}
		config.Save()
	}
	return context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: config.SrCredentials.Key,
		Password: config.SrCredentials.Secret,
	}), nil
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
