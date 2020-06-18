//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/update_token_handler.go --pkg mock --selfpkg github.com/confluentinc/cli update_token_handler.go UpdateTokenHandler

package auth

import (
	"github.com/confluentinc/ccloud-sdk-go"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type UpdateTokenHandler interface {
	UpdateCCloudAuthTokenUsingNetrcCredentials(ctx *v3.Context, userAgent string, logger *log.Logger) error
	UpdateConfluentAuthTokenUsingNetrcCredentials(ctx *v3.Context, logger *log.Logger) error
}

type UpdateTokenHandlerImpl struct {
	ccloudTokenHandler    CCloudTokenHandler
	confluentTokenHandler ConfluentTokenHandler
	netrcHandler          *NetrcHandler
}

var (
	failedRefreshTokenMsg  = "Failed to update auth token using refresh token. Error: %s"
	failedNetrcTokenMsg    = "Failed to update auth token using credentials in netrc file. Error: %s"
	successRefreshTokenMsg = "Token successfully updated with refresh token."
	successNetrcTokenMsg   = "Token successfully updated with netrc file credentials."
)

func NewUpdateTokenHandler(netrcHandler *NetrcHandler) UpdateTokenHandler {
	return &UpdateTokenHandlerImpl{
		ccloudTokenHandler:    &CCloudTokenHandlerImpl{},
		confluentTokenHandler: &ConfluentTokenHandlerImp{},
		netrcHandler:          netrcHandler,
	}
}

func (u *UpdateTokenHandlerImpl) UpdateCCloudAuthTokenUsingNetrcCredentials(ctx *v3.Context, userAgent string, logger *log.Logger) error {
	url := ctx.Platform.Server
	client := ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	userSSO, err := u.ccloudTokenHandler.GetUserSSO(client, ctx.Credential.Username)
	if err != nil {
		logger.Debugf("Failed to get userSSO for user email: %s.", ctx.Credential.Username)
	}
	var token string
	if userSSO != nil {
		_, refreshToken, err := u.netrcHandler.getNetrcCredentials("ccloud", true, ctx.Name)
		if err != nil {
			logger.Debugf(failedRefreshTokenMsg, err)
			return err
		}
		token, err = u.ccloudTokenHandler.RefreshSSOToken(client, refreshToken, url)
		if err != nil {
			logger.Debugf(failedRefreshTokenMsg, err)
			return err
		}
		logger.Debug(successRefreshTokenMsg)
	} else {
		email, password, err := u.netrcHandler.getNetrcCredentials(ctx.Config.CLIName, false, ctx.Name)
		if err != nil {
			logger.Debugf(err.Error())
			return err
		}
		token, err = u.ccloudTokenHandler.GetCredentialsToken(client, email, password)
		if err != nil {
			logger.Debugf(failedNetrcTokenMsg, err)
			return err
		}
		logger.Debug(successNetrcTokenMsg)
	}
	return ctx.UpdateAuthToken(token)
}

func (u *UpdateTokenHandlerImpl) UpdateConfluentAuthTokenUsingNetrcCredentials(ctx *v3.Context, logger *log.Logger) error {
	email, password, err := u.netrcHandler.getNetrcCredentials("confluent", false, ctx.Name)
	if err != nil {
		logger.Debugf(err.Error())
		return err
	}
	mdsClientManager := MDSClientManagerImpl{}
	mdsClient, err := mdsClientManager.GetMDSClient(ctx, ctx.Platform.CaCertPath, false, ctx.Platform.Server, logger)
	token, err := u.confluentTokenHandler.GetAuthToken(mdsClient, email, password)
	if err != nil {
		logger.Debugf(failedNetrcTokenMsg, err)
		return err
	}
	err = ctx.UpdateAuthToken(token)
	if err == nil {
		logger.Debugf(successNetrcTokenMsg)
	}
	return err
}
