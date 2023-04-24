//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../mock/netrc_handler.go --pkg mock --selfpkg github.com/confluentinc/cli netrc_handler.go NetrcHandler
package netrc

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	gonetrc "github.com/confluentinc/go-netrc/netrc"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	// For integration test
	NetrcIntegrationTestFile = "/tmp/netrc_test"

	localCredentialsPrefix       = "confluent-cli"
	localCredentialStringFormat  = localCredentialsPrefix + ":%s:%s"
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
	RemoveNetrcCredentials(isCloud bool, ctxName string) (string, error)
	CheckCredentialExist(isCloud bool, ctxName string) (bool, error)
	GetMatchingNetrcMachine(params NetrcMachineParams) (*Machine, error)
	GetFileName() string
}

type NetrcMachineParams struct {
	IgnoreCert bool
	IsCloud    bool
	IsSSO      bool
	Name       string
	URL        string
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

func (n *NetrcHandlerImpl) RemoveNetrcCredentials(isCloud bool, ctxName string) (string, error) {
	netrcFile, err := getNetrc(n.FileName)
	if err != nil {
		return "", err
	}

	// machineName could be either sso: confluent-cli:ccloud-sso-refresh-token:login-cli-mock-email@confluent.io-http://test
	// or non-sso: confluent-cli:ccloud-username-password:login-cli-mock-email@confluent.io-http://test
	var user string
	found := false
	for _, isSSO := range []bool{true, false} {
		machineName := GetLocalCredentialName(isCloud, isSSO, ctxName)
		machine := netrcFile.FindMachine(machineName)
		if machine != nil {
			found = true
			err := removeCredentials(machineName, netrcFile, n.FileName)
			if err != nil {
				return "", err
			}
			user = machine.Login
		}
	}
	if !found {
		err = errors.New(errors.NetrcCredentialsNotFoundErrorMsg)
		return "", err
	}
	return user, nil
}

func removeCredentials(machineName string, netrcFile *gonetrc.Netrc, filename string) error {
	netrcBytes, err := netrcFile.MarshalText()
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, filename)
	}
	var stringBuf []string
	lines := strings.Split(string(netrcBytes), "\n")
	length := len(lines)
	for i := 0; i < length; i++ {
		if strings.Contains(lines[i], machineName) {
			count := 3 // remove 3 non-empty lines or credentials
			for count > 0 {
				if lines[i] != "" {
					count -= 1
				}
				i += 1
			}
		}
		if i < length {
			stringBuf = append(stringBuf, lines[i]+"\n")
		}
	}
	if len(stringBuf) > 0 && stringBuf[len(stringBuf)-1] == "\n" { // remove the last newline
		stringBuf = stringBuf[:len(stringBuf)-1]
	}
	joinedString := strings.Join(stringBuf[:], "")
	joinedString = strings.Replace(joinedString, "\n\n", "\n", -1)
	buf := []byte(joinedString)
	// get file mode
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	filemode := info.Mode()
	err = os.WriteFile(filename, buf, filemode)
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, filename)
	}
	return nil
}

func getNetrc(filename string) (*gonetrc.Netrc, error) {
	n, err := gonetrc.ParseFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(err, errors.NetrcCredentialsNotFoundErrorMsg, filename)
		} else {
			return nil, err // failed to parse the netrc file due to other reasons
		}
	}
	return n, nil
}

func GetLocalCredentialName(isCloud bool, isSSO bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		if isSSO {
			credType = ccloudSSORefreshToken
		} else {
			credType = ccloudUsernamePassword
		}
	}
	return fmt.Sprintf(localCredentialStringFormat, credType.String(), ctxName)
}

// Using the parameters to filter and match machine name
// Returns the first match
// For SSO case the password is the refreshToken
func (n *NetrcHandlerImpl) GetMatchingNetrcMachine(params NetrcMachineParams) (*Machine, error) {
	machines, err := gonetrc.GetMachines(n.FileName)
	if err != nil {
		return nil, err
	}

	regex := getMachineNameRegex(params)
	// Look for the most recent entry matching the regex
	for i := len(machines) - 1; i >= 0; i-- {
		machine := machines[i]
		if regex.Match([]byte(machine.Name)) {
			return &Machine{Name: machine.Name, User: machine.Login, Password: machine.Password, IsSSO: isSSOMachine(machine.Name)}, nil
		}
	}

	return nil, nil
}

func getMachineNameRegex(params NetrcMachineParams) *regexp.Regexp {
	var contextNameRegex string
	if params.Name != "" {
		contextNameRegex = escapeSpecialRegexChars(params.Name)
	} else if params.URL != "" {
		url := strings.ReplaceAll(params.URL, ".", `\.`)
		contextNameRegex = fmt.Sprintf(".*-%s", url)
	} else {
		contextNameRegex = ".*"
	}

	if params.IgnoreCert {
		if idx := strings.Index(contextNameRegex, `\?cacertpath=`); idx != -1 {
			contextNameRegex = contextNameRegex[:idx] + ".*"
		}
	}

	var regexString string
	if params.IsCloud {
		if params.IsSSO {
			regexString = "^" + fmt.Sprintf(localCredentialStringFormat, ccloudSSORefreshTokenString, contextNameRegex)
		} else {
			// if isSSO is not True, we will check for both SSO and non SSO
			ccloudCreds := []string{ccloudUsernamePasswordString, ccloudSSORefreshTokenString}
			credTypeRegex := fmt.Sprintf("(%s)", strings.Join(ccloudCreds, "|"))
			regexString = "^" + fmt.Sprintf(localCredentialStringFormat, credTypeRegex, contextNameRegex)
		}
	} else {
		regexString = "^" + fmt.Sprintf(localCredentialStringFormat, mdsUsernamePasswordString, contextNameRegex)
	}

	return regexp.MustCompile(regexString)
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
	homedir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return homedir + "/_netrc"
	} else {
		return homedir + "/.netrc"
	}
}

func (n *NetrcHandlerImpl) CheckCredentialExist(isCloud bool, ctxName string) (bool, error) {
	netrcFile, err := getNetrc(n.FileName)
	if err != nil {
		return false, err
	}
	machineName1 := GetLocalCredentialName(isCloud, true, ctxName)
	machine1 := netrcFile.FindMachine(machineName1)

	machineName2 := GetLocalCredentialName(isCloud, false, ctxName)
	machine2 := netrcFile.FindMachine(machineName2)

	if machine1 == nil && machine2 == nil {
		return false, nil
	}
	return true, nil
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
