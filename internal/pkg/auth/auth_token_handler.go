//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/cli/internal/pkg/log"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/sso"
)

type AuthTokenHandler interface {
	GetCCloudTokens(clientFactory CCloudClientFactory, url string, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error)
	GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error)
}

type AuthTokenHandlerImpl struct {
}

func NewAuthTokenHandler() AuthTokenHandler {
	return &AuthTokenHandlerImpl{}
}

func (a *AuthTokenHandlerImpl) GetCCloudTokens(clientFactory CCloudClientFactory, url string, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error) {
	anonClient := clientFactory.AnonHTTPClientFactory(url)
	if credentials.IsSSO {
		// For an SSO user, the "Password" field may contain a refresh token. If one exists, try to obtain a new token.
		if credentials.Password != "" {
			if token, refreshToken, err := a.refreshCCloudSSOToken(anonClient, credentials.Password, orgResourceId); err == nil {
				return token, refreshToken, nil
			}
		}
		token, refreshToken, err := a.getCCloudSSOToken(anonClient, noBrowser, credentials.Username, orgResourceId)
		if err != nil {
			return token, refreshToken, err
		}
		err = a.checkSSOEmailMatchesLogin(clientFactory.JwtHTTPClientFactory(context.Background(), token, url), credentials.Username)
		return token, refreshToken, err
	}

	anonClient.HttpClient.Timeout = 30 * time.Second
	log.CliLogger.Debugf("Making login request for %s for org id %s", credentials.Username, orgResourceId)
	token, err := anonClient.Auth.Login(context.Background(), "", credentials.Username, credentials.Password, orgResourceId)
	return token, "", err
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(client *ccloud.Client, noBrowser bool, email, orgResourceId string) (string, string, error) {
	userSSO, err := a.getCCloudUserSSO(client, email, orgResourceId)
	if err != nil {
		log.CliLogger.Debugf("unable to obtain user SSO info: %s", err.Error())
		return "", "", errors.Errorf(errors.FailedToObtainedUserSSOErrorMsg, email)
	}
	if userSSO == "" {
		return "", "", errors.Errorf(errors.NonSSOUserErrorMsg, email)
	}
	idToken, refreshToken, err := sso.Login(client.BaseURL, noBrowser, userSSO)
	if err != nil {
		return "", "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "", "")
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func (a *AuthTokenHandlerImpl) getCCloudUserSSO(client *ccloud.Client, email, orgResourceId string) (string, error) {
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

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(client *ccloud.Client, refreshToken, orgResourceId string) (string, string, error) {
	idToken, refreshToken, err := sso.RefreshTokens(client.BaseURL, refreshToken)
	if err != nil {
		return "", "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "", orgResourceId)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, err
}

func (a *AuthTokenHandlerImpl) GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error) {
	ctx := utils.GetContext()
	basicContext := context.WithValue(ctx, mds.ContextBasicAuth, mds.BasicAuth{UserName: credentials.Username, Password: credentials.Password})
	resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
	if err != nil {
		return "", err
	}
	return resp.AuthToken, nil
}

func (a *AuthTokenHandlerImpl) checkSSOEmailMatchesLogin(client *ccloud.Client, loginEmail string) error {
	getMeReply, err := getCCloudUser(client)
	if err != nil {
		return err
	}
	if getMeReply.User.Email != loginEmail {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SSOCredentialsDoNotMatchLoginCredentials, loginEmail, getMeReply.User.Email), errors.SSOCrdentialsDoNotMatchSuggestions)
	}
	return nil
}
