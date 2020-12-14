//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/login_credentials_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_credentials_manager.go LoginCredentialsManager
package auth

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"

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
	GetCCloudCredentialsFromPrompt(cmd *cobra.Command, client *ccloud.Client) func() (*Credentials, error)
	GetConfluentCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error)
	GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error)

	// Only for Confluent Prerun login
	GetConfluentPrerunCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentPrerunCredentialsFromNetrc(cmd *cobra.Command) func() (*Credentials, error)
}

type LoginCredentialsManagerImpl struct {
	netrcHandler netrc.NetrcHandler
	logger       *log.Logger
	prompt       form.Prompt
}

func NewLoginCredentialsManager(netrcHandler netrc.NetrcHandler, prompt form.Prompt, logger *log.Logger) LoginCredentialsManager {
	return &LoginCredentialsManagerImpl{
		netrcHandler: netrcHandler,
		logger:       logger,
		prompt:       prompt,
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
		if len(email) == 0 {
			email, password = h.getEnvVarCredentials(cmd, envVars.deprecatedUsername, envVars.deprecatedPassword)
		}
		if len(email) == 0 {
			h.logger.Debug("Found no credentials from environment variables")
			return nil, nil
		}
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) getEnvVarCredentials(cmd *cobra.Command, userEnvVar string, passwordEnvVar string) (string, string) {
	user := os.Getenv(userEnvVar)
	if len(user) == 0 {
		return "", ""
	}
	password := os.Getenv(passwordEnvVar)
	if len(password) == 0 {
		return "", ""
	}
	utils.ErrPrintf(cmd, errors.FoundEnvCredMsg, user, userEnvVar, passwordEnvVar)
	return user, password
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

func (h *LoginCredentialsManagerImpl) GetCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		netrcMachine, err := h.getNetrcMachine(filterParams)
		if err != nil {
			h.logger.Debugf("Get netrc machine error: %s", err.Error())
			return nil, err
		}
		utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
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

func (h *LoginCredentialsManagerImpl) GetCCloudCredentialsFromPrompt(cmd *cobra.Command, client *ccloud.Client) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		email := h.promptForUser(cmd, "Email")
		if isSSOUser(email, client) {
			h.logger.Debug("Entered email belongs to an SSO user.")
			return &Credentials{Username: email, IsSSO: true}, nil
		}
		password := h.promptForPassword(cmd)
		return &Credentials{Username: email, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetConfluentCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
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
	utils.Println(cmd, "Enter your Confluent credentials:")
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

func isSSOUser(email string, cloudClient *ccloud.Client) bool {
	userSSO, err := cloudClient.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	// Fine to ignore non-nil err for this request: e.g. what if this fails due to invalid/malicious
	// email, we want to silently continue and give the illusion of password prompt.
	// If Auth0ConnectionName is blank ("local" user) still prompt for password
	if err == nil && userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		return true
	}
	return false
}

// Prerun login for Confluent has two extra environment variables settings: CONFLUENT_URL (required), CONFLUNET_CA_CERT_PATH (optional)
// Those two variables are passed as flags for login command, but for prerun logins they are required as environment variables
// URL and ca-cert-path (if exists) are returned in addtion to username and passowrd
func (h *LoginCredentialsManagerImpl) GetConfluentPrerunCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		url := os.Getenv(ConfluentURLEnvVar)
		if url == "" {
			return nil, errors.New("No URL env var set")
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
		creds.PrerunLoginURL = url
		creds.PrerunLoginCaCertPath = os.Getenv(ConfluentCaCertPathEnvVar)
		return creds, nil
	}
}

// Prerun login for Confluent will extract URL and ca-cert-path (if available) from the netrc machine name
// URL is no longer part of the filter and URL value will be of whichever URL the first context stored in netrc has
// URL and ca-cert-path (if exists) are returned in addtion to username and passowrd
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
		utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password, IsSSO: netrcMachine.IsSSO, PrerunLoginURL: machineContextInfo.URL, PrerunLoginCaCertPath: machineContextInfo.CaCertPath}, nil
	}
}
