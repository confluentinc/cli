package auth

import (
	"fmt"
	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/confluentinc/ccloud-sdk-go"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	netrcfile = "~/.netrc"
)

func getEmailPassword(ctxName string) (string, string, error){
	filename, err := homedir.Expand(netrcfile)
	if err != nil {
		err = fmt.Errorf("an error resolving the netrc filepath at %s has occurred. ", filename)
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

func UpdateAuthToken(ctx *v3.Context, userAgent string) error {
	email, password, err := getEmailPassword(ctx.Name)
	if err != nil {
		return err
	}
	url := ctx.Platform.Server
	client := ccloud.NewClient(&ccloud.Params{BaseURL: url, HttpClient: ccloud.BaseClient, Logger: ctx.Logger, UserAgent: userAgent})
	token, err := GetCCloudAuthToken(client, url, email, password, false)
	if err != nil {
		return err
	}
	ctx.State.AuthToken = token
	err = ctx.Save()
	if err != nil {
		return err
	}
	return nil
}
