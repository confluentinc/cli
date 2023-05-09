//go:generate mocker --dst ../../../mock/login_credentials_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_credentials_manager.go LoginCredentialsManager --prefix ""
package auth

import (
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/auth/sso"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/keychain"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/types"
)

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
	username           string
	password           string
	deprecatedUsername string
	deprecatedPassword string
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
	return nil, errors.New(errors.NoCredentialsFoundErrorMsg)
}

type LoginCredentialsManager interface {
	GetCloudCredentialsFromEnvVar(orgResourceId string) func() (*Credentials, error)
	GetOnPremCredentialsFromEnvVar() func() (*Credentials, error)
	GetSsoCredentialsFromConfig(cfg *v1.Config, url string) func() (*Credentials, error)
	GetCredentialsFromConfig(cfg *v1.Config, filterParams netrc.NetrcMachineParams) func() (*Credentials, error)
	GetCredentialsFromKeychain(cfg *v1.Config, isCloud bool, ctxName string, url string) func() (*Credentials, error)
	GetCredentialsFromNetrc(filterParams netrc.NetrcMachineParams) func() (*Credentials, error)
	GetCloudCredentialsFromPrompt(orgResourceId string) func() (*Credentials, error)
	GetOnPremCredentialsFromPrompt() func() (*Credentials, error)

	GetPrerunCredentialsFromConfig(cfg *v1.Config) func() (*Credentials, error)
	GetOnPremPrerunCredentialsFromEnvVar() func() (*Credentials, error)
	GetOnPremPrerunCredentialsFromNetrc(*cobra.Command, netrc.NetrcMachineParams) func() (*Credentials, error)

	// Needed SSO login for non-prod accounts
	SetCloudClient(client *ccloudv1.Client)
}

type LoginCredentialsManagerImpl struct {
	netrcHandler netrc.NetrcHandler
	prompt       form.Prompt
	client       *ccloudv1.Client
}

func NewLoginCredentialsManager(netrcHandler netrc.NetrcHandler, prompt form.Prompt, client *ccloudv1.Client) LoginCredentialsManager {
	return &LoginCredentialsManagerImpl{
		netrcHandler: netrcHandler,
		prompt:       prompt,
		client:       client,
	}
}

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromEnvVar(orgResourceId string) func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           ConfluentCloudEmail,
		password:           ConfluentCloudPassword,
		deprecatedUsername: DeprecatedConfluentCloudEmail,
		deprecatedPassword: DeprecatedConfluentCloudPassword,
	}
	return h.getCredentialsFromEnvVarFunc(envVars, orgResourceId)
}

