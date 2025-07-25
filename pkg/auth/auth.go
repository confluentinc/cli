package auth

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dghubble/sling"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/keychain"
	"github.com/confluentinc/cli/v4/pkg/secret"
)

const (
	CCloudURL = "https://confluent.cloud"

	ConfluentCloudEmail                       = "CONFLUENT_CLOUD_EMAIL"
	ConfluentCloudPassword                    = "CONFLUENT_CLOUD_PASSWORD"
	ConfluentCloudOrganizationId              = "CONFLUENT_CLOUD_ORGANIZATION_ID"
	ConfluentPlatformUsername                 = "CONFLUENT_PLATFORM_USERNAME"
	ConfluentPlatformPassword                 = "CONFLUENT_PLATFORM_PASSWORD"
	ConfluentPlatformMDSURL                   = "CONFLUENT_PLATFORM_MDS_URL"
	ConfluentPlatformCertificateAuthorityPath = "CONFLUENT_PLATFORM_CERTIFICATE_AUTHORITY_PATH"
	ConfluentPlatformClientCertPath           = "CONFLUENT_PLATFORM_CLIENT_CERT_PATH"
	ConfluentPlatformClientKeyPath            = "CONFLUENT_PLATFORM_CLIENT_KEY_PATH"
	ConfluentPlatformSSO                      = "CONFLUENT_PLATFORM_SSO"

	// Confluent Platform CMF environment variables
	ConfluentPlatformCmfURL                      = "CONFLUENT_CMF_URL"
	ConfluentPlatformCmfClientKeyPath            = "CONFLUENT_CMF_CLIENT_KEY_PATH"
	ConfluentPlatformCmfClientCertPath           = "CONFLUENT_CMF_CLIENT_CERT_PATH"
	ConfluentPlatformCmfCertificateAuthorityPath = "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH"
)

func IsOnPremSSOEnv() bool {
	return strings.ToLower(os.Getenv(ConfluentPlatformSSO)) == "true"
}

func PersistLogout(config *config.Config) error {
	ctx := config.Context()
	if ctx == nil {
		return nil
	}

	if runtime.GOOS == "darwin" && !config.IsTest {
		if err := keychain.Delete(config.IsCloudLogin(), ctx.GetMachineName()); err != nil {
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

func PersistConfluentLoginToConfig(cfg *config.Config, credentials *Credentials, url, token, refreshToken, caCertPath string, save bool) error {
	if credentials.IsSSO || credentials.IsCertificateOnly {
		// on-prem SSO or mTLS certificate only login does not use a username or email
		// For SSO, the sub claim is used in place of a username since it is a unique identifier
		// For mTLS, the sub claim is the certificate SN, and we use the CN field as the username
		subClaim, err := jwt.GetClaim(token, "sub")
		if err != nil {
			return err
		}

		sub, ok := subClaim.(string)
		if !ok {
			return fmt.Errorf(errors.MalformedTokenErrorMsg, "sub")
		}

		credentials.Username = sub
	}
	username := credentials.Username

	state := &config.ContextState{
		AuthToken:        token,
		AuthRefreshToken: refreshToken,
	}

	ctxName := GenerateContextName(username, url, caCertPath)
	return addOrUpdateContext(cfg, false, credentials, ctxName, url, state, caCertPath, "", save, false)
}

func PersistCCloudCredentialsToConfig(config *config.Config, client *ccloudv1.Client, url string, credentials *Credentials, save bool) (string, *ccloudv1.Organization, error) {
	ctxName := GenerateCloudContextName(credentials.Username, url)

	user, err := client.Auth.User()
	if err != nil {
		return "", nil, err
	}

	state := getCCloudContextState(credentials.AuthToken, credentials.AuthRefreshToken, user)

	if err := addOrUpdateContext(config, true, credentials, ctxName, url, state, "", user.GetOrganization().GetResourceId(), save, credentials.IsMFA); err != nil {
		return "", nil, err
	}

	ctx := config.Context()
	// Need to reset CurrentSchemaRegistryEndpoint for every environment because this context is per environment per login
	for _, env := range ctx.Environments {
		env.CurrentSchemaRegistryEndpoint = ""
	}
	if err := config.Save(); err != nil {
		return "", nil, err
	}

	if ctx.CurrentEnvironment == "" && len(user.GetAccounts()) > 0 {
		ctx.SetCurrentEnvironment(user.GetAccounts()[0].GetId())
		if err := config.Save(); err != nil {
			return "", nil, err
		}
	}

	return ctx.CurrentEnvironment, user.GetOrganization(), nil
}

func addOrUpdateContext(cfg *config.Config, isCloud bool, credentials *Credentials, ctxName, url string, state *config.ContextState, caCertPath, organizationId string, save, isMFA bool) error {
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
		ctx.IsMFA = isMFA
	} else {
		if err := cfg.AddContext(ctxName, platform.Name, credential.Name, map[string]*config.KafkaClusterConfig{}, "", "", state, organizationId, "", isMFA); err != nil {
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

// CP users use cacertpath, so include that in the context name
func GenerateContextName(username, url, caCertPath string) string {
	if caCertPath == "" {
		return fmt.Sprintf("login-%s-%s", username, url)
	}
	return fmt.Sprintf("login-%s-%s?cacertpath=%s", username, url, caCertPath)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

func GetDataplaneToken(ctx *config.Context) (string, error) {
	endpoint := strings.TrimSuffix(ctx.GetPlatformServer(), "/") + "/api/access_tokens"

	res := &struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}{}

	if _, err := sling.New().Add("Content-Type", "application/json").Add("Authorization", "Bearer "+ctx.GetAuthToken()).Post(endpoint).BodyJSON(map[string]any{}).ReceiveSuccess(res); err != nil {
		return "", err
	}
	if res.Error != "" {
		return "", errors.New(res.Error)
	}
	return res.Token, nil
}
