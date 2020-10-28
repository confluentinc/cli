//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/sso"
)

// Make into interface in order to create mock for testing
type AuthTokenHandler interface {
	GetCCloudUserSSO(client *ccloud.Client, email string) (*orgv1.User, error)
	GetCCloudCredentialsToken(client *ccloud.Client, email string, password string) (string, error)
	GetCCloudSSOToken(client *ccloud.Client, url string, noBrowser bool, email string, logger *log.Logger) (string, string, error)
	RefreshCCloudSSOToken(client *ccloud.Client, refreshToken string, url string, logger *log.Logger) (string, error)
	GetConfluentAuthToken(mdsClient *mds.APIClient, username string, password string) (string, error)
}

type AuthTokenHandlerImpl struct{}

func (a *AuthTokenHandlerImpl) GetCCloudUserSSO(client *ccloud.Client, email string) (*orgv1.User, error) {
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return nil, err
	}
	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		return userSSO, nil
	}
	return nil, nil
}

func (a *AuthTokenHandlerImpl) GetCCloudCredentialsToken(client *ccloud.Client, email string, password string) (string, error) {
	return client.Auth.Login(context.Background(), "", email, password)
}

func (a *AuthTokenHandlerImpl) GetCCloudSSOToken(client *ccloud.Client, url string, noBrowser bool, email string, logger *log.Logger) (string, string, error) {
	userSSO, err := a.GetCCloudUserSSO(client, email)
	if err != nil {
		return "", "", errors.Errorf(errors.FailedToObtainedUserSSOErrorMsg, email)
	}
	if userSSO == nil {
		return "", "", errors.Errorf(errors.NonSSOUserErrorMsg, email)
	}
	idToken, refreshToken, err := sso.Login(url, noBrowser, userSSO.Sso.Auth0ConnectionName, logger)
	if err != nil {
		return "", "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func (a *AuthTokenHandlerImpl) RefreshCCloudSSOToken(client *ccloud.Client, refreshToken string, url string, logger *log.Logger) (string, error) {
	idToken, err := sso.GetNewIDTokenFromRefreshToken(url, refreshToken, logger)
	if err != nil {
		return "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthTokenHandlerImpl) GetConfluentAuthToken(mdsClient *mds.APIClient, username string, password string) (string, error) {
	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: username, Password: password})
	resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
	if err != nil {
		return "", err
	}
	return resp.AuthToken, nil
}
