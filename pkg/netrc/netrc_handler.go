//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../mock/netrc_handler.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 netrc_handler.go NetrcHandler
package netrc

import (
	"fmt"
	"regexp"
	"strings"
)

const IntegrationTestFile = "netrc_test"

const (
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

type NetrcHandler interface{}

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

func GetLocalCredentialName(isCloud bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		credType = ccloudUsernamePassword
	}
	return fmt.Sprintf(localCredentialStringFormat, credType.String(), ctxName)
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
