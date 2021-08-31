package netrc

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	netrcFilePath         = "test_files/netrc"
	netrcInput            = "test_files/netrc-input"
	outputFileMds         = "test_files/output-mds"
	outputFileCcloudLogin = "test_files/output-ccloud-login"
	outputFileCcloudSSO   = "test_files/output-ccloud-sso"
	inputFileMds          = "test_files/input-mds"
	inputFileCcloudLogin  = "test_files/input-ccloud-login"
	inputFileCcloudSSO    = "test_files/input-ccloud-sso"
	mdsContext            = "login-mds-user-http://test"
	ccloudLoginContext    = "login-ccloud-login-user@confluent.io-http://test"
	ccloudSSOContext      = "login-ccloud-sso-user@confluent.io-http://test"
	netrcUser             = "jamal@jj"
	netrcPassword         = "12345"
	specialCharsContext   = `login-chris+chris@[]{}.*&$(chris)?\<>|chris/@confluent.io-http://the-special-one`

	loginURL          = "http://test"
	ssoFirstURL       = "http://ssofirst"
	ccloudLogin       = "ccloud-login-user@confluent.io"
	ccloudDiffLogin   = "ccloud-login-user-diff-url@confluent.io"
	ccloudDiffURL     = "http://differenturl"
	ccloudSSOLogin    = "ccloud-sso-user@confluent.io"
	mdsLogin          = "mds-user"
	ssoFirstLogin     = "sso-first@confluent.io"
	mockPassword      = "mock-password"
	refreshToken      = "refresh-token"
	specialCharsLogin = `chris+chris@[]{}.*&$(chris)?\<>|chris/@confluent.io`
)

var (
	ccloudMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:" + ccloudLoginContext,
		User:     ccloudLogin,
		Password: mockPassword,
		IsSSO:    false,
	}

	ccloudDiffURLMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:login-" + ccloudDiffLogin + "-" + ccloudDiffURL,
		User:     ccloudDiffLogin,
		Password: mockPassword,
		IsSSO:    false,
	}
	ccloudSSOMachine = &Machine{
		Name:     "confluent-cli:ccloud-sso-refresh-token:" + ccloudSSOContext,
		User:     ccloudSSOLogin,
		Password: refreshToken,
		IsSSO:    true,
	}
	confluentMachine = &Machine{
		Name:     "confluent-cli:mds-username-password:" + mdsContext,
		User:     mdsLogin,
		Password: mockPassword,
		IsSSO:    false,
	}
	ssoFirstMachine = &Machine{
		Name:     "confluent-cli:ccloud-sso-refresh-token:login-sso-first@confluent.io-" + ssoFirstURL,
		User:     ssoFirstLogin,
		Password: refreshToken,
		IsSSO:    true,
	}
	specialCharsMachine = &Machine{
		Name:     "confluent-cli:ccloud-username-password:" + specialCharsContext,
		User:     specialCharsLogin,
		Password: mockPassword,
		IsSSO:    false,
	}
)

