package auth

import (
	"context"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/sso"
)

var (
	netrcfile = "~/.netrc"
	machinePrefix = "confluent-cli:"
)

func UpdateNetrc(ctx *v3.Context, username string, password string) {
	machine, _ := netrc.FindMachine(netrcfile, machinePrefix + ctx.Name)
	machine.UpdateLogin(username)
	machine.UpdatePassword(password)
}

func getEmailPassword(ctxName string) (string, string, error){
	machine, err := netrc.FindMachine(netrcfile, machinePrefix + ctxName)
	if err != nil {
		return "", "", err
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
	token, err := GetAuthToken(client, url, email, password, false)
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


func GetAuthToken(client *ccloud.Client, url string, email string, password string, noBrowser bool) (string, error) {

	// Check if user has an enterprise SSO connection enabled.
	userSSO, err := client.User.CheckEmail(context.Background(), &orgv1.User{Email: email})
	if err != nil {
		return "", err
	}

	token := ""

	if userSSO != nil && userSSO.Sso != nil && userSSO.Sso.Enabled && userSSO.Sso.Auth0ConnectionName != "" {
		idToken, err := sso.Login(url, noBrowser, userSSO.Sso.Auth0ConnectionName)
		if err != nil {
			return "", err
		}

		token, err = client.Auth.Login(context.Background(), idToken, "", "")
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
