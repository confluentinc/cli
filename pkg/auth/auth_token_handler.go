//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../mock/auth_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli/v4 auth_token_handler.go AuthTokenHandler
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/browser"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/auth/mfa"
	"github.com/confluentinc/cli/v4/pkg/auth/sso"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/retry"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type AuthTokenHandler interface {
	GetCCloudTokens(CCloudClientFactory, string, *Credentials, bool, string) (string, string, error)
	GetConfluentToken(*mdsv1.APIClient, *Credentials, bool) (string, string, error)
}

type AuthTokenHandlerImpl struct{}

func NewAuthTokenHandler() AuthTokenHandler {
	return &AuthTokenHandlerImpl{}
}

func (a *AuthTokenHandlerImpl) GetCCloudTokens(clientFactory CCloudClientFactory, url string, credentials *Credentials, noBrowser bool, organizationId string) (string, string, error) {
	client := clientFactory.AnonHTTPClientFactory(url)

	if credentials.AuthRefreshToken != "" {
		if credentials.IsSSO {
			if token, refreshToken, err := a.refreshCCloudSSOToken(client, credentials.AuthRefreshToken, organizationId); err == nil {
				return token, refreshToken, nil
			} else if err != nil {
				log.CliLogger.Debugf("error refreshing SSO token: %s", err.Error())
			}
		} else if credentials.IsMFA {
			if token, refreshToken, err := a.refreshCCloudMFAToken(client, credentials.AuthRefreshToken, organizationId, credentials.Username); err == nil {
				return token, refreshToken, nil
			} else if err != nil {
				log.CliLogger.Debugf("error refreshing MFA token: %s", err.Error())
			}
		} else {
			req := &ccloudv1.AuthenticateRequest{
				RefreshToken:  credentials.AuthRefreshToken,
				OrgResourceId: organizationId,
			}

			if res, err := client.Auth.Login(req); err == nil {
				return res.GetToken(), res.GetRefreshToken(), nil
			}
		}
	}

	// If SSO refresh token is missing or expired, ask for a new one
	if credentials.IsSSO {
		token, refreshToken, err := a.getCCloudSSOToken(client, noBrowser, credentials.Username, organizationId)
		if err != nil {
			return "", "", err
		}

		client = clientFactory.JwtHTTPClientFactory(context.Background(), token, url)
		err = a.checkSSOEmailMatchesLogin(client, credentials.Username)
		return token, refreshToken, err
	}

	if credentials.IsMFA {
		token, refreshToken, err := a.getCCloudMFAToken(client, noBrowser, credentials.Username, organizationId)
		if err != nil {
			return "", "", err
		}

		client = clientFactory.JwtHTTPClientFactory(context.Background(), token, url)
		err = a.checkMFAEmailMatchesLogin(client, credentials.Username)
		return token, refreshToken, err
	}

	client.HttpClient.Timeout = 30 * time.Second
	log.CliLogger.Debugf("Making login request for %s for org id %s", credentials.Username, organizationId)

	req := &ccloudv1.AuthenticateRequest{
		Email:         credentials.Username,
		Password:      credentials.Password,
		OrgResourceId: organizationId,
	}

	res, err := client.Auth.Login(req)
	if err != nil {
		return "", "", err
	}
	if res.GetOrganization().GetMfaEnforcedAt() != nil && res.GetUser().GetAuthType() == ccloudv1.AuthType_AUTH_TYPE_LOCAL {
		output.Printf(false, "Please be aware that you will be required to enroll in MFA by %s\n", convertDateFormat(res.GetOrganization().GetMfaEnforcedAt()))
	}

	if utils.IsOrgEndOfFreeTrialSuspended(res.GetOrganization().GetSuspensionStatus()) {
		log.CliLogger.Debugf(errors.EndOfFreeTrialErrorMsg, res.GetOrganization().GetSuspensionStatus())
		return res.GetToken(), res.GetRefreshToken(), &errors.EndOfFreeTrialError{OrgId: res.GetOrganization().GetName()}
	}

	return res.GetToken(), res.GetRefreshToken(), nil
}

