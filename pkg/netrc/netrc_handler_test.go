package netrc

import (
	"testing"
)

const (
	netrcFilePath          = "test_files/netrc"
	mdsContext             = "login-mds-user-http://test"
	ccloudNetrcMachineName = "login-ccloud-login-user@confluent.io-http://test"
	specialCharsContext    = `login-chris+chris@[]{}.*&$(chris)?\<>|chris/@confluent.io-http://the-special-one`

	loginURL          = "http://test"
	ccloudLogin       = "ccloud-login-user@confluent.io"
	ccloudDiffLogin   = "ccloud-login-user-diff-url@confluent.io"
	ccloudDiffURL     = "http://differenturl"
	mdsLogin          = "mds-user"
	mockPassword      = "mock-password"
	specialCharsLogin = `chris+chris@[]{}.*&$(chris)?\<>|chris/@confluent.io`
)

var (
	ccloudMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:" + ccloudNetrcMachineName,
		User:     ccloudLogin,
		Password: mockPassword,
	}

	ccloudDiffURLMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:login-" + ccloudDiffLogin + "-" + ccloudDiffURL,
		User:     ccloudDiffLogin,
		Password: mockPassword,
	}
	confluentMachine = &Machine{
		Name:     "confluent-cli:mds-username-password:" + mdsContext,
		User:     mdsLogin,
		Password: mockPassword,
	}
	specialCharsMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:" + specialCharsContext,
		User:     specialCharsLogin,
		Password: mockPassword,
	}
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
		params        NetrcMachineParams
		matchNames    []string
		nonMatchNames []string
	}{
		{
			name: "ccloud-ctx-name-regex",
			params: NetrcMachineParams{
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
			params: NetrcMachineParams{
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
			params: NetrcMachineParams{
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
			params: NetrcMachineParams{
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
			params: NetrcMachineParams{
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
			params: NetrcMachineParams{
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