func (h *LoginCredentialsManagerImpl) getCredentialsFromEnvVarFunc(envVars environmentVariables, orgResourceId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		email, password := h.getEnvVarCredentials(envVars.username, envVars.password)
		if h.isSSOUser(email, orgResourceId) {
			log.CliLogger.Debugf("%s=%s belongs to an SSO user.", ConfluentCloudEmail, email)
			return &Credentials{Username: email, IsSSO: true}, nil
		}

		if email == "" {
			email, password = h.getEnvVarCredentials(envVars.deprecatedUsername, envVars.deprecatedPassword)
			if email != "" {
				output.ErrPrintf(errors.DeprecatedEnvVarWarningMsg, envVars.deprecatedUsername, envVars.username)
			}
			if password != "" {
				output.ErrPrintf(errors.DeprecatedEnvVarWarningMsg, envVars.deprecatedPassword, envVars.password)
			}
		}

		if password == "" {
			log.CliLogger.Debug("Did not find full credential set from environment variables")
			return nil, nil
		}

		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getEnvVarCredentials(userEnvVar string, passwordEnvVar string) (string, string) {
	username := os.Getenv(userEnvVar)
	if len(username) == 0 {
		return "", ""
	}
	password := os.Getenv(passwordEnvVar)
	if len(password) == 0 {
		return username, ""
	}
	log.CliLogger.Warnf(errors.FoundEnvCredMsg, username, userEnvVar, passwordEnvVar)
	return username, password
}

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromEnvVar() func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           ConfluentPlatformUsername,
		password:           ConfluentPlatformPassword,
		deprecatedUsername: DeprecatedConfluentPlatformUsername,
		deprecatedPassword: DeprecatedConfluentPlatformPassword,
	}
	return h.getCredentialsFromEnvVarFunc(envVars, "")
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromConfig(cfg *v1.Config, filterParams netrc.NetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		var loginCredential *v1.LoginCredential
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

func (h *LoginCredentialsManagerImpl) GetSsoCredentialsFromConfig(cfg *v1.Config, url string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		ctx := cfg.Context()

		if ctx.GetPlatformServer() != url {
			return nil, nil
		}

		credentials := &Credentials{
			IsSSO:            ctx.GetUser().GetAuthType() == ccloudv1.AuthType_AUTH_TYPE_SSO || ctx.GetUser().GetSocialConnection() != "",
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

func (h *LoginCredentialsManagerImpl) GetPrerunCredentialsFromConfig(cfg *v1.Config) func() (*Credentials, error) {
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

func (h *LoginCredentialsManagerImpl) GetCredentialsFromNetrc(filterParams netrc.NetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(filterParams)
		if err != nil {
			log.CliLogger.Debugf("Get netrc machine error: %s", err.Error())
			return nil, err
		}

		log.CliLogger.Debugf(errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getNetrcMachine(filterParams netrc.NetrcMachineParams) (*netrc.Machine, error) {
	log.CliLogger.Debugf("Searching for netrc machine with filter: %+v", filterParams)
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil {
		return nil, err
	}
	if netrcMachine == nil {
		return nil, errors.Errorf("found no netrc machine using the filter: %+v", filterParams)
	}
	return netrcMachine, err
}

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromPrompt(orgResourceId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		output.Println("Enter your Confluent Cloud credentials:")
		email := h.promptForUser("Email")
		if h.isSSOUser(email, orgResourceId) {
			log.CliLogger.Debug("Entered email belongs to an SSO user.")
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		password := h.promptForPassword()
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromPrompt() func() (*Credentials, error) {
	return func() (*Credentials, error) {
		output.Println("Enter your Confluent credentials:")
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

func (h *LoginCredentialsManagerImpl) isSSOUser(email, orgId string) bool {
	if h.client == nil {
		return false
	}

	if email != "" && types.Contains([]string{"fedramp", "fedramp-internal"}, sso.GetCCloudEnvFromBaseUrl(h.client.BaseURL)) {
		return true
	}

	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(h.client.BaseURL)
	log.CliLogger.Tracef("h.client.BaseURL: %s", h.client.BaseURL)
	log.CliLogger.Tracef("auth0ClientId: %s", auth0ClientId)
	req := &ccloudv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      auth0ClientId,
		OrgResourceId: orgId,
	}
	res, err := h.client.User.LoginRealm(req)
	// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
	// email, we want to silently continue and give the illusion of password prompt.
	return err == nil && res.GetIsSso()
}

// Prerun login for Confluent has two extra environment variables settings: CONFLUENT_MDS_URL (required), CONFLUNET_CA_CERT_PATH (optional)
// Those two variables are passed as flags for login command, but for prerun logins they are required as environment variables.
// URL and ca-cert-path (if exists) are returned in addition to username and password
func (h *LoginCredentialsManagerImpl) GetOnPremPrerunCredentialsFromEnvVar() func() (*Credentials, error) {
	return func() (*Credentials, error) {
		url := GetEnvWithFallback(ConfluentPlatformMDSURL, DeprecatedConfluentPlatformMDSURL)
		if url == "" {
			return nil, errors.New(errors.NoURLEnvVarErrorMsg)
		}

		envVars := environmentVariables{
			username:           ConfluentPlatformUsername,
			password:           ConfluentPlatformPassword,
			deprecatedUsername: DeprecatedConfluentPlatformUsername,
			deprecatedPassword: DeprecatedConfluentPlatformPassword,
		}

		creds, _ := h.getCredentialsFromEnvVarFunc(envVars, "")()
		if creds == nil {
			return nil, errors.New(errors.NoCredentialsFoundErrorMsg)
		}
		creds.PrerunLoginURL = url
		creds.PrerunLoginCaCertPath = GetEnvWithFallback(ConfluentPlatformCACertPath, DeprecatedConfluentPlatformCACertPath)

		return creds, nil
	}
}

// Prerun login for Confluent will extract URL and ca-cert-path (if available) from the netrc machine name
// URL is no longer part of the filter and URL value will be of whichever URL the first context stored in netrc has
// URL and ca-cert-path (if exists) are returned in addition to username and password
func (h *LoginCredentialsManagerImpl) GetOnPremPrerunCredentialsFromNetrc(cmd *cobra.Command, netrcMachineParams netrc.NetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(netrcMachineParams)
		if err != nil {
			log.CliLogger.Debugf("Get netrc machine error: %s", err.Error())
			return nil, err
		}
		machineContextInfo, err := netrc.ParseNetrcMachineName(netrcMachine.Name)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Debugf(errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password, PrerunLoginURL: machineContextInfo.URL, PrerunLoginCaCertPath: machineContextInfo.CaCertPath}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromKeychain(cfg *v1.Config, isCloud bool, ctxName, url string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		if runtime.GOOS == "darwin" {
			username, password, err := keychain.Read(isCloud, ctxName, url)
			if err == nil && password != "" {
				return &Credentials{Username: username, Password: password}, nil
			}
			return nil, errors.New(errors.NoValidKeychainCredentialErrorMsg)
		}
		return nil, errors.New(errors.KeychainNotAvailableErrorMsg)
	}
}

func (h *LoginCredentialsManagerImpl) SetCloudClient(client *ccloudv1.Client) {
	h.client = client
}

func matchLoginCredentialWithFilter(loginCredential *v1.LoginCredential, filterParams netrc.NetrcMachineParams) bool {
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
