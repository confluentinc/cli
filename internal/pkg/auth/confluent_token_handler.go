//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst mock/confluent_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli confluent_token_handler.go ConfluentTokenHandler
package auth

import (
	"context"

	mds "github.com/confluentinc/mds-sdk-go"
)

type ConfluentTokenHandler interface {
	GetAuthToken(mdsClient *mds.APIClient, email string, password string) (string, error)
}

type ConfluentTokenHandlerImp struct{}

func (c *ConfluentTokenHandlerImp) GetAuthToken(mdsClient *mds.APIClient, email string, password string) (string, error) {
	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: email, Password: password})
	resp, _, err := mdsClient.TokensAndAuthenticationApi.GetToken(basicContext)
	if err != nil {
		return "", err
	}
	return resp.AuthToken, nil
}
