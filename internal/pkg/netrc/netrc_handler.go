//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../mock/netrc_handler.go --pkg mock --selfpkg github.com/confluentinc/cli netrc_handler.go NetrcHandler
package netrc

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/atrox/homedir"

	"github.com/confluentinc/cli/internal/pkg/errors"
	gonetrc "github.com/confluentinc/go-netrc/netrc"
)

const (
	// For integration test
	NetrcIntegrationTestFile = "/tmp/netrc_test"

	netrcCredentialStringFormat  = "confluent-cli:%s:%s"
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
	GetMatchingNetrcMachine(params GetMatchingNetrcMachineParams) (*Machine, error)
	GetFileName() string
}

type GetMatchingNetrcMachineParams struct {
	CLIName string
	IsSSO   bool
	CtxName string
	URL     string
}

type Machine struct {
	Name     string
	User     string
	Password string
	IsSSO    bool
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

func getOrCreateNetrc(filename string) (*gonetrc.Netrc, error) {
	n, err := gonetrc.ParseFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.OpenFile(filename, os.O_CREATE, 0600)
			if err != nil {
				return nil, errors.Wrapf(err, errors.CreateNetrcFileErrorMsg, filename)
			}
			n, err = gonetrc.ParseFile(filename)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return n, nil
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
	return fmt.Sprintf(netrcCredentialStringFormat, credType.String(), ctxName)
}

// Using the parameters to filter and match machine name
// Returns the first match
// For SSO case the password is the refreshToken
func (n *NetrcHandlerImpl) GetMatchingNetrcMachine(params GetMatchingNetrcMachineParams) (*Machine, error) {
	filename, err := homedir.Expand(n.FileName)
	if err != nil {
		return nil, errors.Wrapf(err, errors.ResolvingNetrcFilepathErrorMsg, filename)
	}
	if params.CLIName == "" {
		return nil, errors.New(errors.NetrcCLINameMissingErrorMsg)
	}
	machines, err := gonetrc.GetMachines(filename)
	if err != nil {
		return nil, err
	}

	regex := getMachineNameRegex(params)
	for _, machine := range machines {
		if regex.Match([]byte(machine.Name)) {
			return &Machine{Name: machine.Name, User: machine.Login, Password: machine.Password, IsSSO: isSSOMachine(machine.Name)}, nil
		}
	}

	return nil, nil
}

func getMachineNameRegex(params GetMatchingNetrcMachineParams) *regexp.Regexp {
	var contextNameRegex string
	if params.CtxName != "" {
		contextNameRegex = escapeSpecialRegexChars(params.CtxName)
	} else if params.URL != "" {
		url := strings.ReplaceAll(params.URL, ".", `\.`)
		contextNameRegex = fmt.Sprintf(".*-%s", url)
	} else {
		contextNameRegex = ".*"
	}

	var regexString string
	if params.CLIName == "ccloud" {
		if params.IsSSO {
			regexString = "^" + fmt.Sprintf(netrcCredentialStringFormat, ccloudSSORefreshTokenString, contextNameRegex)
		} else {
			// if isSSO is not True, we will check for both SSO and non SSO
			ccloudCreds := []string{ccloudUsernamePasswordString, ccloudSSORefreshTokenString}
			credTypeRegex := fmt.Sprintf("(%s)", strings.Join(ccloudCreds, "|"))
			regexString = "^" + fmt.Sprintf(netrcCredentialStringFormat, credTypeRegex, contextNameRegex)
		}
	} else {
		regexString = "^" + fmt.Sprintf(netrcCredentialStringFormat, mdsUsernamePasswordString, contextNameRegex)
	}

	return regexp.MustCompile(regexString)
}

func escapeSpecialRegexChars(s string) string {
	specialChars := `\^${}[]().*+?|<>-&`
	res := ""
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			res += `\`
		}
		res += string(c)
	}
	return res
}

func isSSOMachine(machineName string) bool {
	return strings.Contains(machineName, ccloudSSORefreshTokenString)
}

func (n *NetrcHandlerImpl) GetFileName() string {
	return n.FileName
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
