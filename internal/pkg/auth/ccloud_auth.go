package auth

import (
	"context"
	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/sso"
)

func GetCCloudAuthToken(client *ccloud.Client, url string, email string, password string, noBrowser bool) (string, error) {
	// Check if user has an enterprise SSO connection enabled.
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return "", err
	}
	token := ""
	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		token, err = getSSoToken(client, url, noBrowser, userSSO)
		if err != nil {
			return "", err
		}
	} else {
		token, err = client.Auth.Login(context.Background(), "", email, password)
		if err != nil {
			return "", err
		}
	}
	return token, nil
}

func getSSoToken(client *ccloud.Client, url string, noBrowser bool, userSSO *orgv1.User) (string, error) {
	idToken, err := sso.Login(url, noBrowser, userSSO.Sso.Auth0ConnectionName)
	if err != nil {
		return "", err
	}

	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetCCloudAuthTokenClient(ctx *v3.Context, userAgent string, url string) *ccloud.Client {
	return ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: ctx.Logger, UserAgent: userAgent})
}
