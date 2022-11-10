package netrc

import (
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
	inputFileMds          = "test_files/input-mds"
	inputFileCcloudLogin  = "test_files/input-ccloud-login"
	mdsContext            = "login-mds-user-http://test"
	ccloudLoginContext    = "login-ccloud-login-user@confluent.io-http://test"
	netrcUser             = "jamal@jj"
	netrcPassword         = "12345"
	specialCharsContext   = `login-chris+chris@[]{}.*&$(chris)?\<>|chris/@confluent.io-http://the-special-one`

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
		Name:     "confluent-cli:ccloud-username-password:" + ccloudLoginContext,
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
				Name:    mdsContext,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud login context",
			want: ccloudMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				Name:    ccloudLoginContext,
			},
			file: netrcFilePath,
		},
		{
			name: "No file error",
			params: NetrcMachineParams{
				IsCloud: false,
				Name:    mdsContext,
			},
			wantErr: true,
			file:    "wrong-file",
		},
		{
			name: "Context doesn't exist",
			want: nil,
			params: NetrcMachineParams{
				IsCloud: true,
				Name:    "non-existent-context",
			},
			file: netrcFilePath,
		},
		{
			name: "Context name with special characters",
			want: specialCharsMachine,
			params: NetrcMachineParams{
				IsCloud: true,
				Name:    specialCharsContext,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, _ := os.CreateTemp("", "tempNetrc.json")

			originalNetrc, err := os.ReadFile(tt.inputFile)
			require.NoError(t, err)
			err = os.WriteFile(tempFile.Name(), originalNetrc, 0600)
			require.NoError(t, err)

			netrcHandler := NewNetrcHandler(tempFile.Name())
			err = netrcHandler.WriteNetrcCredentials(tt.isCloud, tt.contextName, netrcUser, netrcPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			gotBytes, err := os.ReadFile(tempFile.Name())
			require.NoError(t, err)
			got := utils.NormalizeNewLines(string(gotBytes))

			wantBytes, err := os.ReadFile(tt.wantFile)
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
				Name:    ccloudCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(true, ccloudCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(false, ccloudCtxName),
			},
		},
		{
			name: "ccloud-all-regex",
			params: NetrcMachineParams{
				IsCloud: true,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName(true, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "confluent-ctx-name-regex",
			params: NetrcMachineParams{
				IsCloud: false,
				Name:    confluentCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(false, confluentCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(true, confluentCtxName),
			},
		},
		{
			name: "confluent-regex",
			params: NetrcMachineParams{
				IsCloud: false,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName(false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName(true, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "ccloud-special-chars",
			params: NetrcMachineParams{
				IsCloud: true,
				Name:    specialCharsCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(true, specialCharsCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(true, ccloudCtxName),
			},
		},
		{
			name: "confluent-special-chars",
			params: NetrcMachineParams{
				IsCloud: false,
				Name:    specialCharsCtxName,
			},
			matchNames: []string{
				getNetrcMachineName(false, specialCharsCtxName),
			},
			nonMatchNames: []string{
				getNetrcMachineName(false, ccloudCtxName),
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
						"Params: IsCloud=%t URL=%s", machineName, regex.String(), tt.params.IsCloud, tt.params.URL)
				}
			}
			for _, machineName := range tt.nonMatchNames {
				if regex.Match([]byte(machineName)) {
					t.Errorf("Got: regex.Match=true Expect: false\n"+
						"Machine name: %s \n"+
						"Regex String: %s\n"+
						"Params: IsCloud=%t URL=%s", machineName, regex.String(), tt.params.IsCloud, tt.params.URL)
				}
			}
		})
	}
}
