package auth

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dghubble/sling"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/keychain"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/secret"
)

const (
	CCloudURL = "https://confluent.cloud"

	ConfluentCloudEmail          = "CONFLUENT_CLOUD_EMAIL"
	ConfluentCloudPassword       = "CONFLUENT_CLOUD_PASSWORD"
	ConfluentCloudOrganizationId = "CONFLUENT_CLOUD_ORGANIZATION_ID"
	ConfluentPlatformUsername    = "CONFLUENT_PLATFORM_USERNAME"
	ConfluentPlatformPassword    = "CONFLUENT_PLATFORM_PASSWORD"
	ConfluentPlatformMDSURL      = "CONFLUENT_PLATFORM_MDS_URL"
	ConfluentPlatformCACertPath  = "CONFLUENT_PLATFORM_CA_CERT_PATH"

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
		output.ErrPrintf(false, errors.DeprecatedEnvVarWarningMsg, deprecated, current)
		return val
	}

	return ""
}

func PersistLogout(config *config.Config) error {
	ctx := config.Context()
	if ctx == nil {
		return nil
	}

	if runtime.GOOS == "darwin" && !config.IsTest {
		if err := keychain.Delete(config.IsCloudLogin(), ctx.GetNetrcMachineName()); err != nil {
			return err
		}
	}

	delete(ctx.Config.SavedCredentials, ctx.Name)
	if err := ctx.DeleteUserAuth(); err != nil {
		return err
	}
	ctx.Config.CurrentContext = ""
	return config.Save()
}

func PersistConfluentLoginToConfig(cfg *config.Config, credentials *Credentials, url, token, caCertPath string, isLegacyContext, save bool) error {
	username := credentials.Username
	state := &config.ContextState{AuthToken: token}
	var ctxName string
	if isLegacyContext {
		ctxName = GenerateContextName(username, url, "")
	} else {
		ctxName = GenerateContextName(username, url, caCertPath)
	}
	return addOrUpdateContext(cfg, false, credentials, ctxName, url, state, caCertPath, "", save)
}

func PersistCCloudCredentialsToConfig(config *config.Config, client *ccloudv1.Client, url string, credentials *Credentials, save bool) (string, *ccloudv1.Organization, error) {
	ctxName := GenerateCloudContextName(credentials.Username, url)

	user, err := client.Auth.User()
	if err != nil {
		return "", nil, err
	}

	state := getCCloudContextState(credentials.AuthToken, credentials.AuthRefreshToken, user)

	if err := addOrUpdateContext(config, true, credentials, ctxName, url, state, "", user.GetOrganization().GetResourceId(), save); err != nil {
		return "", nil, err
	}

	ctx := config.Context()
	if ctx.CurrentEnvironment == "" && len(user.GetAccounts()) > 0 {
		ctx.SetCurrentEnvironment(user.GetAccounts()[0].GetId())
		if err := config.Save(); err != nil {
			return "", nil, err
		}
	}

	return ctx.CurrentEnvironment, user.GetOrganization(), nil
}

func addOrUpdateContext(cfg *config.Config, isCloud bool, credentials *Credentials, ctxName, url string, state *config.ContextState, caCertPath, organizationId string, save bool) error {
	platform := &config.Platform{
		Name:       strings.TrimSuffix(strings.TrimPrefix(url, "https://"), "/"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	if err := cfg.SavePlatform(platform); err != nil {
		return err
	}

	credential := &config.Credential{
		Name:     generateCredentialName(credentials.Username),
		Username: credentials.Username,
		// don't save password if they entered it interactively.
	}
	if err := cfg.SaveCredential(credential); err != nil {
		return err
	}

	if save && !credentials.IsSSO {
		salt, nonce, err := secret.GenerateSaltAndNonce()
		if err != nil {
			return err
		}
		encryptedPassword, err := secret.Encrypt(credentials.Username, credentials.Password, salt, nonce)
		if err != nil {
			return err
		}

		loginCredential := &config.LoginCredential{
			IsCloud:           isCloud,
			Url:               url,
			Username:          credentials.Username,
			EncryptedPassword: encryptedPassword,
			Salt:              salt,
			Nonce:             nonce,
		}
		if err := cfg.SaveLoginCredential(ctxName, loginCredential); err != nil {
			return err
		}
	}

	stateSalt, stateNonce, err := secret.GenerateSaltAndNonce()
	if err != nil {
		return err
	}
	state.Salt = stateSalt
	state.Nonce = stateNonce

	if ctx, ok := cfg.Contexts[ctxName]; ok {
		cfg.ContextStates[ctxName] = state
		ctx.State = state

		ctx.Platform = platform
		ctx.PlatformName = platform.Name

		ctx.Credential = credential
		ctx.CredentialName = credential.Name
		ctx.LastOrgId = organizationId
	} else {
		if err := cfg.AddContext(ctxName, platform.Name, credential.Name, map[string]*config.KafkaClusterConfig{}, "", state, organizationId, ""); err != nil {
			return err
		}
	}

	return cfg.UseContext(ctxName)
}

func getCCloudContextState(token, refreshToken string, user *ccloudv1.GetMeReply) *config.ContextState {
	return &config.ContextState{
		Auth: &config.AuthConfig{
			User:         user.GetUser(),
			Organization: user.GetOrganization(),
		},
		AuthToken:        token,
		AuthRefreshToken: refreshToken,
	}
}

func GenerateCloudContextName(username, url string) string {
	return GenerateContextName(username, url, "")
}

// if CP users use cacertpath then include that in the context name
// (legacy CP users may still have context without cacertpath in the name but have cacertpath stored)
func GenerateContextName(username, url, caCertPath string) string {
	if caCertPath == "" {
		return fmt.Sprintf("login-%s-%s", username, url)
	}
	return fmt.Sprintf("login-%s-%s?cacertpath=%s", username, url, caCertPath)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

func GetDataplaneToken(authenticatedState *config.ContextState, server string) (string, error) {
	endpoint := strings.Trim(server, "/") + "/api/access_tokens"

	res := &struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}{}

	if _, err := sling.New().Add("Content-Type", "application/json").Add("Authorization", "Bearer "+authenticatedState.AuthToken).Post(endpoint).BodyJSON(map[string]any{}).ReceiveSuccess(res); err != nil {
		return "", err
	}
	if res.Error != "" {
		return "", errors.New(res.Error)
	}
	return res.Token, nil
}
