package auth

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/sso"
)

func GetCCloudAuthToken(client *ccloud.Client, url string, email string, password string, noBrowser bool) (string, string, error) {
	userSSO, err := getUserSSO(client, email)
	if err != nil {
		return "", "", err
	}
	token := ""
	refreshToken := ""
	// Check if user has an enterprise SSO connection enabled.
	if userSSO != nil {
		token, refreshToken, err = getSSOToken(client, url, noBrowser, userSSO)
	} else {
		token, err = getCredentialsToken(client, email, password)
	}
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func UpdateCCloudAuthToken(ctx *v3.Context, userAgent string, logger *log.Logger) error {
	url := ctx.Platform.Server
	client := ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	userSSO, err := getUserSSO(client, ctx.Credential.Username)
	if err != nil {
		logger.Debugf("Failed to get userSSO for user email: %s.", ctx.Credential.Username)
	}
	var token string
	if userSSO != nil {
		token, err = refreshSSOToken(client, ctx, url)
		if err != nil {
			logger.Debugf("Failed to update auth token using refresh token. Error: %s", err)
			return err
		}
		logger.Debug("Token successfully updated with refresh token.")
	} else {
		email, password, err := getNetrcCredentials(ctx.Name)
		if err != nil {
			logger.Debugf(netrcErrorString, err.Error())
			return err
		}
		token, err = getCredentialsToken(client, email, password)
		if err != nil {
			logger.Debugf("Failed to update auth token using credentials in netrc file. Error: %s", err)
			return err
		}
		logger.Debug("Token successfully updated with netrc file credentials.")
	}
	return updateContext(ctx, token)
}

func getUserSSO(client *ccloud.Client, email string) (*orgv1.User, error) {
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return nil, err
	}
	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		return userSSO, nil
	}
	return nil, nil
}

func getCredentialsToken(client *ccloud.Client, email string, password string) (string, error) {
	return client.Auth.Login(context.Background(), "", email, password)
}

func getSSOToken(client *ccloud.Client, url string, noBrowser bool, userSSO *orgv1.User) (string, string, error) {
	idToken, refreshToken, err := sso.Login(url, noBrowser, userSSO.Sso.Auth0ConnectionName)
	if err != nil {
		return "", "", err
	}

	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func refreshSSOToken(client *ccloud.Client, ctx *v3.Context, url string) (string, error) {
	idToken, err := sso.GetNewIDTokenFromRefreshToken(url, ctx.State.RefreshToken)
	if err != nil {
		return "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", err
	}
	return token, nil
}
