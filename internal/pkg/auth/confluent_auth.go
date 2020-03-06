package auth

import (
	"context"
	"github.com/confluentinc/mds-sdk-go"
)

func GetConfluentAuthToken(mdsClient *mds.APIClient, email string, password string) (string, error){
	basicContext := context.WithValue(context.Background(), mds.ContextBasicAuth, mds.BasicAuth{UserName: email, Password: password})
	resp, _, err := mdsClient.TokensAuthenticationApi.GetToken(basicContext, "")
	if err != nil {
		return "", err
	}
	return resp.AuthToken, nil
}
