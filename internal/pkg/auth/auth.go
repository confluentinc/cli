package auth

import (
	"context"
	"fmt"
	"strings"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/dghubble/sling"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/secret"
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

	ConfluentURLEnvVar        = "CONFLUENT_MDS_URL"
	ConfluentCaCertPathEnvVar = "CONFLUENT_CA_CERT_PATH"
)

func PersistLogoutToConfig(config *v3.Config) error {
	ctx := config.Context()
	if ctx == nil {
		return nil
	}

	delete(ctx.Config.SavedCredentials, ctx.Name)

	err := ctx.DeleteUserAuth()
	if err != nil {
		return err
	}
	ctx.Config.CurrentContext = ""
	return config.Save()
}

func PersistConfluentLoginToConfig(config *v3.Config, credentials *Credentials, url string, token string, caCertPath string, isLegacyContext, save bool) error {
	username := credentials.Username
	state := &v2.ContextState{
		Auth:      nil,
		AuthToken: token,
	}
	var ctxName string
	if isLegacyContext {
		ctxName = GenerateContextName(username, url, "")
	} else {
		ctxName = GenerateContextName(username, url, caCertPath)
	}
	return addOrUpdateContext(config, credentials, ctxName, url, state, caCertPath, save)
}

func PersistCCloudLoginToConfig(config *v3.Config, credentials *Credentials, url string, token string, client *ccloud.Client, save bool) (*orgv1.Account, error) {
	ctxName := GenerateCloudContextName(credentials.Username, url)
	state, err := getCCloudContextState(config, ctxName, credentials.Username, url, token, client)
	if err != nil {
		return nil, err
	}
	err = addOrUpdateContext(config, credentials, ctxName, url, state, "", save)
	if err != nil {
		return nil, err
	}
	return state.Auth.Account, nil
}

func addOrUpdateContext(config *v3.Config, credentials *Credentials, ctxName, url string, state *v2.ContextState, caCertPath string, save bool) error {
	credName := generateCredentialName(credentials.Username)
	platform := &v2.Platform{
		Name:       strings.TrimPrefix(url, "https://"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	credential := &v2.Credential{
		Name:     credName,
		Username: credentials.Username,
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

	if save && !credentials.IsSSO {
		salt, err := secret.GenerateRandomBytes(secret.SaltLength)
		if err != nil {
			return err
		}
		nonce, err := secret.GenerateRandomBytes(secret.NonceLength)
		if err != nil {
			return err
		}

		encryptedPassword, err := secret.Encrypt(credentials.Username, credentials.Password, salt, nonce)
		if err != nil {
			return err
		}

		loginCredential := &v2.LoginCredential{
			Url:               url,
			Username:          credentials.Username,
			EncryptedPassword: encryptedPassword,
			Salt:              salt,
			Nonce:             nonce,
		}
		if err := config.SaveLoginCredential(ctxName, loginCredential); err != nil {
			return err
		}
	}

	stateSalt, err := secret.GenerateRandomBytes(secret.SaltLength)
	if err != nil {
		return err
	}
	stateNonce, err := secret.GenerateRandomBytes(secret.NonceLength)
	if err != nil {
		return err
	}

	state.Salt = stateSalt
	state.Nonce = stateNonce

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

func getCCloudContextState(config *v3.Config, ctxName string, token string, client *ccloud.Client) (*v2.ContextState, error) {
	user, err := getCCloudUser(client)
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

func getCCloudUser(client *ccloud.Client) (*flowv1.GetMeReply, error) {
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return nil, err
	}
	if len(user.Accounts) == 0 {
		return nil, errors.Errorf(errors.NoEnvironmentFoundErrorMsg)
	}
	return user, nil
}

func GenerateCloudContextName(username string, url string) string {
	return GenerateContextName(username, url, "")
}

// if CP users use cacertpath then include that in the context name
// (legacy CP users may still have context without cacertpath in the name but have cacertpath stored)
func GenerateContextName(username string, url string, caCertPath string) string {
	if caCertPath == "" {
		return fmt.Sprintf("login-%s-%s", username, url)
	}
	return fmt.Sprintf("login-%s-%s?cacertpath=%s", username, url, caCertPath)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

func GetRemoteAPIName(cliName string) string {
	if cliName == "ccloud" {
		return "Confluent Cloud"
	}
	return "Confluent Platform"
}

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func GetBearerToken(authenticatedState *v2.ContextState, server string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	_, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).Body(strings.NewReader("{}")).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}

	return responses.Token, nil
}
