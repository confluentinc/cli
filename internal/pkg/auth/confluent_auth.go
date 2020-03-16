package auth

import (
	"context"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
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

func UpdateConfluentAuthToken(ctx *v3.Context, logger *log.Logger) error {
	email, password, err := getNetrcCredentials(ctx.Name)
	if err != nil {
		logger.Debugf(netrcErrorString, err.Error())
		return err
	}
	mdsClientManager := MDSClientManagerImpl{}
	mdsClient, err := mdsClientManager.GetMDSClient(ctx, ctx.Platform.CaCertPath, false, ctx.Platform.Server, logger)
	token, err := GetConfluentAuthToken(mdsClient, email, password)
	if err != nil {
		logger.Debugf("Failed to update auth token. Error: %s", err)
		return err
	}
	err = updateContext(ctx, token)
	if err == nil {
		logger.Debugf("Successfully updated auth token.")
	}
	return err
}
