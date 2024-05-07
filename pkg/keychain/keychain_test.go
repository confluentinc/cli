package keychain

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func isIdenticalMachine(expect, actual *Machine) bool {
	return expect.Name == actual.Name &&
		expect.User == actual.User &&
		expect.Password == actual.Password
}

func TestGetMachineNameRegex(t *testing.T) {
	url := "https://confluent.cloud"
	ccloudCtxName := "login-csreesangkom@confleunt.io-https://confluent.cloud"
	confluentCtxName := "login-csreesangkom@confluent.io-http://localhost:8090"
	specialCharsCtxName := `login-csreesangkom+adoooo+\/@-+\^${}[]().*+?|<>-&@confleunt.io-https://confluent.cloud`
	tests := []struct {
		name          string
		params        MachineParams
		matchNames    []string
		nonMatchNames []string
	}{
		{
			name: "ccloud-ctx-name-regex",
			params: MachineParams{
				IsCloud: true,
				Name:    ccloudCtxName,
			},
			matchNames: []string{
				GetLocalCredentialName(true, ccloudCtxName),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(true, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				GetLocalCredentialName(false, ccloudCtxName),
			},
		},
		{
			name: "ccloud-all-regex",
			params: MachineParams{
				IsCloud: true,
				URL:     url,
			},
			matchNames: []string{
				GetLocalCredentialName(true, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(true, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				GetLocalCredentialName(false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "confluent-ctx-name-regex",
			params: MachineParams{
				IsCloud: false,
				Name:    confluentCtxName,
			},
			matchNames: []string{
				GetLocalCredentialName(false, confluentCtxName),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				GetLocalCredentialName(true, confluentCtxName),
			},
		},
		{
			name: "confluent-regex",
			params: MachineParams{
				IsCloud: false,
				URL:     url,
			},
			matchNames: []string{
				GetLocalCredentialName(false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				GetLocalCredentialName(true, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "ccloud-special-chars",
			params: MachineParams{
				IsCloud: true,
				Name:    specialCharsCtxName,
			},
			matchNames: []string{
				GetLocalCredentialName(true, specialCharsCtxName),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(true, ccloudCtxName),
			},
		},
		{
			name: "confluent-special-chars",
			params: MachineParams{
				IsCloud: false,
				Name:    specialCharsCtxName,
			},
			matchNames: []string{
				GetLocalCredentialName(false, specialCharsCtxName),
			},
			nonMatchNames: []string{
				GetLocalCredentialName(false, ccloudCtxName),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			regex := getMachineNameRegex(test.params)
			for _, machineName := range test.nonMatchNames {
				if regex.Match([]byte(machineName)) {
					t.Errorf("Got: regex.Match=true Expect: false\n"+
						"Machine name: %s \n"+
						"Regex String: %s\n"+
						"Params: IsCloud=%t URL=%s", machineName, regex.String(), test.params.IsCloud, test.params.URL)
				}
			}
		})
	}
}

func getMachineNameRegex(params MachineParams) *regexp.Regexp {
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
