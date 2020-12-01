//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/log_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli login_token_handler.go LoginCredentialsManager
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
}

// Get login credentials using the functions from LoginCredentialsManager
// Functions are called in order and credentials are returned right away if found from a function
func GetLoginCredentials(credentialsFuncs ...func() (*Credentials, error)) (*Credentials, error) {
	var credentials *Credentials
	var err error
	for _, credentialsFunc := range credentialsFuncs {
		credentials, err = credentialsFunc()
		if err == nil && credentials.Username != "" {
			break
		}
	}
	return credentials, err
}

type LoginCredentialsManager interface {
	GetCCloudCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetCCloudCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error)
	GetCCloudCredentialsFromPrompt(cmd *cobra.Command, client *ccloud.Client) func() (*Credentials, error)
	GetConfluentCredentialsFromEnvVar(cmd *cobra.Command) func() (*Credentials, error)
	GetConfluentCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error)
	GetConfluentCredentialsFromPrompt(cmd *cobra.Command) func() (*Credentials, error)
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
	return func() (*Credentials, error) {
		email, password := h.getEnvVarCredentials(cmd, CCloudEmailEnvVar, CCloudPasswordEnvVar)
		if len(email) == 0 {
			email, password = h.getEnvVarCredentials(cmd, CCloudEmailDeprecatedEnvVar, CCloudPasswordDeprecatedEnvVar)
		}
		if len(email) == 0 {
			h.logger.Debug("Found no credentials from environment variables")
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
	return func() (*Credentials, error) {
		username, password := h.getEnvVarCredentials(cmd, ConfluentUsernameEnvVar, ConfluentPasswordEnvVar)
		if len(username) == 0 {
			username, password = h.getEnvVarCredentials(cmd, ConfluentUsernameDeprecatedEnvVar, ConfluentPasswordDeprecatedEnvVar)
		}
		if len(username) == 0 {
			h.logger.Debug("Found no credentials from environment variables")
		}
		return &Credentials{Username: username, Password: password}, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetCCloudCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		h.logger.Debugf("Searching for netrc machine with filter: %+v", filterParams)
		netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
		if err != nil || netrcMachine == nil {
			h.logger.Debug("Failed to get netrc machine for credentials")
			if err != nil {
				h.logger.Debugf("Get netrc machine error: %s", err.Error())
			}
			return nil, err
		}
		utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		creds := &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password}
		if netrcMachine.IsSSO {
			creds.IsSSO = true
		}
		return creds, nil
	}
}

func (h *LoginCredentialsManagerImpl) GetConfluentCredentialsFromNetrc(cmd *cobra.Command, filterParams netrc.GetMatchingNetrcMachineParams) func() (*Credentials, error) {
	return func() (*Credentials, error) {
		h.logger.Debugf("Searching for netrc machine with filter: %+v", filterParams)
		netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
		if err != nil || netrcMachine == nil {
			h.logger.Debug("Failed to get netrc machine for credentials")
			if err != nil {
				h.logger.Debugf("Get netrc machine error: %s", err.Error())
			}
			return nil, err
		}
		utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
		return &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password}, nil
	}
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
