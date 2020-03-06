package auth

import (
	"fmt"
	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/ccloud-sdk-go"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	netrcfile = "~/.netrc"
	netrcErrorString = "Unable to get email and password from Netrc file: %s"
)

func getEmailPassword(ctxName string) (string, string, error){
	filename, err := homedir.Expand(netrcfile)
	if err != nil {
		err = fmt.Errorf("an error resolving the Netrc filepath at %s has occurred. ", filename)
		return "", "", err
	}
	machine, err := netrc.FindMachine(filename, ctxName)
	if err != nil {
		return "", "", err
	}
	if machine == nil {
		return "", "", errors.Errorf("Login credential not in netrc file.")
	}
	return machine.Login, machine.Password, nil
}

func UpdateCCloudAuthToken(ctx *v3.Context, userAgent string, logger *log.Logger) error {
	email, password, err := getEmailPassword(ctx.Name)
	if err != nil {
		logger.Debugf(netrcErrorString, err.Error())
		return err
	}
	url := ctx.Platform.Server
	client := ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	token, err := GetCCloudAuthToken(client, url, email, password, false)
	if err != nil {
		return err
	}
	return updateContext(ctx, token)
}

func UpdateConfluentAuthToken(ctx *v3.Context, logger *log.Logger) error {
	email, password, err := getEmailPassword(ctx.Name)
	if err != nil {
		logger.Debugf(netrcErrorString, err.Error())
		return err
	}
	mdsClientManager := MDSClientManagerImpl{}
	mdsClient, err := mdsClientManager.GetMDSClient(ctx, ctx.Platform.CaCertPath, false, ctx.Platform.Server, logger)
	token, err := GetConfluentAuthToken(mdsClient, email, password)
	if err != nil {
		return err
	}
	return updateContext(ctx, token)
}

func updateContext(ctx *v3.Context, token string) error {
	ctx.State.AuthToken = token
	err := ctx.Save()
	if err != nil {
		return err
	}
	return nil
}
