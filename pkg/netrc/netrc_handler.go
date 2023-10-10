//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../mock/netrc_handler.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 netrc_handler.go NetrcHandler
package netrc

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	gonetrc "github.com/confluentinc/go-netrc/netrc"
)

const (
	NetrcIntegrationTestFile = "netrc_test"

	localCredentialsPrefix           = "confluent-cli"
	localCredentialStringFormat      = localCredentialsPrefix + ":%s:%s"
	mdsUsernamePasswordString        = "mds-username-password"
	ccloudUsernamePasswordString     = "ccloud-username-password"
	netrcCredentialsNotFoundErrorMsg = `login credentials not found in netrc file "%s"`
	writeToNetrcFileErrorMsg         = `unable to write to netrc file "%s": %w`
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

func (n *NetrcHandlerImpl) RemoveNetrcCredentials(isCloud bool, ctxName string) (string, error) {
	netrcFile, err := getNetrc(n.FileName)
	if err != nil {
		return "", err
	}

	machineName := GetLocalCredentialName(isCloud, ctxName)
	machine := netrcFile.FindMachine(machineName)
	if machine != nil {
		if err := removeCredentials(machineName, netrcFile, n.FileName); err != nil {
			return "", err
		}
		return machine.Login, nil
	} else {
		err = fmt.Errorf(netrcCredentialsNotFoundErrorMsg, n.FileName)
		return "", err
	}
}

func removeCredentials(machineName string, netrcFile *gonetrc.Netrc, filename string) error {
	netrcBytes, err := netrcFile.MarshalText()
	if err != nil {
		return fmt.Errorf(writeToNetrcFileErrorMsg, filename, err)
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
				i++
			}
		}
		if i < length {
			stringBuf = append(stringBuf, lines[i]+"\n")
		}
	}
	if len(stringBuf) > 0 && stringBuf[len(stringBuf)-1] == "\n" { // remove the last newline
		stringBuf = stringBuf[:len(stringBuf)-1]
	}
	joinedString := strings.Join(stringBuf, "")
	joinedString = strings.ReplaceAll(joinedString, "\n\n", "\n")
	buf := []byte(joinedString)
	// get file mode
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	filemode := info.Mode()
	if err := os.WriteFile(filename, buf, filemode); err != nil {
		return fmt.Errorf(writeToNetrcFileErrorMsg, filename, err)
	}
	return nil
}

func getNetrc(filename string) (*gonetrc.Netrc, error) {
	n, err := gonetrc.ParseFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			message := fmt.Sprintf(netrcCredentialsNotFoundErrorMsg, filename)
			return nil, fmt.Errorf("%s: %w", message, err)
		} else {
			return nil, err // failed to parse the netrc file due to other reasons
		}
	}
	return n, nil
}

func GetLocalCredentialName(isCloud bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		credType = ccloudUsernamePassword
	}
	return fmt.Sprintf(localCredentialStringFormat, credType.String(), ctxName)
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
		regexString = "^" + fmt.Sprintf(localCredentialStringFormat, credTypeRegex, contextNameRegex)
	} else {
		regexString = "^" + fmt.Sprintf(localCredentialStringFormat, mdsUsernamePasswordString, contextNameRegex)
	}

	return regexp.MustCompile(regexString)
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

	name := GetLocalCredentialName(isCloud, ctxName)
	return netrcFile.FindMachine(name) != nil, nil
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
