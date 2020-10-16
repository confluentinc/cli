//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst mock/netrc_handler.go --pkg mock --selfpkg github.com/confluentinc/cli netrc_handler.go NetrcHandler
package auth

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/atrox/homedir"
	"github.com/csreesan/go-netrc/netrc"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	// For integration test
	NetrcIntegrationTestFile = "/tmp/netrc_test"

	confluentCliName             = "confluent-cli"
	mdsUsernamePasswordString    = "mds-username-password"
	ccloudUsernamePasswordString = "ccloud-username-password"
	ccloudSSORefreshTokenString  = "ccloud-sso-refresh-token"
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

type NetrcHandler interface {
	WriteNetrcCredentials(cliName string, isSSO bool, ctxName string, username string, password string) error
	GetNetrcCredentials(cliName string, isSSO bool, ctxName string) (username string, password string, err error)
	GetMatchingNetrcCredentials(cliName string, url string) (username string, password string, err error)
	GetFileName() string
}

func NewNetrcHandler(netrcFilePath string) *NetrcHandlerImpl {
	return &NetrcHandlerImpl{FileName: netrcFilePath}
}

type NetrcHandlerImpl struct {
	FileName string
}

func (n *NetrcHandlerImpl) WriteNetrcCredentials(cliName string, isSSO bool, ctxName string, username string, password string) error {
	filename, err := homedir.Expand(n.FileName)
	if err != nil {
		return errors.Wrapf(err, errors.ResolvingNetrcFilepathErrorMsg, filename)
	}

	netrcFile, err := getOrCreateNetrc(filename)
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, filename)
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
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, filename)
	}
	err = ioutil.WriteFile(filename, netrcBytes, 0600)
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, filename)
	}
	return nil
}

// for username-password credentials the return values are self-explanatory but for sso case the password is the refreshToken
func (n *NetrcHandlerImpl) GetNetrcCredentials(cliName string, isSSO bool, ctxName string) (username string, password string, err error) {
	filename, err := homedir.Expand(n.FileName)
	if err != nil {
		return "", "", errors.Wrapf(err, errors.ResolvingNetrcFilepathErrorMsg, filename)
	}
	machineName := getNetrcMachineName(cliName, isSSO, ctxName)
	machine, err := netrc.FindMachine(filename, machineName)
	if err != nil {
		return "", "", errors.Wrapf(err, errors.GetNetrcCredentialsErrorMsg, filename)
	}
	if machine == nil {
		return "", "", errors.Errorf(errors.NetrcCredentialsNotFoundErrorMsg, filename)
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
				return nil, errors.Wrapf(err, errors.CreateNetrcFileErrorMsg, filename)
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

func (n *NetrcHandlerImpl) GetFileName() string {
	return n.FileName
}

func (n *NetrcHandlerImpl) GetMatchingNetrcCredentials(cliName string, url string) (string, string, error) {
	netrcFile, err := netrc.ParseFile(n.FileName)
	if err != nil {
		return "", "", err
	}
	machines := netrcFile.GetMachines()
	for _, m := range machines {
		if isMatchingMachine(m, cliName, url) {
			return m.Login, m.Password, nil
		}
	}
	return "", "", nil
}

func isMatchingMachine(machine *netrc.Machine, cliName string, url string) bool {
	if cliName == "ccloud" {
		if strings.HasPrefix(machine.Name, "ccloud") && strings.HasSuffix(machine.Name, url) {
			return true
		}
	}
	return strings.HasPrefix(machine.Name, "mds") && strings.HasSuffix(machine.Name, url)
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
