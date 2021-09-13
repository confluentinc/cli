package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dghubble/sling"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	CCloudURL = "https://confluent.cloud"

	ConfluentCloudEmail         = "CONFLUENT_CLOUD_EMAIL"
	ConfluentCloudPassword      = "CONFLUENT_CLOUD_PASSWORD"
	ConfluentPlatformUsername   = "CONFLUENT_PLATFORM_USERNAME"
	ConfluentPlatformPassword   = "CONFLUENT_PLATFORM_PASSWORD"
	ConfluentPlatformMDSURL     = "CONFLUENT_PLATFORM_MDS_URL"
	ConfluentPlatformCACertPath = "CONFLUENT_PLATFORM_CA_CERT_PATH"

	DeprecatedConfluentCloudEmail         = "CCLOUD_EMAIL"
	DeprecatedConfluentCloudPassword      = "CCLOUD_PASSWORD"
	DeprecatedConfluentPlatformUsername   = "CONFLUENT_USERNAME"
	DeprecatedConfluentPlatformPassword   = "CONFLUENT_PASSWORD"
	DeprecatedConfluentPlatformMDSURL     = "CONFLUENT_MDS_URL"
	DeprecatedConfluentPlatformCACertPath = "CONFLUENT_CA_CERT_PATH"
)

// GetEnvWithFallback calls os.GetEnv() twice, once for the current var and once for the deprecated var.
func GetEnvWithFallback(current, deprecated string) string {
	if val := os.Getenv(current); val != "" {
		return val
	}

	if val := os.Getenv(deprecated); val != "" {
		_, _ = fmt.Fprintf(os.Stderr, errors.DeprecatedEnvVarWarningMsg, deprecated, current)
		return val
	}

	return ""
}

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

func PersistConfluentLoginToConfig(config *v3.Config, username string, url string, token string, caCertPath string, isLegacyContext bool) error {
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
	return addOrUpdateContext(config, ctxName, username, url, state, caCertPath)
}

func PersistCCloudLoginToConfig(config *v3.Config, email string, url string, token string, client *ccloud.Client) (*orgv1.Account, error) {
	ctxName := GenerateCloudContextName(email, url)
	state, err := getCCloudContextState(config, ctxName, email, url, token, client)
	if err != nil {
		return nil, err
	}
	err = addOrUpdateContext(config, ctxName, email, url, state, "")
	if err != nil {
		return nil, err
	}
	return state.Auth.Account, nil
}

func addOrUpdateContext(config *v3.Config, ctxName string, username string, url string, state *v2.ContextState, caCertPath string) error {
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

	if err := config.SavePlatform(platform); err != nil {
		return err
	}

	if err := config.SaveCredential(credential); err != nil {
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
		if err := config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{}, "", nil, state); err != nil {
			return err
		}
	}

	return config.UseContext(ctxName)
}

func getCCloudContextState(config *v3.Config, ctxName string, email string, url string, token string, client *ccloud.Client) (*v2.ContextState, error) {
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
