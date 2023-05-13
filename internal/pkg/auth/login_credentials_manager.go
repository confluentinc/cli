//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/login_credentials_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_credentials_manager.go LoginCredentialsManager
package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/sso"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go-v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
	GetCCloudCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetCCloudCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error)
	GetCredentialsFromConfig(cfg *v3.Config, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error)
	GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error)

	// Only for Confluent Prerun login
	GetConfluentPrerunCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentPrerunCredentialsFromNetrc(cmd *cobra.Command) func() (*Credentials, error)

	// Needed SSO login for non-prod accounts
	SetCCloudClient(client *ccloud.Client)
}

type LoginCredentialsManagerImpl struct {
	netrcHandler netrc.NetrcHandler
	logger       *log.Logger
	prompt       form.Prompt
	client       *ccloud.Client
}

func NewLoginCredentialsManager(netrcHandler netrc.NetrcHandler, prompt form.Prompt, logger *log.Logger, client *ccloud.Client) LoginCredentialsManager {
	return &LoginCredentialsManagerImpl{
		netrcHandler: netrcHandler,
		logger:       logger,
		prompt:       prompt,
		client:       client,
	}
}

func (h *LoginCredentialsManagerImpl) GetCCloudCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           CCloudEmailEnvVar,
		password:           CCloudPasswordEnvVar,
		deprecatedUsername: CCloudEmailDeprecatedEnvVar,
		deprecatedPassword: CCloudPasswordDeprecatedEnvVar,
	}
	return h.getCredentialsFromEnvVarFunc(cmd, envVars)
}

func (h *LoginCredentialsManagerImpl) getCredentialsFromEnvVarFunc(cmd *cobra.Command, envVars environmentVariables) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		email, password := h.getEnvVarCredentials(cmd, envVars.username, envVars.password)
		if h.isSSOUser(email) {
			h.logger.Debugf("CCLOUD_EMAIL=%s belongs to an SSO user.", email)
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		if len(email) == 0 {
			email, password = h.getEnvVarCredentials(cmd, envVars.deprecatedUsername, envVars.deprecatedPassword)
		}
		if len(password) == 0 {
			h.logger.Debug("Did not find full credential set from environment variables")
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
	if h.logger.GetLevel() >= log.WARN {
		utils.ErrPrintf(cmd, errors.FoundEnvCredMsg, username, userEnvVar, passwordEnvVar)
	}
	return username, password
}

func (h *LoginCredentialsManagerImpl) GetConfluentCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
	envVars := environmentVariables{
		username:           ConfluentUsernameEnvVar,
		password:           ConfluentPasswordEnvVar,
		deprecatedUsername: ConfluentUsernameDeprecatedEnvVar,
		deprecatedPassword: ConfluentPasswordDeprecatedEnvVar,
	}
	return h.getCredentialsFromEnvVarFunc(cmd, envVars)
}

func (h *LoginCredentialsManagerImpl) GetCredentialsFromConfig(cfg *v3.Config, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		var loginCredential *v2.LoginCredential
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

func (h *LoginCredentialsManagerImpl) GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(filterParams)
		if err != nil {
			h.logger.Debugf("Get netrc machine error: %s", err.Error())
			return nil, err
		}
		if h.logger.GetLevel() >= log.WARN {
			utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		}
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password, IsSSO: netrcMachine.IsSSO}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getNetrcMachine(filterParams netrc.GetMatchingNetrcMachineParams) (*netrc.Machine, error) {
	h.logger.Debugf("Searching for netrc machine with filter: %+v", filterParams)
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil {
		return nil, err
	}
	if netrcMachine == nil {
		return nil, errors.Errorf("Found no netrc machine using the filter: %+v", filterParams)
	}
	return netrcMachine, err
}

func (h *LoginCredentialsManagerImpl) GetCCloudCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		utils.Println(cmd, "Enter your Confluent Cloud credentials:")
		email := h.promptForUser(cmd, "Email")
		if h.isSSOUser(email) {
			h.logger.Debug("Entered email belongs to an SSO user.")
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		password := h.promptForPassword(cmd)
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetConfluentCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		utils.Println(cmd, "Enter your Confluent credentials:")
		username := h.promptForUser(cmd, "Username")
		password := h.promptForPassword(cmd)
		return &Credentials{Username: username, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) promptForUser(cmd *cobra.Command, userField string) string {
	// HACK: SSO integration test extracts email from env var
	// TODO: remove this hack once we implement prompting for integration test
	if testEmail := os.Getenv(CCloudEmailDeprecatedEnvVar); len(testEmail) > 0 {
		h.logger.Debugf("Using test email \"%s\" found from env var \"%s\"", testEmail, CCloudEmailDeprecatedEnvVar)
		return testEmail
	}

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

func (h *LoginCredentialsManagerImpl) isSSOUser(email string) bool {
	if h.client == nil {
		return false
	}
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(h.client.BaseURL)
	h.logger.Debugf("cloudClient.BaseURL: %s", h.client.BaseURL)
	h.logger.Debugf("auth0ClientId: %s", auth0ClientId)
	loginRealmReply, err := h.client.User.LoginRealm(context.Background(),
		&flowv1.GetLoginRealmRequest{
			Email:    email,
			ClientId: auth0ClientId,
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
func (h *LoginCredentialsManagerImpl) GetConfluentPrerunCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		url := os.Getenv(ConfluentURLEnvVar)
		if url == "" {
			return nil, errors.New(errors.NoURLEnvVarErrorMsg)
		}
		envVars := environmentVariables{
			username:           ConfluentUsernameEnvVar,
			password:           ConfluentPasswordEnvVar,
			deprecatedUsername: ConfluentUsernameDeprecatedEnvVar,
			deprecatedPassword: ConfluentPasswordDeprecatedEnvVar,
		}
		creds, err := h.getCredentialsFromEnvVarFunc(cmd, envVars)()
		if err != nil {
			return nil, err
		}
		if creds == nil {
			return nil, errors.New(errors.NoCredentialsFoundErrorMsg)
		}
		creds.PrerunLoginURL = url
		creds.PrerunLoginCaCertPath = os.Getenv(ConfluentCaCertPathEnvVar)
		return creds, nil
	}
}

// Prerun login for Confluent will extract URL and ca-cert-path (if available) from the netrc machine name
// URL is no longer part of the filter and URL value will be of whichever URL the first context stored in netrc has
// URL and ca-cert-path (if exists) are returned in addition to username and password
func (h *LoginCredentialsManagerImpl) GetConfluentPrerunCredentialsFromNetrc(cmd *cobra.Command) func() (*Credentials, error) {
	filterParams := netrc.GetMatchingNetrcMachineParams{
		CLIName: "confluent",
	}
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(filterParams)
		if err != nil {
			h.logger.Debugf("Get netrc machine error: %s", err.Error())
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

func (h *LoginCredentialsManagerImpl) SetCCloudClient(client *ccloud.Client) {
	h.client = client
}

func matchLoginCredentialWithFilter(loginCredential *v2.LoginCredential, filterParams netrc.GetMatchingNetrcMachineParams) bool {
	if loginCredential == nil {
		return false
	}
	if loginCredential.Url != filterParams.URL {
		return false
	}
	fmt.Println("trying to match? params ctx name:", filterParams.CtxName)
	if filterParams.CtxName != "" && !strings.Contains(filterParams.CtxName, loginCredential.Username) {
		return false
	}
	return true
}
