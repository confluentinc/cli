package auth

import (
	"fmt"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	confluentCliName = "confluent-cli"
	mdsUsernamePasswordString = "mds-username-password"
	ccloudUsernamePasswordString = "ccloud-username-password"
	ccloudSSORefreshTokenString = "ccloud-sso-refreshtoken"

	resolvingFilePathErrMsg = "An error resolving the netrc filepath at %s has occurred. Error: %s"
	netrcErrorMsg = "Unable to get credentials from Netrc file. Error: %s"
)

type netrcCredentialType int

const (
	mdsUsernamePassword netrcCredentialType = iota
	ccloudUsernamePassword
	ccloudSSORefreshToken
)

func (c netrcCredentialType) String() string {
	credTypes := [...]string{mdsUsernamePasswordString, ccloudUsernamePasswordString, ccloudSSORefreshTokenString}
	return credTypes[c]
}

func NewNetrcHandler() *netrcHandler {
	var netrcFile string
	if runtime.GOOS == "windows" {
		netrcFile = "~/_netrc"
	} else {
		netrcFile = "~/.netrc"
	}
	return &netrcHandler{fileName:netrcFile}
}

type netrcHandler struct {
	fileName string
}

func (n *netrcHandler) WriteNetrcCredentials(ctx *v3.Context, username string, password string, sso bool) error {
	filename, err := homedir.Expand(n.fileName)
	if err != nil {
		return fmt.Errorf(resolvingFilePathErrMsg, filename, err)
	}

	netrcFile, err := getOrCreateNetrc(filename)
	if err != nil {
		return err
	}

	machineName := getNetrcMachineName(ctx.Config.CLIName, sso, ctx.Name)

	machine := netrcFile.FindMachine(machineName)
	if machine == nil {
		machine = netrcFile.NewMachine(machineName, username, password, "")
	} else {
		machine.UpdateLogin(username)
		machine.UpdatePassword(password)
	}
	netrcBytes, err := netrcFile.MarshalText()
	err = ioutil.WriteFile(filename, netrcBytes, 0600)
	if err != nil {
		return errors.Wrapf(err, "Unable to write to netrc file %s. Error: %s", filename, err)
	}
	return nil
}

func (n *netrcHandler) getNetrcCredentials(ctxName string) (username string, password string, err error) {
	filename, err := homedir.Expand(n.fileName)
	if err != nil {
		return "", "", fmt.Errorf(resolvingFilePathErrMsg, filename, err)
	}
	machine, err := netrc.FindMachine(filename, ctxName)
	if err != nil {
		return "", "", fmt.Errorf(netrcErrorMsg, err)
	}
	if machine == nil {
		return "", "", errors.Errorf("Login credential not in netrc file.")
	}
	return machine.Login, machine.Password, nil
}

func (n *netrcHandler) getRefreshToken(ctxName string) (refreshToken string, err error) {
	_, refreshToken, err = n.getNetrcCredentials(ctxName)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func getNetrcMachineName(cliName string, sso bool, ctxName string) string {
	var credType netrcCredentialType
	if cliName == "confluent" {
		credType = mdsUsernamePassword
	} else {
		if sso {
			credType = ccloudSSORefreshToken
		} else {
			credType = ccloudUsernamePassword
		}
	}
	return strings.Join([]string{confluentCliName, credType.String(), ctxName}, ":")
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
