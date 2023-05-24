package auth

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dghubble/sling"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/keychain"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/secret"
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
		output.ErrPrintf(errors.DeprecatedEnvVarWarningMsg, deprecated, current)
		return val
	}

	return ""
}

func PersistLogout(config *v1.Config) error {
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

func PersistConfluentLoginToConfig(config *v1.Config, credentials *Credentials, url, token, caCertPath string, isLegacyContext, save bool) error {
	username := credentials.Username
	state := &v1.ContextState{AuthToken: token}
	var ctxName string
	if isLegacyContext {
		ctxName = GenerateContextName(username, url, "")
	} else {
		ctxName = GenerateContextName(username, url, caCertPath)
	}
	return addOrUpdateContext(config, false, credentials, ctxName, url, state, caCertPath, "", save)
}

func PersistCCloudCredentialsToConfig(config *v1.Config, client *ccloudv1.Client, url string, credentials *Credentials, save bool) (string, *ccloudv1.Organization, error) {
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

func addOrUpdateContext(config *v1.Config, isCloud bool, credentials *Credentials, ctxName, url string, state *v1.ContextState, caCertPath, orgResourceId string, save bool) error {
	platform := &v1.Platform{
		Name:       strings.TrimPrefix(url, "https://"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	if err := config.SavePlatform(platform); err != nil {
		return err
	}

	credential := &v1.Credential{
		Name:     generateCredentialName(credentials.Username),
		Username: credentials.Username,
		// don't save password if they entered it interactively.
	}
	if err := config.SaveCredential(credential); err != nil {
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

		loginCredential := &v1.LoginCredential{
			IsCloud:           isCloud,
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
		ctx.LastOrgId = orgResourceId
	} else {
		if err := config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{}, "", nil, state, orgResourceId, ""); err != nil {
			return err
		}
	}

	return config.UseContext(ctxName)
}

func getCCloudContextState(token, refreshToken string, user *ccloudv1.GetMeReply) *v1.ContextState {
	return &v1.ContextState{
		Auth: &v1.AuthConfig{
			User:         user.GetUser(),
			Organization: user.GetOrganization(),
		},
		AuthToken:        token,
		AuthRefreshToken: refreshToken,
	}
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

func GetBearerToken(authenticatedState *v1.ContextState, server, clusterId string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"
	clusterIds := map[string][]string{"clusterIds": {clusterId}}

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	if _, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).BodyJSON(clusterIds).Post(accessTokenEndpoint).ReceiveSuccess(responses); err != nil {
		return "", err
	}
	return responses.Token, nil
}

func GetJwtTokenForV2Client(authenticatedState *v1.ContextState, server string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"

	responses := new(response)
	_, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).Body(strings.NewReader("{}")).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}
	if responses.Error != "" {
		return "", errors.New(responses.Error)
	}
	return responses.Token, nil
}