func TestGetMatchingNetrcMachineWithContextName(t *testing.T) {
	tests := []struct {
		name    string
		want    *Machine
		params  NetrcMachineParams
		wantErr bool
		file    string
	}{
		{
			name: "mds context",
			want: confluentMachine,
			params: NetrcMachineParams{
				IsCloud: false,
				CtxName: mdsContext,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud login context",
			want: ccloudMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				CtxName: ccloudLoginContext,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud sso context",
			want: ccloudSSOMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				CtxName: ccloudSSOContext,
				IsSSO:   true,
			},
			file: netrcFilePath,
		},
		{
			name: "No file error",
			params: NetrcMachineParams{
				IsCloud: false,
				CtxName: mdsContext,
			},
			wantErr: true,
			file:    "wrong-file",
		},
		{
			name: "Context doesn't exist",
			want: nil,
			params: NetrcMachineParams{
				IsCloud: true,
				CtxName: "non-existent-context",
			},
			file: netrcFilePath,
		},
		{
			name: "Context name with special characters",
			want: specialCharsMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				CtxName: specialCharsContext,
			},
			file: netrcFilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := NewNetrcHandler(tt.file)
			var machine *Machine
			var err error
			if machine, err = netrcHandler.GetMatchingNetrcMachine(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingNetrcMachine error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if !t.Failed() {
				if tt.want == nil {
					if machine != nil {
						t.Error("GetMatchingNetrcMachine expect nil machine but got non nil machine")
					}
				} else {
					if machine == nil {
						t.Errorf("Expected to find want : %+v but found no machines", machine)
					}
					if !isIdenticalMachine(tt.want, machine) {
						t.Errorf("GetMatchingNetrcMachine mismatch\ngot: %+v \nwant: %+v", machine, tt.want)
					}
				}

			}
		})
	}
}

func isIdenticalMachine(expect, actual *Machine) bool {
	return expect.Name == actual.Name &&
		expect.IsSSO == actual.IsSSO &&
		expect.User == actual.User &&
		expect.Password == actual.Password
}

func TestGetMatchingNetrcMachineFromURL(t *testing.T) {
	tests := []struct {
		name    string
		want    *Machine
		params  NetrcMachineParams
		wantErr bool
		file    string
	}{
		{
			name: "ccloud login with url",
			want: ccloudMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				URL:     loginURL,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud login no url",
			want: ccloudDiffURLMachine,
			params: NetrcMachineParams{
				IsCloud: true,
			},
			file: netrcFilePath,
		},
		{
			name: "confluent login with url",
			want: confluentMachine,
			params: NetrcMachineParams{
				IsCloud: false,
				URL:     loginURL,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud sso with url",
			want: ccloudSSOMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				IsSSO:   true,
				URL:     loginURL,
			},
			file: netrcFilePath,
		},
		{
			name: "no sso specified but sso comes first",
			want: ssoFirstMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				URL:     ssoFirstURL,
			},
			file: netrcFilePath,
		},
		{
			name: "No file error",
			params: NetrcMachineParams{
				IsCloud: false,
			},
			wantErr: true,
			file:    "wrong-file",
		},
		{
			name: "URL doesn't exist",
			want: nil,
			params: NetrcMachineParams{
				IsCloud: true,
				URL:     "http://dontexist",
			},
			file: netrcFilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := NewNetrcHandler(tt.file)
			var machine *Machine
			var err error
			if machine, err = netrcHandler.GetMatchingNetrcMachine(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingNetrcMachine error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if !t.Failed() {
				if tt.want == nil {
					if machine != nil {
						t.Error("GetMatchingNetrcMachine expect nil machine but got non nil machine")
					}
				} else {
					if machine == nil {
						t.Errorf("Expected to find want : %+v but found no machines", machine)
					}
					if !isIdenticalMachine(tt.want, machine) {
						t.Errorf("GetMatchingNetrcMachine mismatch \ngot: %+v \nwant: %+v", machine, tt.want)
					}
				}

			}
		})
	}
}