func (a *AuthTokenHandlerImpl) getCCloudSSOToken(client *ccloudv1.Client, noBrowser bool, email, organizationId string) (string, string, error) {
	connectionName, err := a.getSsoConnectionName(client, email, organizationId)
	if err != nil {
		log.CliLogger.Debugf("unable to obtain user SSO info: %v", err)
		return "", "", fmt.Errorf(`unable to obtain SSO info for user "%s"`, email)
	}
	if connectionName == "" {
		return "", "", fmt.Errorf(`tried to obtain SSO token for non SSO user "%s"`, email)
	}

	idToken, refreshToken, err := sso.Login(client.BaseURL, noBrowser, connectionName)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: organizationId,
	}
	res, err := login(client, req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
}

func (a *AuthTokenHandlerImpl) getCCloudMFAToken(client *ccloudv1.Client, noBrowser bool, email, organizationId string) (string, string, error) {
	connectionName, err := a.getMfaConnectionName(client, email, organizationId)
	if err != nil {
		return "", "", fmt.Errorf(`unable to obtain MFA info for user "%s: %v"`, email, err)
	}
	if connectionName == "" {
		return "", "", fmt.Errorf(`tried to obtain MFA token for non MFA user "%s"`, email)
	}

	idToken, refreshToken, err := mfa.Login(client.BaseURL, email, connectionName, noBrowser)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: organizationId,
	}
	res, err := client.Auth.Login(req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
}

func (a *AuthTokenHandlerImpl) getSsoConnectionName(client *ccloudv1.Client, email, organizationId string) (string, error) {
	req := &ccloudv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      sso.GetAuth0CCloudClientIdFromBaseUrl(client.BaseURL),
		OrgResourceId: organizationId,
	}
	loginRealmReply, err := client.User.LoginRealm(req)
	if err != nil {
		return "", err
	}
	if loginRealmReply.GetIsSso() {
		return loginRealmReply.GetRealm(), nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) getMfaConnectionName(client *ccloudv1.Client, email, organizationId string) (string, error) {
	req := &ccloudv1.GetLoginRealmRequest{
		Email:         email,
		ClientId:      sso.GetAuth0CCloudClientIdFromBaseUrl(client.BaseURL),
		OrgResourceId: organizationId,
	}
	loginRealmReply, err := client.User.LoginRealm(req)
	if err != nil {
		return "", err
	}
	if loginRealmReply.GetMfaRequired() {
		return loginRealmReply.GetRealm(), nil
	}
	return "", nil
}

func (a *AuthTokenHandlerImpl) refreshCCloudSSOToken(client *ccloudv1.Client, refreshToken, organizationId string) (string, string, error) {
	idToken, refreshToken, err := sso.RefreshTokens(client.BaseURL, refreshToken)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: organizationId,
	}

	res, err := login(client, req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
}
func (a *AuthTokenHandlerImpl) refreshCCloudMFAToken(client *ccloudv1.Client, refreshToken, organizationId, email string) (string, string, error) {
	idToken, refreshToken, err := mfa.RefreshTokens(client.BaseURL, refreshToken, email)
	if err != nil {
		return "", "", err
	}

	req := &ccloudv1.AuthenticateRequest{
		IdToken:       idToken,
		OrgResourceId: organizationId,
	}

	res, err := login(client, req)
	if err != nil {
		return "", "", err
	}

	return res.GetToken(), refreshToken, err
}

