package auth

import (
	"context"
	"fmt"
	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/sso"
	"io/ioutil"
	"os"
)

var (
	netrcfile = "~/.netrc"
	machinePrefix = "confluent-cli:"
)

func UpdateNetrc(ctxName string, username string, password string) error {
	filename, err := homedir.Expand(netrcfile)
	if err != nil {
		err = fmt.Errorf("an error resolving the netrc filepath at %s has occurred. "+
			"Please try moving the file to a different location", filename)
		return err
	}
	n, err := getOrCreateNetrc(filename)
	if err != nil {
		return err
	}
	machine := n.FindMachine(machinePrefix + ctxName)
	if machine == nil {
		machine = n.NewMachine(machinePrefix + ctxName, username, password, "")
	} else {
		machine.UpdateLogin(username)
		machine.UpdatePassword(password)
	}
	netrcBytes, err := n.MarshalText()
	err = ioutil.WriteFile(filename, netrcBytes, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write netrc file: %s", filename)
	}
	return err
}

func getOrCreateNetrc(filename string) (*netrc.Netrc, error) {
	n, err := netrc.ParseFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.OpenFile(filename, os.O_CREATE, 0600)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create netrc file: %s", filename)
			}
			n, err = netrc.ParseFile(filename)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return n, nil
}

func getEmailPassword(ctxName string) (string, string, error){
	filename, err := homedir.Expand(netrcfile)
	if err != nil {
		err = fmt.Errorf("an error resolving the netrc filepath at %s has occurred. "+
			"Please try moving the file to a different location", filename)
		return "", "", err
	}
	machine, err := netrc.FindMachine(filename, machinePrefix + ctxName)
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
