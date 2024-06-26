//go:generate mocker --dst ../../mock/login_credentials_manager.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 login_credentials_manager.go LoginCredentialsManager --prefix ""
package auth

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v3/pkg/auth/sso"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/jwt"
	"github.com/confluentinc/cli/v3/pkg/keychain"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/secret"
)

const stopNonInteractiveMsg = "remove these credentials or use the `--prompt` flag to bypass non-interactive login"

type Credentials struct {
	Username string
	Password string
	IsSSO    bool
	Salt     []byte
	Nonce    []byte

	AuthToken        string
	AuthRefreshToken string

	// Only for Confluent Prerun login
	PrerunLoginURL        string
	PrerunLoginCaCertPath string
}

func (c *Credentials) IsFullSet() bool {
	return c.Username != "" && (c.IsSSO || c.Password != "" || c.AuthRefreshToken != "")
}

type environmentVariables struct {
	username string
	password string
}

// Get login credentials using the functions from LoginCredentialsManager
// Functions are called in order and credentials are returned right away if found from a function without attempting the other functions
func GetLoginCredentials(credentialsFuncs ...func() (*Credentials, error)) (*Credentials, error) {
	var credentials *Credentials
	var err error
	for _, credentialsFunc := range credentialsFuncs {
		credentials, err = credentialsFunc()
		if err == nil && credentials != nil && credentials.IsFullSet() {
			return credentials, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf(errors.NoCredentialsFoundErrorMsg)
}

type LoginCredentialsManager interface {
	GetCloudCredentialsFromEnvVar(string) func() (*Credentials, error)
	GetOnPremCredentialsFromEnvVar() func() (*Credentials, error)
	GetSsoCredentialsFromConfig(*config.Config, string) func() (*Credentials, error)
	GetCredentialsFromConfig(*config.Config, config.MachineParams) func() (*Credentials, error)
	GetCredentialsFromKeychain(bool, string, string) func() (*Credentials, error)
	GetOnPremSsoCredentials(url, caCertPath string, unsafeTrace bool) func() (*Credentials, error)
	GetOnPremSsoCredentialsFromConfig(*config.Config, bool) func() (*Credentials, error)
	GetCloudCredentialsFromPrompt(string) func() (*Credentials, error)
	GetOnPremCredentialsFromPrompt() func() (*Credentials, error)

	GetPrerunCredentialsFromConfig(*config.Config) func() (*Credentials, error)
	GetOnPremPrerunCredentialsFromEnvVar() func() (*Credentials, error)

	// Needed SSO login for non-prod accounts
	SetCloudClient(*ccloudv1.Client)
}

type LoginCredentialsManagerImpl struct {
	prompt form.Prompt
	client *ccloudv1.Client
}

func NewLoginCredentialsManager(prompt form.Prompt, client *ccloudv1.Client) LoginCredentialsManager {
	return &LoginCredentialsManagerImpl{
		prompt: prompt,
		client: client,
	}
}

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromEnvVar(organizationId string) func() (*Credentials, error) {
	envVars := environmentVariables{
		username: ConfluentCloudEmail,
		password: ConfluentCloudPassword,
	}
	return h.getCredentialsFromEnvVarFunc(envVars, organizationId)
}

func (h *LoginCredentialsManagerImpl) getCredentialsFromEnvVarFunc(envVars environmentVariables, organizationId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		email, password := h.getEnvVarCredentials(envVars.username, envVars.password)
		if h.isSSOUser(email, organizationId) {
			log.CliLogger.Debugf("%s=%s belongs to an SSO user.", ConfluentCloudEmail, email)
			return &Credentials{Username: email, IsSSO: true}, nil
		}

		if password == "" {
			log.CliLogger.Debug("Did not find full credential set from environment variables")
			return nil, nil
		}

		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getEnvVarCredentials(userEnvVar, passwordEnvVar string) (string, string) {
	username := os.Getenv(userEnvVar)
	if username == "" {
		return "", ""
	}
	password := os.Getenv(passwordEnvVar)
	if password == "" {
		return username, ""
	}
	log.CliLogger.Warnf(`Found credentials for user "%s" from environment variables "%s" and "%s" (%s)`, username, userEnvVar, passwordEnvVar, stopNonInteractiveMsg)
	return username, password
}

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromEnvVar() func() (*Credentials, error) {
	envVars := environmentVariables{
		username: ConfluentPlatformUsername,
		password: ConfluentPlatformPassword,
	}
	return h.getCredentialsFromEnvVarFunc(envVars, "")
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromConfig(cfg *config.Config, filterParams config.MachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		var loginCredential *config.LoginCredential
		ctx := cfg.Context()
		if ctx == nil {
			for _, item := range cfg.SavedCredentials {
				if matchLoginCredentialWithFilter(item, filterParams) {
					loginCredential = item
				}
			}
		} else if matchLoginCredentialWithFilter(cfg.SavedCredentials[ctx.Name], filterParams) {
			loginCredential = cfg.SavedCredentials[ctx.Name]
		}

		if loginCredential == nil {
			return nil, nil
		}

		password, err := secret.Decrypt(loginCredential.Username, loginCredential.EncryptedPassword, loginCredential.Salt, loginCredential.Nonce)
		if err != nil {
			return nil, err
		}

		credentials := &Credentials{
			Username: loginCredential.Username,
			Password: password,
			Salt:     loginCredential.Salt,
			Nonce:    loginCredential.Nonce,
		}
		return credentials, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetSsoCredentialsFromConfig(cfg *config.Config, url string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		ctx := cfg.Context()

		if ctx.GetPlatformServer() != url {
			return nil, nil
		}

		credentials := &Credentials{
			IsSSO:            ctx.IsSso(),
			Username:         ctx.GetUser().GetEmail(),
			AuthToken:        ctx.GetAuthToken(),
			AuthRefreshToken: ctx.GetAuthRefreshToken(),
		}

		if credentials.IsSSO {
			return credentials, nil
		}

		return nil, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetPrerunCredentialsFromConfig(cfg *config.Config) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		ctx := cfg.Context()
		if ctx == nil {
			return nil, nil
		}

		credentials := &Credentials{
			IsSSO:            ctx.GetUser().GetAuthType() == ccloudv1.AuthType_AUTH_TYPE_SSO || ctx.GetUser().GetSocialConnection() != "",
			Username:         ctx.GetUser().GetEmail(),
			AuthToken:        ctx.GetAuthToken(),
			AuthRefreshToken: ctx.GetAuthRefreshToken(),
		}
		log.CliLogger.Tracef("credentials: %#v", credentials)

		return credentials, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetOnPremSsoCredentialsFromConfig(cfg *config.Config, unsafeTrace bool) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		ctx := cfg.Context()
		if ctx == nil {
			return nil, nil
		}

		url := ctx.GetPlatform().GetServer()
		caCertPath := ctx.GetPlatform().GetCaCertPath()

		// on-prem SSO login does not use a username or email
		// the sub claim is used in place of a username since it is a unique identifier
		subClaim, err := jwt.GetClaim(ctx.GetAuthToken(), "sub")
		if err != nil {
			return nil, nil
		}

		sub, ok := subClaim.(string)
		if !ok {
			return nil, nil
		}

		if GenerateContextName(sub, url, caCertPath) == ctx.Name {
			return &Credentials{
				Username:         sub,
				IsSSO:            h.isOnPremSSOUser(url, caCertPath, unsafeTrace),
				AuthToken:        ctx.GetAuthToken(),
				AuthRefreshToken: ctx.GetAuthRefreshToken(),
			}, nil
		}

		return nil, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetOnPremSsoCredentials(url, caCertPath string, unsafeTrace bool) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		// For on-prem SSO logins, the sub claim of the Confluent Token is used in place of the Username
		// A placeholder is used here since we don't have the token yet
		return &Credentials{
			Username: "placeholder",
			IsSSO:    h.isOnPremSSOUser(url, caCertPath, unsafeTrace),
		}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromPrompt(organizationId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		output.Println(false, "Enter your Confluent Cloud credentials:")
		email := h.promptForUser("Email")
		if h.isSSOUser(email, organizationId) {
			log.CliLogger.Debug("Entered email belongs to an SSO user.")
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		password := h.promptForPassword()
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromPrompt() func() (*Credentials, error) {
	return func() (*Credentials, error) {
		output.Println(false, "Enter your Confluent credentials:")
		username := h.promptForUser("Username")
		password := h.promptForPassword()
		return &Credentials{Username: username, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) promptForUser(userField string) string {
	f := form.New(form.Field{ID: userField, Prompt: userField})
	if err := f.Prompt(h.prompt); err != nil {
		return ""
	}

	return f.Responses[userField].(string)
}

func (h *LoginCredentialsManagerImpl) promptForPassword() string {
	passwordField := "Password"
	f := form.New(form.Field{ID: passwordField, Prompt: passwordField, IsHidden: true})
	if err := f.Prompt(h.prompt); err != nil {
		return ""
	}
	return f.Responses[passwordField].(string)
}

func (h *LoginCredentialsManagerImpl) isSSOUser(email, organizationId string) bool {
	if h.client == nil {
		return false
	}

	if email != "" && slices.Contains([]string{"prod-us-gov", "devel-us-gov", "infra-us-gov"}, sso.GetCCloudEnvFromBaseUrl(h.client.BaseURL)) {
		return true
	}

	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(h.client.BaseURL)
	log.CliLogger.Tracef("h.client.BaseURL: %s", h.client.BaseURL)
	log.CliLogger.Tracef("auth0ClientId: %s", auth0ClientId)
	req := &ccloudv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      auth0ClientId,
		OrgResourceId: organizationId,
	}
	res, err := h.client.User.LoginRealm(req)
	// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
	// email, we want to silently continue and give the illusion of password prompt.
	return err == nil && res.GetIsSso()
}

func (h *LoginCredentialsManagerImpl) isOnPremSSOUser(url, caCertPath string, unsafeTrace bool) bool {
	clientManager := &MDSClientManagerImpl{}
	client, err := clientManager.GetMDSClient(url, caCertPath, unsafeTrace)
	if err != nil {
		return false
	}

	featuresInfo, _, err := client.MetadataServiceOperationsApi.Features(context.Background())
	if err != nil {
		return false
	}
	return featuresInfo.Features["oidc.login.device.1.enabled"]
}

// Prerun login for Confluent has two extra environment variables settings: CONFLUENT_MDS_URL (required), CONFLUNET_CA_CERT_PATH (optional)
// Those two variables are passed as flags for login command, but for prerun logins they are required as environment variables.
// URL and certificate-authority-path (if exists) are returned in addition to username and password
func (h *LoginCredentialsManagerImpl) GetOnPremPrerunCredentialsFromEnvVar() func() (*Credentials, error) {
	return func() (*Credentials, error) {
		url := os.Getenv(ConfluentPlatformMDSURL)
		if url == "" {
			return nil, fmt.Errorf(errors.NoUrlEnvVarErrorMsg)
		}

		envVars := environmentVariables{
			username: ConfluentPlatformUsername,
			password: ConfluentPlatformPassword,
		}

		creds, _ := h.getCredentialsFromEnvVarFunc(envVars, "")()
		if creds == nil {
			return nil, fmt.Errorf(errors.NoCredentialsFoundErrorMsg)
		}
		creds.PrerunLoginURL = url
		creds.PrerunLoginCaCertPath = os.Getenv(ConfluentPlatformCertificateAuthorityPath)

		return creds, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromKeychain(isCloud bool, ctxName string, url string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		if runtime.GOOS == "darwin" {
			username, password, err := keychain.Read(isCloud, ctxName, url)
			if err == nil && password != "" {
				log.CliLogger.Debugf(`Found credentials for user "%s" from keychain (%s)`, username, stopNonInteractiveMsg)
				return &Credentials{Username: username, Password: password}, nil
			}
			return nil, fmt.Errorf("no matching credentials found in keychain")
		}
		return nil, fmt.Errorf("keychain not available on platforms other than darwin")
	}
}

func (h *LoginCredentialsManagerImpl) SetCloudClient(client *ccloudv1.Client) {
	h.client = client
}

func matchLoginCredentialWithFilter(loginCredential *config.LoginCredential, filterParams config.MachineParams) bool {
	if loginCredential == nil {
		return false
	}
	if loginCredential.Url != filterParams.URL {
		return false
	}
	if loginCredential.IsCloud != filterParams.IsCloud {
		return false
	}
	if filterParams.Name != "" && !strings.Contains(filterParams.Name, loginCredential.Username) {
		return false
	}
	return true
}
