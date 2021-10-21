//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"
	"time"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/sso"
)

type AuthTokenHandler interface {
	GetCCloudTokens(client *ccloud.Client, credentials *Credentials, noBrowser bool) (string, string, error)
	GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error)
}

type AuthTokenHandlerImpl struct {
	logger *log.Logger
}

func NewAuthTokenHandler(logger *log.Logger) AuthTokenHandler {
	return &AuthTokenHandlerImpl{logger}
}

// Second string returned is refresh token if the user performs SSO login
func (a *AuthTokenHandlerImpl) GetCCloudTokens(client *ccloud.Client, credentials *Credentials, noBrowser bool) (string, string, error) {
	if credentials.IsSSO {
		// SSO password is the refresh token, if not present then user must perform SSO login, if present then refresh token automatically obtains a new token
		if credentials.Password != "" {
			token, err := a.refreshCCloudSSOToken(client, credentials.Password)
			return token, "", err
		} else {
			return a.getCCloudSSOToken(client, noBrowser, credentials.Username)
		}
	}
	client.HttpClient.Timeout = 30 * time.Second
	token, err := client.Auth.Login(context.Background(), "", credentials.Username, credentials.Password, "")
	return token, "", err
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(client *ccloud.Client, noBrowser bool, email string) (string, string, error) {
	userSSO, err := a.getCCloudUserSSO(client, email)
	if err != nil {
		return "", "", errors.Errorf(errors.FailedToObtainedUserSSOErrorMsg, email)
	}
	if userSSO == "" {
		return "", "", errors.Errorf(errors.NonSSOUserErrorMsg, email)
	}
	idToken, refreshToken, err := sso.Login(client.BaseURL, noBrowser, userSSO, a.logger)
	if err != nil {
		return "", "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "", "")
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func (a *AuthTokenHandlerImpl) getCCloudUserSSO(client *ccloud.Client, email string) (string, error) {
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(client.BaseURL)
	loginRealmReply, err := client.User.LoginRealm(context.Background(),
		&flowv1.GetLoginRealmRequest{
			Email:    email,
			ClientId: auth0ClientId,
		})
	if err != nil {
		return "", err
	}
	if loginRealmReply.IsSso {
		return loginRealmReply.Realm, nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(client *ccloud.Client, refreshToken string) (string, error) {
	idToken, err := sso.GetNewIDTokenFromRefreshToken(client.BaseURL, refreshToken, a.logger)
	if err != nil {
		return "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "", "")
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthTokenHandlerImpl) GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error) {
	ctx := utils.GetContext(a.logger)
	basicContext := context.WithValue(ctx, mds.ContextBasicAuth, mds.BasicAuth{UserName: credentials.Username, Password: credentials.Password})
	resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
	if err != nil {
		return "", err
	}
	return resp.AuthToken, nil
}
