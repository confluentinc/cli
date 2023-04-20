//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"
	"fmt"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/auth/sso"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type AuthTokenHandler interface {
	GetCCloudTokens(clientFactory CCloudClientFactory, url string, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error)
	GetConfluentToken(mdsClient *mds.APIClient, credentials *Credentials) (string, error)
}

type AuthTokenHandlerImpl struct{}

func NewAuthTokenHandler() AuthTokenHandler {
	return &AuthTokenHandlerImpl{}
}

func (a *AuthTokenHandlerImpl) GetCCloudTokens(clientFactory CCloudClientFactory, url string, credentials *Credentials, noBrowser bool, orgResourceId string) (string, string, error) {
	client := clientFactory.AnonHTTPClientFactory(url)

	if credentials.AuthRefreshToken != "" {
		if credentials.IsSSO {
			if token, refreshToken, err := a.refreshCCloudSSOToken(client, credentials.AuthRefreshToken, orgResourceId); err == nil {
				return token, refreshToken, nil
			}
		} else {
			req := &ccloudv1.AuthenticateRequest{
				RefreshToken:  credentials.AuthRefreshToken,
				OrgResourceId: orgResourceId,
			}
			if res, err := client.Auth.Login(context.Background(), req); err == nil {
				return res.GetToken(), res.GetRefreshToken(), nil
			}
		}
	}

	// If SSO refresh token is missing or expired, ask for a new one
	if credentials.IsSSO {
		token, refreshToken, err := a.getCCloudSSOToken(client, noBrowser, credentials.Username, orgResourceId)
		if err != nil {
			return "", "", err
		}

		client = clientFactory.JwtHTTPClientFactory(context.Background(), token, url)
		err = a.checkSSOEmailMatchesLogin(client, credentials.Username)
		return token, refreshToken, err
	}

	client.HttpClient.Timeout = 30 * time.Second
	log.CliLogger.Debugf("Making login request for %s for org id %s", credentials.Username, orgResourceId)

	req := &ccloudv1.AuthenticateRequest{
		Email:         credentials.Username,
		Password:      credentials.Password,
		OrgResourceId: orgResourceId,
	}

	res, err := client.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	if utils.IsOrgEndOfFreeTrialSuspended(res.GetOrganization().GetSuspensionStatus()) {
		log.CliLogger.Debugf(errors.EndOfFreeTrialErrorMsg, res.GetOrganization().GetSuspensionStatus())
		return res.GetToken(), res.GetRefreshToken(), &errors.EndOfFreeTrialError{OrgId: res.GetOrganization().GetName()}
	}

	return res.GetToken(), res.GetRefreshToken(), nil
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(client *ccloudv1.Client, noBrowser bool, email, orgResourceId string) (string, string, error) {
	var auth0ConnectionName string
	if types.Contains([]string{"fedramp", "fedramp-internal"}, sso.GetCCloudEnvFromBaseUrl(client.BaseURL)) {
		auth0ConnectionName = ""
	} else {
		userSSO, err := a.getCCloudUserSSO(client, email, orgResourceId)
		if err != nil {
			log.CliLogger.Debugf("unable to obtain user SSO info: %v", err)
			return "", "", errors.Errorf(errors.FailedToObtainedUserSSOErrorMsg, email)
		}
		if userSSO == "" {
			return "", "", errors.Errorf(errors.NonSSOUserErrorMsg, email)
		}
		auth0ConnectionName = userSSO
	}

	idToken, refreshToken, err := sso.Login(client.BaseURL, noBrowser, auth0ConnectionName)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{IdToken: idToken}

	res, err := client.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
}

func (a *AuthTokenHandlerImpl) getCCloudUserSSO(client *ccloudv1.Client, email, orgResourceId string) (string, error) {
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(client.BaseURL)
	req := &ccloudv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      auth0ClientId,
		OrgResourceId: orgResourceId,
	}
	loginRealmReply, err := client.User.LoginRealm(context.Background(), req)
	if err != nil {
		return "", err
	}
	if loginRealmReply.IsSso {
		return loginRealmReply.Realm, nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(client *ccloudv1.Client, refreshToken, orgResourceId string) (string, string, error) {
	idToken, refreshToken, err := sso.RefreshTokens(client.BaseURL, refreshToken)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: orgResourceId,
	}

	res, err := client.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
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

func (a *AuthTokenHandlerImpl) checkSSOEmailMatchesLogin(client *ccloudv1.Client, loginEmail string) error {
	getMeReply, err := client.Auth.User(context.Background())
	if err != nil {
		return err
	}
	if getMeReply.GetUser().GetEmail() != loginEmail {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SSOCredentialsDoNotMatchLoginCredentialsErrorMsg, loginEmail, getMeReply.GetUser().GetEmail()), errors.SSOCredentialsDoNotMatchSuggestions)
	}
	return nil
}
