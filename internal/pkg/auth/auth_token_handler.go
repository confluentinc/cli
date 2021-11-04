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
	GetCCloudTokens(client *ccloud.Client, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error)
	GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error)
}

type AuthTokenHandlerImpl struct {
	logger *log.Logger
}

func NewAuthTokenHandler(logger *log.Logger) AuthTokenHandler {
	return &AuthTokenHandlerImpl{logger}
}

func (a *AuthTokenHandlerImpl) GetCCloudTokens(client *ccloud.Client, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error) {
	if credentials.IsSSO {
		// For an SSO user, the "Password" field may contain a refresh token. If one exists, try to obtain a new token.
		if credentials.Password != "" {
			if token, refreshToken, err := a.refreshCCloudSSOToken(client, credentials.Password, orgResourceId); err == nil {
				return token, refreshToken, nil
			}
		}
		return a.getCCloudSSOToken(client, noBrowser, credentials.Username, orgResourceId)
	}

	client.HttpClient.Timeout = 30 * time.Second
	token, err := client.Auth.Login(context.Background(), "", credentials.Username, credentials.Password, orgResourceId)
	return token, "", err
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(client *ccloud.Client, noBrowser bool, email string, orgResourceId string) (string, string, error) {
	userSSO, err := a.getCCloudUserSSO(client, email, orgResourceId)
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

func (a *AuthTokenHandlerImpl) getCCloudUserSSO(client *ccloud.Client, email string, orgResourceId string) (string, error) {
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(client.BaseURL)
	loginRealmReply, err := client.User.LoginRealm(context.Background(),
		&flowv1.GetLoginRealmRequest{
			Email:         email,
			ClientId:      auth0ClientId,
			OrgResourceId: orgResourceId,
		})
	if err != nil {
		return "", err
	}
	if loginRealmReply.IsSso {
		return loginRealmReply.Realm, nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(client *ccloud.Client, refreshToken string, orgResourceId string) (string, string, error) {
	idToken, refreshToken, err := sso.RefreshTokens(client.BaseURL, refreshToken, a.logger)
	if err != nil {
		return "", "", err
	}

<<<<<<< HEAD
	token, err := client.Auth.Login(context.Background(), idToken, "", "", orgResourceId)
=======
	token, err := client.Auth.Login(context.Background(), idToken, "", "", "")
>>>>>>> 547809ffd970c8f051c005e427fc5f172dbaac34
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, err
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
