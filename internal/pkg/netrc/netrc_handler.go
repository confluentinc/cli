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
	NetrcIntegrationTestFile = "netrc_test"

	netrcCredentialsPrefix       = "confluent-cli"
	netrcCredentialStringFormat  = netrcCredentialsPrefix + ":%s:%s"
	mdsUsernamePasswordString    = "mds-username-password"
	ccloudUsernamePasswordString = "ccloud-username-password"
)

type netrcCredentialType int

const (
	mdsUsernamePassword netrcCredentialType = iota
	ccloudUsernamePassword
)

func (c netrcCredentialType) String() string {
	credTypes := [...]string{mdsUsernamePasswordString, ccloudUsernamePasswordString}
	return credTypes[c]
}

type NetrcHandler interface {
	WriteNetrcCredentials(isCloud bool, ctxName string, username string, password string) error
	RemoveNetrcCredentials(isCloud bool, ctxName string) (string, error)
	CheckCredentialExist(isCloud bool, ctxName string) (bool, error)
	GetMatchingNetrcMachine(params NetrcMachineParams) (*Machine, error)
	GetFileName() string
}

type NetrcMachineParams struct {
	IgnoreCert bool
	IsCloud    bool
	Name       string
	URL        string
}

type Machine struct {
	Name     string
	User     string
	Password string
}

func NewNetrcHandler(netrcFilePath string) *NetrcHandlerImpl {
	return &NetrcHandlerImpl{FileName: netrcFilePath}
}

type NetrcHandlerImpl struct {
	FileName string
}

func (n *NetrcHandlerImpl) WriteNetrcCredentials(isCloud bool, ctxName, username, password string) error {
	netrcFile, err := getOrCreateNetrc(n.FileName)
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, n.FileName)
	}

	machineName := getNetrcMachineName(isCloud, ctxName)

	machine := netrcFile.FindMachine(machineName)
	if machine == nil {
		netrcFile.NewMachine(machineName, username, password, "")
	} else {
		machine.UpdateLogin(username)
		machine.UpdatePassword(password)
	}

	netrcBytes, err := netrcFile.MarshalText()
	if err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, n.FileName)
	}

	if err := os.WriteFile(n.FileName, netrcBytes, 0600); err != nil {
		return errors.Wrapf(err, errors.WriteToNetrcFileErrorMsg, n.FileName)
	}

	return nil
}

func (n *NetrcHandlerImpl) RemoveNetrcCredentials(isCloud bool, ctxName string) (string, error) {
	netrcFile, err := getNetrc(n.FileName)
	if err != nil {
		return "", err
	}

	machineName := getNetrcMachineName(isCloud, ctxName)
	machine := netrcFile.FindMachine(machineName)
	if machine != nil {
		err := removeCredentials(machineName, netrcFile, n.FileName)
		if err != nil {
			return "", err
		}
		return machine.Login, nil
	} else {
		err = errors.New(errors.NetrcCredentialsNotFoundErrorMsg)
		return "", err
	}
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

func getNetrcMachineName(isCloud bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		credType = ccloudUsernamePassword
	}
	return fmt.Sprintf(netrcCredentialStringFormat, credType.String(), ctxName)
}

// Using the parameters to filter and match machine name
// Returns the first match
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
			return &Machine{Name: machine.Name, User: machine.Login, Password: machine.Password}, nil
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
		credTypeRegex := fmt.Sprintf("(%s)", ccloudUsernamePasswordString)
		regexString = "^" + fmt.Sprintf(netrcCredentialStringFormat, credTypeRegex, contextNameRegex)
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

	name := getNetrcMachineName(isCloud, ctxName)
	return netrcFile.FindMachine(name) != nil, nil
}
