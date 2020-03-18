//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/update_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli update_token_handler.go UpdateTokenHandler

package auth

import (
	"github.com/confluentinc/ccloud-sdk-go"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type UpdateTokenHandler interface {
	UpdateCCloudAuthToken(ctx *v3.Context, userAgent string, logger *log.Logger) error
	UpdateConfluentAuthToken(ctx *v3.Context, logger *log.Logger) error
}

type UpdateTokenHandlerImpl struct {}

func (u *UpdateTokenHandlerImpl) UpdateCCloudAuthToken(ctx *v3.Context, userAgent string, logger *log.Logger) error {
	tokenHandler := CCloudTokenHandlerImpl{}
	url := ctx.Platform.Server
	client := ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	userSSO, err := tokenHandler.GetUserSSO(client, ctx.Credential.Username)
	if err != nil {
		logger.Debugf("Failed to get userSSO for user email: %s.", ctx.Credential.Username)
	}
	var token string
	if userSSO != nil {
		token, err = tokenHandler.RefreshSSOToken(client, ctx, url)
		if err != nil {
			logger.Debugf("Failed to update auth token using refresh token. Error: %s", err)
			return err
		}
		logger.Debug("Token successfully updated with refresh token.")
	} else {
		netrcHandler := netrcHandler{fileName:netrcfile}
		email, password, err := netrcHandler.getNetrcCredentials(ctx.Name)
		if err != nil {
			logger.Debugf(netrcErrorString, err.Error())
			return err
		}
		token, err = tokenHandler.GetCredentialsToken(client, email, password)
		if err != nil {
			logger.Debugf("Failed to update auth token using credentials in netrc file. Error: %s", err)
			return err
		}
		logger.Debug("Token successfully updated with netrc file credentials.")
	}
	return ctx.UpdateAuthToken(token)
}

func (u *UpdateTokenHandlerImpl) UpdateConfluentAuthToken(ctx *v3.Context, logger *log.Logger) error {
	netrcHandler := netrcHandler{fileName:netrcfile}
	email, password, err := netrcHandler.getNetrcCredentials(ctx.Name)
	if err != nil {
		logger.Debugf(netrcErrorString, err.Error())
		return err
	}
	mdsClientManager := MDSClientManagerImpl{}
	mdsClient, err := mdsClientManager.GetMDSClient(ctx, ctx.Platform.CaCertPath, false, ctx.Platform.Server, logger)
	tokenHandler := ConfluentTokenHandlerImp{}
	token, err := tokenHandler.GetAuthToken(mdsClient, email, password)
	if err != nil {
		logger.Debugf("Failed to update auth token. Error: %s", err)
		return err
	}
	err = ctx.UpdateAuthToken(token)
	if err == nil {
		logger.Debugf("Successfully updated auth token.")
	}
	return err
}