func TestNetrcWriter(t *testing.T) {
	tests := []struct {
		name        string
		inputFile   string
		wantFile    string
		isCloud     bool
		isSSO       bool
		contextName string
		wantErr     bool
	}{
		{
			name:        "add mds context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileMds,
			contextName: mdsContext,
			isCloud:     false,
		},
		{
			name:        "add ccloud login context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileCcloudLogin,
			contextName: ccloudLoginContext,
			isCloud:     true,
		},
		{
			name:        "add ccloud sso context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileCcloudSSO,
			contextName: ccloudSSOContext,
			isCloud:     true,
			isSSO:       true,
		},
		{
			name:        "update mds context credential",
			inputFile:   inputFileMds,
			wantFile:    outputFileMds,
			contextName: mdsContext,
			isCloud:     false,
		},
		{
			name:        "update ccloud login context credential",
			inputFile:   inputFileCcloudLogin,
			wantFile:    outputFileCcloudLogin,
			contextName: ccloudLoginContext,
			isCloud:     true,
		},
		{
			name:        "update ccloud sso context credential",
			inputFile:   inputFileCcloudSSO,
			wantFile:    outputFileCcloudSSO,
			contextName: ccloudSSOContext,
			isCloud:     true,
			isSSO:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, _ := ioutil.TempFile("", "tempNetrc.json")

			originalNetrc, err := ioutil.ReadFile(tt.inputFile)
			require.NoError(t, err)
			err = ioutil.WriteFile(tempFile.Name(), originalNetrc, 0600)
			require.NoError(t, err)

			netrcHandler := NewNetrcHandler(tempFile.Name())
			err = netrcHandler.WriteNetrcCredentials(tt.isCloud, tt.isSSO, tt.contextName, netrcUser, netrcPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			gotBytes, err := ioutil.ReadFile(tempFile.Name())
			require.NoError(t, err)
			got := utils.NormalizeNewLines(string(gotBytes))

			wantBytes, err := ioutil.ReadFile(tt.wantFile)
			require.NoError(t, err)
			want := utils.NormalizeNewLines(string(wantBytes))

			if got != want {
				t.Errorf("got: \n%s\nwant: \n%s\n", got, want)
			}
			_ = os.Remove(tempFile.Name())
		})
	}
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
				CtxName: ccloudCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(true, true, ccloudCtxName),
				getNetrcMachineName(true, false, ccloudCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(true, true, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(false, false, ccloudCtxName),
			},
		},
		{
			name: "ccloud-sso-regex",
			params: NetrcMachineParams{
				IsCloud: true,
				IsSSO:   true,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName(true, true, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+url),
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(false, false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "ccloud-all-regex",
			params: NetrcMachineParams{
				IsCloud: true,
				IsSSO:   false,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName(true, true, "login-csreesangkom@confleunt.io-"+url),
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(false, false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "confluent-ctx-name-regex",
			params: NetrcMachineParams{
				IsCloud: false,
				CtxName: confluentCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(false, false, confluentCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(true, false, confluentCtxName),
			},
		},
		{
			name: "confluent-regex",
			params: NetrcMachineParams{
				IsCloud: false,
				IsSSO:   false,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName(false, false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(true, false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "ccloud-special-chars",
			params: NetrcMachineParams{
				IsCloud: true,
				CtxName: specialCharsCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(true, false, specialCharsCtxName),
				getNetrcMachineName(true, true, specialCharsCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, false, ccloudCtxName),
				getNetrcMachineName(true, true, ccloudCtxName),
			},
		},
		{
			name: "confluent-special-chars",
			params: NetrcMachineParams{
				IsCloud: false,
				CtxName: specialCharsCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(false, false, specialCharsCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, false, ccloudCtxName),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := getMachineNameRegex(tt.params)
			for _, machineName := range tt.matchNames {
				if !regex.Match([]byte(machineName)) {
					t.Errorf("Got: regex.Match=false Expect: true\n"+
						"Machine name: %s \n"+
						"Regex String: %s \n"+
						"Params: IsCloud=%t IsSSO=%t URL=%s", machineName, regex.String(), tt.params.IsCloud, tt.params.IsSSO, tt.params.URL)
				}
			}
			for _, machineName := range tt.nonMatchNames {
				if regex.Match([]byte(machineName)) {
					t.Errorf("Got: regex.Match=true Expect: false\n"+
						"Machine name: %s \n"+
						"Regex String: %s\n"+
						"Params: IsCloud=%t IsSSO=%t URL=%s", machineName, regex.String(), tt.params.IsCloud, tt.params.IsSSO, tt.params.URL)
				}
			}
		})
	}
}
