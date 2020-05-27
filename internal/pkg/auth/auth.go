package auth

import (
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

// If user is sso then will return refresh token, but if user is email password login then refresh token is ""
func GetCCloudAuthToken(client *ccloud.Client, url string, email string, password string, noBrowser bool) (string, string, error) {
	tokenHandler := CCloudTokenHandlerImpl{}
	userSSO, err := tokenHandler.GetUserSSO(client, email)
	if err != nil {
		return "", "", err
	}
	token := ""
	refreshToken := ""
	// Check if user has an enterprise SSO connection enabled.
	if userSSO != nil {
		token, refreshToken, err = tokenHandler.GetSSOToken(client, url, noBrowser, userSSO)
	} else {
		token, err = tokenHandler.GetCredentialsToken(client, email, password)
	}
	if err != nil {
		return "", "", err
	}
	return token, refreshToken, nil
}

func GetConfluentAuthToken(mdsClient *mds.APIClient, email string, password string) (string, error) {
	tokenHandler := ConfluentTokenHandlerImp{}
	return tokenHandler.GetAuthToken(mdsClient, email, password)
}
