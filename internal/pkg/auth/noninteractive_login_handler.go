//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/noninteractive_login_handler.go --pkg mock --selfpkg github.com/confluentinc/cli noninteractive_login_handler.go NonInteractiveLoginHandler
package auth

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Credentials struct {
	Username     string
	Password     string
	RefreshToken string
}

type NonInteractiveLoginHandler interface {
	GetCCloudTokenAndCredentialsFromEnvVar(cmd *cobra.Command, client *ccloud.Client) (string, *Credentials, error)
	GetCCloudTokenAndCredentialsFromNetrc(cmd *cobra.Command, client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error)
	GetCCloudTokenAndCredentialsFromPrompt(cmd *cobra.Command, prompt form.Prompt, client *ccloud.Client, url string) (string, *Credentials, error)
	GetConfluentTokenAndCredentialsFromEnvVar(cmd *cobra.Command, client *mds.APIClient) (string, *Credentials, error)
	GetConfluentTokenAndCredentialsFromNetrc(cmd *cobra.Command, client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error)
	GetConfluentTokenAndCredentialsFromPrompt(cmd *cobra.Command, prompt form.Prompt, client *mds.APIClient) (string, *Credentials, error)
}

type NonInteractiveLoginHandlerImpl struct {
	authTokenHandler AuthTokenHandler
	netrcHandler     netrc.NetrcHandler
	logger           *log.Logger
}

func NewNonInteractiveLoginHandler(authTokenHandler AuthTokenHandler, netrcHandler netrc.NetrcHandler, logger *log.Logger) NonInteractiveLoginHandler {
	return &NonInteractiveLoginHandlerImpl{
		authTokenHandler: authTokenHandler,
		netrcHandler:     netrcHandler,
		logger:           logger,
	}
}

func (h *NonInteractiveLoginHandlerImpl) GetCCloudTokenAndCredentialsFromEnvVar(cmd *cobra.Command, client *ccloud.Client) (string, *Credentials, error) {
	email, password := h.getEnvVarCredentials(cmd, CCloudEmailEnvVar, CCloudPasswordEnvVar)
	if len(email) == 0 {
		email, password = h.getEnvVarCredentials(cmd, CCloudEmailDeprecatedEnvVar, CCloudPasswordDeprecatedEnvVar)
	}
	if len(email) == 0 {
		return "", nil, nil
	}
	token, err := h.authTokenHandler.GetCCloudCredentialsToken(client, email, password)
	if err != nil {
		return "", nil, err
	}
	return token, &Credentials{Username: email, Password: password}, nil
}

func (h *NonInteractiveLoginHandlerImpl) getEnvVarCredentials(cmd *cobra.Command, userEnvVar string, passwordEnvVar string) (string, string) {
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

func (h *NonInteractiveLoginHandlerImpl) GetConfluentTokenAndCredentialsFromEnvVar(cmd *cobra.Command, client *mds.APIClient) (string, *Credentials, error) {
	username, password := h.getEnvVarCredentials(cmd, ConfluentUsernameEnvVar, ConfluentPasswordEnvVar)
	if len(username) == 0 {
		username, password = h.getEnvVarCredentials(cmd, ConfluentUsernameDeprecatedEnvVar, ConfluentPasswordDeprecatedEnvVar)
	}
	if len(username) == 0 {
		return "", nil, nil
	}
	token, err := h.authTokenHandler.GetConfluentAuthToken(client, username, password)
	if err != nil {
		return "", nil, err
	}
	return token, &Credentials{Username: username, Password: password}, nil
}

func (h *NonInteractiveLoginHandlerImpl) GetCCloudTokenAndCredentialsFromNetrc(cmd *cobra.Command, client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error) {
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil || netrcMachine == nil {
		return "", nil, err
	}
	utils.ErrPrint(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
	var token string
	creds := &Credentials{Username: netrcMachine.User}
	if netrcMachine.IsSSO {
		token, err = h.authTokenHandler.RefreshCCloudSSOToken(client, netrcMachine.Password, url, h.logger)
		creds.RefreshToken = netrcMachine.Password
	} else {
		token, err = h.authTokenHandler.GetCCloudCredentialsToken(client, netrcMachine.User, netrcMachine.Password)
		creds.Password = netrcMachine.Password
	}
	if err != nil {
		return "", nil, err
	}
	return token, creds, nil
}

func (h *NonInteractiveLoginHandlerImpl) GetConfluentTokenAndCredentialsFromNetrc(cmd *cobra.Command, client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error) {
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil || netrcMachine == nil {
		return "", nil, err
	}
	utils.ErrPrintf(cmd, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
	token, err := h.authTokenHandler.GetConfluentAuthToken(client, netrcMachine.User, netrcMachine.Password)
	if err != nil {
		return "", nil, err
	}
	return token, &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password}, nil
}

func (h *NonInteractiveLoginHandlerImpl) GetCCloudTokenAndCredentialsFromPrompt(cmd *cobra.Command, prompt form.Prompt, client *ccloud.Client, url string) (string, *Credentials, error) {
	email := h.promptForUser(cmd, prompt, "Email")
	if isSSOUser(email, client) {
		noBrowser, err := cmd.Flags().GetBool("no-browser")
		if err != nil {
			return "", nil, err
		}
		token, refreshToken, err := h.authTokenHandler.GetCCloudSSOToken(client, url, noBrowser, email, h.logger)
		if err != nil {
			return "", nil, err
		}
		return token, &Credentials{Username: email, RefreshToken: refreshToken}, nil
	}
	password := h.promptForPassword(cmd, prompt)
	token, err := h.authTokenHandler.GetCCloudCredentialsToken(client, email, password)
	if err != nil {
		return "", nil, nil
	}
	return token, &Credentials{Username: email, Password: password}, nil
}

func (h *NonInteractiveLoginHandlerImpl) GetConfluentTokenAndCredentialsFromPrompt(cmd *cobra.Command, prompt form.Prompt, client *mds.APIClient) (string, *Credentials, error) {
	username := h.promptForUser(cmd, prompt, "Username")
	password := h.promptForPassword(cmd, prompt)
	token, err := h.authTokenHandler.GetConfluentAuthToken(client, username, password)
	if err != nil {
		return "", nil, nil
	}
	return token, &Credentials{Username: username, Password: password}, nil
}

func (h *NonInteractiveLoginHandlerImpl) promptForUser(cmd *cobra.Command, prompt form.Prompt, userField string) string {
	// HACK: SSO integration test extracts email from env var
	// TODO: remove this hack once we implement prompting for integration test
	if testEmail := os.Getenv(CCloudEmailDeprecatedEnvVar); len(testEmail) > 0 {
		h.logger.Debugf("Using test email \"%s\" found from env var \"%s\"", testEmail, CCloudEmailDeprecatedEnvVar)
		return testEmail
	}
	utils.Println(cmd, "Enter your Confluent credentials:")
	f := form.New(form.Field{ID: userField, Prompt: userField})
	if err := f.Prompt(cmd, prompt); err != nil {
		return ""
	}
	return f.Responses[userField].(string)
}

func (h *NonInteractiveLoginHandlerImpl) promptForPassword(cmd *cobra.Command, prompt form.Prompt) string {
	passwordField := "Password"
	f := form.New(form.Field{ID: passwordField, Prompt: passwordField, IsHidden: true})
	if err := f.Prompt(cmd, prompt); err != nil {
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
