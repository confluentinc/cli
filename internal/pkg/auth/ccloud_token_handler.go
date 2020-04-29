//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst mock/ccloud_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli ccloud_token_handler.go CCloudTokenHandler
package auth

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/internal/pkg/sso"
)

// Make into interface in order to create mock for testing
type CCloudTokenHandler interface {
	GetUserSSO(client *ccloud.Client, email string) (*orgv1.User, error)
	GetCredentialsToken(client *ccloud.Client, email string, password string) (string, error)
	GetSSOToken(client *ccloud.Client, url string, noBrowser bool, userSSO *orgv1.User) (string, string, error)
	RefreshSSOToken(client *ccloud.Client, refreshToken string, url string) (string, error)
}

type CCloudTokenHandlerImpl struct{}

func (c *CCloudTokenHandlerImpl) GetUserSSO(client *ccloud.Client, email string) (*orgv1.User, error) {
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return nil, err
	}
	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		return userSSO, nil
	}
	return nil, nil
}

func (c *CCloudTokenHandlerImpl) GetCredentialsToken(client *ccloud.Client, email string, password string) (string, error) {
	return client.Auth.Login(context.Background(), "", email, password)
}

func (c *CCloudTokenHandlerImpl) GetSSOToken(client *ccloud.Client, url string, noBrowser bool, userSSO *orgv1.User) (string, string, error) {
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

func (c *CCloudTokenHandlerImpl) RefreshSSOToken(client *ccloud.Client, refreshToken string, url string) (string, error) {
	idToken, err := sso.GetNewIDTokenFromRefreshToken(url, refreshToken)
	if err != nil {
		return "", err
	}
	token, err := client.Auth.Login(context.Background(), idToken, "", "")
	if err != nil {
		return "", err
	}
	return token, nil
}
