//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/login_credentials_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_credentials_manager.go LoginCredentialsManager
package auth

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/keychain"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/sso"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Credentials struct {
	Username string
	Password string
	IsSSO    bool
	Salt     []byte
	Nonce    []byte

	AuthToken string

	// Only for Confluent Prerun login
	PrerunLoginURL        string
	PrerunLoginCaCertPath string
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
		if err == nil && credentials != nil && credentials.Username != "" {
			return credentials, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return nil, errors.New(errors.NoCredentialsFoundErrorMsg)
}

type LoginCredentialsManager interface {
	GetCloudCredentialsFromEnvVar(cmd *cobra.Command, orgResourceId string) func() (*Credentials, error)
	GetOnPremCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetSsoCredentialsFromConfig(cfg *v1.Config) func() (*Credentials, error)
	GetCredentialsFromConfig(cfg *v1.Config, filterParams netrc.NetrcMachineParams) func() (*Credentials, error)
	GetCredentialsFromKeychain(cfg *v1.Config, isCloud bool, ctxName string, url string) func() (*Credentials, error)
	GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.NetrcMachineParams) func() (*Credentials, error)
	GetCloudCredentialsFromPrompt(cmd *cobra.Command, orgResourceId string) func() (*Credentials, error)
	GetOnPremCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error)

	// Only for Confluent Prerun login
	GetPrerunCredentialsFromConfig(cfg *v1.Config) func() (*Credentials, error)
	GetOnPremPrerunCredentialsFromEnvVar(*cobra.Command) func() (*Credentials, error)
	GetOnPremPrerunCredentialsFromNetrc(*cobra.Command, netrc.NetrcMachineParams) func() (*Credentials, error)

	// Needed SSO login for non-prod accounts
	SetCloudClient(client *ccloud.Client)
}

type LoginCredentialsManagerImpl struct {
	netrcHandler netrc.NetrcHandler
	prompt       form.Prompt
	client       *ccloud.Client
}

func NewLoginCredentialsManager(netrcHandler netrc.NetrcHandler, prompt form.Prompt, client *ccloud.Client) LoginCredentialsManager {
	return &LoginCredentialsManagerImpl{
		netrcHandler: netrcHandler,
		prompt:       prompt,
		client:       client,
	}
}

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromEnvVar(cmd *cobra.Command, orgResourceId string) func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           ConfluentCloudEmail,
		password:           ConfluentCloudPassword,
		deprecatedUsername: DeprecatedConfluentCloudEmail,
		deprecatedPassword: DeprecatedConfluentCloudPassword,
	}
	return h.getCredentialsFromEnvVarFunc(cmd, envVars, orgResourceId)
}

func (h *LoginCredentialsManagerImpl) getCredentialsFromEnvVarFunc(cmd *cobra.Command, envVars environmentVariables, orgResourceId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		email, password := h.getEnvVarCredentials(cmd, envVars.username, envVars.password)
		if h.isSSOUser(email, orgResourceId) {
			log.CliLogger.Debugf("%s=%s belongs to an SSO user.", ConfluentCloudEmail, email)
			return &Credentials{Username: email, IsSSO: true}, nil
		}

		if email == "" {
			email, password = h.getEnvVarCredentials(cmd, envVars.deprecatedUsername, envVars.deprecatedPassword)
			if email != "" {
				_, _ = fmt.Fprintf(os.Stderr, errors.DeprecatedEnvVarWarningMsg, envVars.deprecatedUsername, envVars.username)
			}
			if password != "" {
				_, _ = fmt.Fprintf(os.Stderr, errors.DeprecatedEnvVarWarningMsg, envVars.deprecatedPassword, envVars.password)
			}
		}

		if password == "" {
			log.CliLogger.Debug("Did not find full credential set from environment variables")
			return nil, nil
		}

		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getEnvVarCredentials(cmd *cobra.Command, userEnvVar string, passwordEnvVar string) (string, string) {
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

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           ConfluentPlatformUsername,
		password:           ConfluentPlatformPassword,
		deprecatedUsername: DeprecatedConfluentPlatformUsername,
		deprecatedPassword: DeprecatedConfluentPlatformPassword,
	}
	return h.getCredentialsFromEnvVarFunc(cmd, envVars, "")
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
		return credentials, err
	}
}

