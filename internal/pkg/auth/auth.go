package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	CCloudURL = "https://confluent.cloud"

	CCloudEmailEnvVar       = "CCLOUD_EMAIL"
	ConfluentUsernameEnvVar = "CONFLUENT_USERNAME"
	CCloudPasswordEnvVar    = "CCLOUD_PASSWORD"
	ConfluentPasswordEnvVar = "CONFLUENT_PASSWORD"

	CCloudEmailDeprecatedEnvVar       = "XX_CCLOUD_EMAIL"
	ConfluentUsernameDeprecatedEnvVar = "XX_CONFLUENT_USERNAME"
	CCloudPasswordDeprecatedEnvVar    = "XX_CCLOUD_PASSWORD"
	ConfluentPasswordDeprecatedEnvVar = "XX_CONFLUENT_PASSWORD"
)

func PersistLogoutToConfig(config *v3.Config) error {
	ctx := config.Context()
	if ctx == nil {
		return nil
	}
	err := ctx.DeleteUserAuth()
	if err != nil {
		return err
	}
	ctx.Config.CurrentContext = ""
	return config.Save()
}

func PersistConfluentLoginToConfig(config *v3.Config, username string, url string, token string, caCertPath string) error {
	state := &v2.ContextState{
		Auth:      nil,
		AuthToken: token,
	}
	return addOrUpdateContext(config, username, url, state, caCertPath)
}

func PersistCCloudLoginToConfig(config *v3.Config, email string, url string, token string, client *ccloud.Client) (*orgv1.Account, error) {
	state, err := getCCloudContextState(config, email, url, token, client)
	if err != nil {
		return nil, err
	}
	err = addOrUpdateContext(config, email, url, state, "")
	if err != nil {
		return nil, err
	}
	return state.Auth.Account, nil
}

func addOrUpdateContext(config *v3.Config, username string, url string, state *v2.ContextState, caCertPath string) error {
	ctxName := GenerateContextName(username, url)
	credName := generateCredentialName(username)
	platform := &v2.Platform{
		Name:       strings.TrimPrefix(url, "https://"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	credential := &v2.Credential{
		Name:     credName,
		Username: username,
		// don't save password if they entered it interactively.
	}
	err := config.SavePlatform(platform)
	if err != nil {
		return err
	}
	err = config.SaveCredential(credential)
	if err != nil {
		return err
	}
	if ctx, ok := config.Contexts[ctxName]; ok {
		config.ContextStates[ctxName] = state
		ctx.State = state

		ctx.Platform = platform
		ctx.PlatformName = platform.Name

		ctx.Credential = credential
		ctx.CredentialName = credential.Name
	} else {
		err = config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{},
			"", nil, state)
	}
	if err != nil {
		return err
	}
	err = config.SetContext(ctxName)
	if err != nil {
		return err
	}
	return nil
}

func getCCloudContextState(config *v3.Config, email string, url string, token string, client *ccloud.Client) (*v2.ContextState, error) {
	ctxName := GenerateContextName(email, url)
	user, err := getCCloudUser(token, client)
	if err != nil {
		return nil, err
	}
	var state *v2.ContextState
	ctx, err := config.FindContext(ctxName)
	if err == nil {
		state = ctx.State
	} else {
		state = new(v2.ContextState)
	}
	state.AuthToken = token

	if state.Auth == nil {
		state.Auth = &v1.AuthConfig{}
	}

	// Always overwrite the user, organization, and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	state.Auth.User = user.User
	state.Auth.Accounts = user.Accounts
	state.Auth.Organization = user.Organization

	// Default to 0th environment if no suitable environment is already configured
	hasGoodEnv := false
	if state.Auth.Account != nil {
		for _, acc := range state.Auth.Accounts {
			if acc.Id == state.Auth.Account.Id {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		state.Auth.Account = state.Auth.Accounts[0]
	}

	return state, nil
}

func getCCloudUser(token string, client *ccloud.Client) (*orgv1.GetUserReply, error) {
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return nil, err
	}
	if len(user.Accounts) == 0 {
		return nil, errors.Errorf(errors.NoEnvironmentFoundErrorMsg)
	}
	return user, nil
}

func GenerateContextName(username string, url string) string {
	return fmt.Sprintf("login-%s-%s", username, url)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}
