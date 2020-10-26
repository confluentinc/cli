package auth

import (
	"fmt"
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

const (
	CCloudEmailEnvVar       = "CCLOUD_EMAIL"
	ConfluentUsernameEnvVar = "CONFLUENT_USERNAME"
	CCloudPasswordEnvVar    = "CCLOUD_PASSWORD"
	ConfluentPasswordEnvVar = "CONFLUENT_PASSWORD"

	CCloudEmailDeprecatedEnvVar       = "XX_CCLOUD_EMAIL"
	ConfluentUsernameDeprecatedEnvVar = "XX_CONFLUENT_USERNAME"
	CCloudPasswordDeprecatedEnvVar    = "XX_CCLOUD_PASSWORD"
	ConfluentPasswordDeprecatedEnvVar = "XX_CONFLUENT_PASSWORD"
)

type Credentials struct {
	Username     string
	Password     string
	RefreshToken string
}

type NonInteractiveLoginHandler interface {
	GetCCloudTokenAndCredentialsFromEnvVar(client *ccloud.Client) (string, *Credentials, error)
	GetCCloudTokenAndCredentialsFromNetrc(client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error)
	GetConfluentTokenAndCredentialsFromEnvVar(client *mds.APIClient) (string, *Credentials, error)
	GetConfluentTokenAndCredentialsFromNetrc(client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error)
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

func (h *NonInteractiveLoginHandlerImpl) GetCCloudTokenAndCredentialsFromEnvVar(client *ccloud.Client) (string, *Credentials, error) {
	email, password := h.getEnvVarCredentials(CCloudEmailEnvVar, CCloudPasswordEnvVar)
	if len(email) == 0 {
		email, password = h.getEnvVarCredentials(CCloudEmailDeprecatedEnvVar, CCloudPasswordDeprecatedEnvVar)
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

func (h *NonInteractiveLoginHandlerImpl) getEnvVarCredentials(userEnvVar string, passwordEnvVar string) (string, string) {
	user := os.Getenv(userEnvVar)
	if len(user) == 0 {
		return "", ""
	}
	password := os.Getenv(passwordEnvVar)
	if len(password) == 0 {
		return "", ""
	}
	fmt.Fprintf(os.Stderr, errors.FoundEnvCredMsg, user, userEnvVar, passwordEnvVar)
	return user, password
}

func (h *NonInteractiveLoginHandlerImpl) GetConfluentTokenAndCredentialsFromEnvVar(client *mds.APIClient) (string, *Credentials, error) {
	username, password := h.getEnvVarCredentials(ConfluentUsernameEnvVar, ConfluentPasswordEnvVar)
	if len(username) == 0 {
		username, password = h.getEnvVarCredentials(ConfluentUsernameDeprecatedEnvVar, ConfluentPasswordDeprecatedEnvVar)
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

func (h *NonInteractiveLoginHandlerImpl) GetCCloudTokenAndCredentialsFromNetrc(client *ccloud.Client, url string, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error) {
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil || netrcMachine == nil {
		return "", nil, err
	}
	fmt.Fprintf(os.Stderr, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
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

func (h *NonInteractiveLoginHandlerImpl) GetConfluentTokenAndCredentialsFromNetrc(client *mds.APIClient, filterParams netrc.GetMatchingNetrcMachineParams) (string, *Credentials, error) {
	netrcMachine, err := h.netrcHandler.GetMatchingNetrcMachine(filterParams)
	if err != nil || netrcMachine == nil {
		return "", nil, err
	}
	fmt.Fprintf(os.Stderr, errors.FoundNetrcCredMsg, netrcMachine.User, h.netrcHandler.GetFileName())
	token, err := h.authTokenHandler.GetConfluentAuthToken(client, netrcMachine.User, netrcMachine.Password)
	return token, &Credentials{Username: netrcMachine.User, Password: netrcMachine.Password}, nil
}
