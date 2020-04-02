package auth

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	// For integration test
	NetrcIntegrationTestFile = "/tmp/netrc_test"

	confluentCliName             = "confluent-cli"
	mdsUsernamePasswordString    = "mds-username-password"
	ccloudUsernamePasswordString = "ccloud-username-password"
	ccloudSSORefreshTokenString  = "ccloud-sso-refresh-token"

	resolvingFilePathErrMsg = "An error resolving the netrc filepath at %s has occurred. Error: %s"
	netrcGetErrorMsg        = "Unable to get credentials from Netrc file. Error: %s"
	netrcWriteErrorMsg      = "Unable to write credentials to Netrc file. Error: %s"
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

func GetNetrcFilePath(isIntegrationTest bool) string {
	if isIntegrationTest {
		return NetrcIntegrationTestFile
	}
	if runtime.GOOS == "windows" {
		return "~/_netrc"
	} else {
		return "~/.netrc"
	}
}

func NewNetrcHandler(netrcFilePath string) *NetrcHandler {
	return &NetrcHandler{FileName: netrcFilePath}
}

type NetrcHandler struct {
	FileName string
}

func (n *NetrcHandler) WriteNetrcCredentials(cliName string, isSSO bool, ctxName string, username string, password string) error {
	filename, err := homedir.Expand(n.FileName)
	if err != nil {
		return fmt.Errorf(resolvingFilePathErrMsg, filename, err)
	}

	netrcFile, err := getOrCreateNetrc(filename)
	if err != nil {
		return fmt.Errorf(netrcWriteErrorMsg, err)
	}

	machineName := getNetrcMachineName(cliName, isSSO, ctxName)

	machine := netrcFile.FindMachine(machineName)
	if machine == nil {
		machine = netrcFile.NewMachine(machineName, username, password, "")
	} else {
		machine.UpdateLogin(username)
		machine.UpdatePassword(password)
	}
	netrcBytes, err := netrcFile.MarshalText()
	if err != nil {
		return fmt.Errorf(netrcWriteErrorMsg, err)
	}
	err = ioutil.WriteFile(filename, netrcBytes, 0600)
	if err != nil {
		return fmt.Errorf("Unable to write to netrc file %s. Error: %s", filename, err)
	}
	return nil
}

// for username-password credentials the return values are self-explanatory but for sso case the password is the refreshToken
func (n *NetrcHandler) getNetrcCredentials(cliName string, isSSO bool, ctxName string) (username string, password string, err error) {
	filename, err := homedir.Expand(n.FileName)
	if err != nil {
		return "", "", fmt.Errorf(resolvingFilePathErrMsg, filename, err)
	}
	machineName := getNetrcMachineName(cliName, isSSO, ctxName)
	machine, err := netrc.FindMachine(filename, machineName)
	if err != nil {
		return "", "", fmt.Errorf(netrcGetErrorMsg, err)
	}
	if machine == nil {
		return "", "", errors.Errorf("Login credential not in netrc file.")
	}
	return machine.Login, machine.Password, nil
}

func getNetrcMachineName(cliName string, isSSO bool, ctxName string) string {
	var credType netrcCredentialType
	if cliName == "confluent" {
		credType = mdsUsernamePassword
	} else {
		if isSSO {
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