func (h *LoginCredentialsManagerImpl) GetSsoCredentialsFromConfig(cfg *v1.Config) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		credentials, _ := h.GetPrerunCredentialsFromConfig(cfg)()

		// For `confluent login`, only retrieve credentials from the config file if SSO (prevents a breaking change)
		if credentials != nil && credentials.IsSSO {
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
			IsSSO:     ctx.GetUser().GetAuthType() == orgv1.AuthType_AUTH_TYPE_SSO || ctx.GetUser().GetSocialConnection() != "",
			Username:  ctx.GetUser().GetEmail(),
			AuthToken: ctx.GetAuthToken(),
		}
		log.CliLogger.Tracef("credentials: %#v", credentials)

		return credentials, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.NetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(filterParams)
		if err != nil {
			log.CliLogger.Debugf("Get netrc machine error: %s", err.Error())
			return nil, err
		}
		if log.CliLogger.GetLevel() >= log.WARN {
			utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		}
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password, IsSSO: netrcMachine.IsSSO}, nil
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

func (h *LoginCredentialsManagerImpl) GetCloudCredentialsFromPrompt(cmd *cobra.Command, orgResourceId string) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		utils.Println(cmd, "Enter your Confluent Cloud credentials:")
		email := h.promptForUser(cmd, "Email")
		if h.isSSOUser(email, orgResourceId) {
			log.CliLogger.Debug("Entered email belongs to an SSO user.")
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		password := h.promptForPassword(cmd)
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetOnPremCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		utils.Println(cmd, "Enter your Confluent credentials:")
		username := h.promptForUser(cmd, "Username")
		password := h.promptForPassword(cmd)
		return &Credentials{Username: username, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) promptForUser(cmd *cobra.Command, userField string) string {
	f := form.New(form.Field{ID: userField, Prompt: userField})
	if err := f.Prompt(cmd, h.prompt); err != nil {
		return ""
	}

	return f.Responses[userField].(string)
}

func (h *LoginCredentialsManagerImpl) promptForPassword(cmd *cobra.Command) string {
	passwordField := "Password"
	f := form.New(form.Field{ID: passwordField, Prompt: passwordField, IsHidden: true})
	if err := f.Prompt(cmd, h.prompt); err != nil {
		return ""
	}
	return f.Responses[passwordField].(string)
}

func (h *LoginCredentialsManagerImpl) isSSOUser(email, orgId string) bool {
	if h.client == nil {
		return false
	}
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(h.client.BaseURL)
	log.CliLogger.Debugf("cloudClient.BaseURL: %s", h.client.BaseURL)
	log.CliLogger.Debugf("auth0ClientId: %s", auth0ClientId)
	loginRealmReply, err := h.client.User.LoginRealm(context.Background(),
		&flowv1.GetLoginRealmRequest{
			Email:         email,
			ClientId:      auth0ClientId,
			OrgResourceId: orgId,
		})
	// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
	// email, we want to silently continue and give the illusion of password prompt.
	if err == nil && loginRealmReply.IsSso {
		return true
	}
	return false
}

// Prerun login for Confluent has two extra environment variables settings: CONFLUENT_MDS_URL (required), CONFLUNET_CA_CERT_PATH (optional)
// Those two variables are passed as flags for login command, but for prerun logins they are required as environment variables.
// URL and ca-cert-path (if exists) are returned in addition to username and password
func (h *LoginCredentialsManagerImpl) GetOnPremPrerunCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
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

		creds, _ := h.getCredentialsFromEnvVarFunc(cmd, envVars, "")()
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
		// TODO: change to verbosity level logging
		utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password, IsSSO: netrcMachine.IsSSO, PrerunLoginURL: machineContextInfo.URL, PrerunLoginCaCertPath: machineContextInfo.CaCertPath}, nil
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

func (h *LoginCredentialsManagerImpl) SetCloudClient(client *ccloud.Client) {
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
