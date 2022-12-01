//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/cli/internal/pkg/auth/sso"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
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
	privateClient := clientFactory.PrivateAnonHTTPClientFactory(url)

	if credentials.AuthRefreshToken != "" {
		if credentials.IsSSO {
			if token, refreshToken, err := a.refreshCCloudSSOToken(privateClient, credentials.AuthRefreshToken, orgResourceId); err == nil {
				return token, refreshToken, nil
			}
		} else {
			if orgResourceId == "" {
				orgResourceId = credentials.OrgResourceId
			}
			req := &flowv1.AuthenticateRequest{
				RefreshToken:  credentials.AuthRefreshToken,
				OrgResourceId: orgResourceId,
			}
			if res, err := privateClient.Auth.Login(context.Background(), req); err == nil {
				return res.Token, res.RefreshToken, nil
			}
		}
	}

	// If SSO refresh token is missing or expired, ask for a new one
	if credentials.IsSSO {
		token, refreshToken, err := a.getCCloudSSOToken(privateClient, noBrowser, credentials.Username, orgResourceId)
		if err != nil {
			return "", "", err
		}

		privateClient = clientFactory.PrivateJwtHTTPClientFactory(context.Background(), token, url)
		err = a.checkSSOEmailMatchesLogin(privateClient, credentials.Username)
		return token, refreshToken, err
	}

	privateClient.HttpClient.Timeout = 30 * time.Second
	log.CliLogger.Debugf("Making login request for %s for org id %s", credentials.Username, orgResourceId)

	req := &flowv1.AuthenticateRequest{
		Email:         credentials.Username,
		Password:      credentials.Password,
		OrgResourceId: orgResourceId,
	}

	res, err := privateClient.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	if utils.IsOrgEndOfFreeTrialSuspended(res.GetOrganization().GetSuspensionStatus()) {
		log.CliLogger.Debugf(errors.EndOfFreeTrialErrorMsg, res.GetOrganization().GetSuspensionStatus())
		return res.Token, res.RefreshToken, &errors.EndOfFreeTrialError{OrgId: res.GetOrganization().GetName()}
	}

	return res.Token, res.RefreshToken, nil
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(privateClient *ccloud.Client, noBrowser bool, email, orgResourceId string) (string, string, error) {
	userSSO, err := a.getCCloudUserSSO(privateClient, email, orgResourceId)
	if err != nil {
		log.CliLogger.Debugf("unable to obtain user SSO info: %v", err)
		return "", "", errors.Errorf(errors.FailedToObtainedUserSSOErrorMsg, email)
	}
	if userSSO == "" {
		return "", "", errors.Errorf(errors.NonSSOUserErrorMsg, email)
	}

	idToken, refreshToken, err := sso.Login(privateClient.BaseURL, noBrowser, userSSO)
	if err != nil {
		return "", "", err
	}

	req := &flowv1.AuthenticateRequest{IdToken: idToken}

	res, err := privateClient.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	return res.Token, refreshToken, err
}

func (a *AuthTokenHandlerImpl) getCCloudUserSSO(privateClient *ccloud.Client, email, orgResourceId string) (string, error) {
	auth0ClientId := sso.GetAuth0CCloudClientIdFromBaseUrl(privateClient.BaseURL)
	req := &flowv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      auth0ClientId,
		OrgResourceId: orgResourceId,
	}
	loginRealmReply, err := privateClient.User.LoginRealm(context.Background(), req)
	if err != nil {
		return "", err
	}
	if loginRealmReply.IsSso {
		return loginRealmReply.Realm, nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(privateClient *ccloud.Client, refreshToken, orgResourceId string) (string, string, error) {
	idToken, refreshToken, err := sso.RefreshTokens(privateClient.BaseURL, refreshToken)
	if err != nil {
		return "", "", err
	}

	req := &flowv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: orgResourceId,
	}

	res, err := privateClient.Auth.Login(context.Background(), req)
	if err != nil {
		return "", "", err
	}

	return res.Token, refreshToken, err
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

func (a *AuthTokenHandlerImpl) checkSSOEmailMatchesLogin(privateClient *ccloud.Client, loginEmail string) error {
	getMeReply, err := privateClient.Auth.User(context.Background())
	if err != nil {
		return err
	}
	if getMeReply.User.Email != loginEmail {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.SSOCredentialsDoNotMatchLoginCredentialsErrorMsg, loginEmail, getMeReply.User.Email), errors.SSOCredentialsDoNotMatchSuggestions)
	}
	return nil
}