func (a *AuthTokenHandlerImpl) GetConfluentToken(mdsClient *mdsv1.APIClient, credentials *Credentials, noBrowser bool) (string, string, error) {
	ctx := utils.GetContext()
	if credentials.IsCertificateOnly {
		resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(context.Background())
		if err != nil {
			return "", "", err
		}
		return resp.AuthToken, "", nil
	}
	if credentials.IsSSO {
		if credentials.AuthRefreshToken != "" {
			if authToken, err := refreshConfluentToken(mdsClient, credentials); err == nil {
				return authToken, credentials.AuthRefreshToken, nil
			}
		}

		resp, _, err := mdsClient.SSODeviceAuthorizationApi.Security10OidcDeviceAuthenticatePost(context.Background())
		if err != nil {
			return "", "", err
		}

		if noBrowser {
			output.Println(false, "Navigate to the following link in your browser to authenticate:")
			output.Println(false, resp.VerificationUri)
		} else {
			output.Printf(false, "You will now be redirected to the SSO authentication page in your browser. Your activation code is \"%s\". If you are not redirected, please use the following link: %s\n", resp.UserCode, resp.VerificationUri)
			if err := form.ConfirmEnter(); err != nil {
				return "", "", err
			}
			if err := browser.OpenURL(resp.VerificationUri); err != nil {
				return "", "", fmt.Errorf("unable to open web browser for SSO authentication: %w", err)
			}
		}

		checkDeviceAuthRequest := mdsv1.CheckDeviceAuthRequest{
			UserCode: resp.UserCode,
			Key:      resp.Key,
		}

		var authToken, refreshToken string
		err = retry.Retry(time.Duration(resp.Interval)*time.Second, time.Duration(resp.ExpiresIn)*time.Second, func() error {
			checkAuthResp, _, err := mdsClient.SSODeviceAuthorizationApi.CheckDeviceAuth(context.Background(), checkDeviceAuthRequest)
			if err != nil {
				return fmt.Errorf("%s: %w", checkAuthResp.Error, err)
			}

			if !checkAuthResp.Complete {
				return fmt.Errorf("%s: %s", checkAuthResp.Status, checkAuthResp.Description)
			}

			authToken = checkAuthResp.AuthToken
			refreshToken = checkAuthResp.RefreshToken
			return nil
		})
		if err != nil {
			return "", "", err
		}

		return authToken, refreshToken, nil
	} else {
		basicAuth := mdsv1.BasicAuth{
			UserName: credentials.Username,
			Password: credentials.Password,
		}
		basicContext := context.WithValue(ctx, mdsv1.ContextBasicAuth, basicAuth)
		resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
		if err != nil {
			return "", "", err
		}
		return resp.AuthToken, "", nil
	}
}

func refreshConfluentToken(mdsClient *mdsv1.APIClient, credentials *Credentials) (string, error) {
	req := mdsv1.ExtendAuthRequest{
		AccessToken:  credentials.AuthToken,
		RefreshToken: credentials.AuthRefreshToken,
	}
	resp, _, err := mdsClient.SSODeviceAuthorizationApi.ExtendDeviceAuth(context.Background(), req)
	if err != nil {
		return "", err
	}

	return resp.AuthToken, nil
}

func (a *AuthTokenHandlerImpl) checkMFAEmailMatchesLogin(client *ccloudv1.Client, loginEmail string) error {
	getMeReply, err := client.Auth.User()
	if err != nil {
		return err
	}
	if !strings.EqualFold(getMeReply.GetUser().GetEmail(), loginEmail) {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf("expected login credentials for %s but got credentials for %s", loginEmail, getMeReply.GetUser().GetEmail()),
			"Please re-login and use the same email at the prompt and in the login portal.",
		)
	}
	return nil
}

func (a *AuthTokenHandlerImpl) checkSSOEmailMatchesLogin(client *ccloudv1.Client, loginEmail string) error {
	getMeReply, err := client.Auth.User()
	if err != nil {
		return err
	}
	if !strings.EqualFold(getMeReply.GetUser().GetEmail(), loginEmail) {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf("expected SSO credentials for %s but got credentials for %s", loginEmail, getMeReply.GetUser().GetEmail()),
			"Please re-login and use the same email at the prompt and in the SSO portal.",
		)
	}
	return nil
}

func login(client *ccloudv1.Client, req *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
	if sso.IsOkta(client.BaseURL) {
		return client.Auth.OktaLogin(req)
	} else {
		return client.Auth.Login(req)
	}
}

func convertDateFormat(mfaEnforcedAt *types.Timestamp) string {
	if mfaEnforcedAt == nil {
		return "Invalid Date"
	}
	date := time.Unix(mfaEnforcedAt.Seconds, int64(mfaEnforcedAt.Nanos)).UTC().Truncate(time.Microsecond)
	return date.Format("01/02/2006")
}
