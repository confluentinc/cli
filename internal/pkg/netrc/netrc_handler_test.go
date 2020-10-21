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
	mdsContext            = "mds-context"
	ccloudLoginContext    = "ccloud-login"
	ccloudSSOContext      = "ccloud-sso"
	netrcUser             = "jamal@jj"
	netrcPassword         = "12345"
)

func TestGetNetrcCredentialsWithContextName(t *testing.T) {
	tests := []struct {
		name    string
		want    []string
		params  GetMatchingNetrcCredentialsParams
		wantErr bool
		file    string
	}{
		{
			name: "mds context",
			want: []string{netrcUser, netrcPassword},
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "confluent",
				CtxName: mdsContext,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud login context",
			want: []string{netrcUser, netrcPassword},
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "ccloud",
				CtxName: ccloudLoginContext,
			},
			file: netrcFilePath,
		},
		{
			name: "ccloud sso context",
			want: []string{netrcUser, netrcPassword},
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "ccloud",
				CtxName: ccloudSSOContext,
				IsSSO:   true,
			},
			file: netrcFilePath,
		},
		{
			name: "No file error",
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "confluent",
				CtxName: mdsContext,
			},
			wantErr: true,
			file:    "wrong-file",
		},
		{
			name: "Context doesn't exist",
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "ccloud",
				CtxName: "non-existent-context",
			},
			wantErr: true,
			file:    netrcFilePath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netrcHandler := NewNetrcHandler(tt.file)
			var username, password string
			var err error
			if username, password, err = netrcHandler.GetMatchingNetrcCredentials(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("GetNetrcCredentials error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if len(tt.want) != 0 && !t.Failed() && username != tt.want[0] {
				t.Errorf("GetNetrcCredentials username got: %+v, want: %+v", username, tt.want[0])
			}
			if len(tt.want) == 2 && !t.Failed() && password != tt.want[1] {
				t.Errorf("GetNetrcCredentials password got: %+v, want: %+v", password, tt.want[1])
			}
		})
	}
}

func TestNetrcWriter(t *testing.T) {
	tests := []struct {
		name        string
		inputFile   string
		wantFile    string
		cliName     string
		isSSO       bool
		contextName string
		wantErr     bool
	}{
		{
			name:        "add mds context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileMds,
			contextName: mdsContext,
			cliName:     "confluent",
		},
		{
			name:        "add ccloud login context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileCcloudLogin,
			contextName: ccloudLoginContext,
			cliName:     "ccloud",
		},
		{
			name:        "add ccloud sso context credential",
			inputFile:   netrcInput,
			wantFile:    outputFileCcloudSSO,
			contextName: ccloudSSOContext,
			cliName:     "ccloud",
			isSSO:       true,
		},
		{
			name:        "update mds context credential",
			inputFile:   inputFileMds,
			wantFile:    outputFileMds,
			contextName: mdsContext,
			cliName:     "confluent",
		},
		{
			name:        "update ccloud login context credential",
			inputFile:   inputFileCcloudLogin,
			wantFile:    outputFileCcloudLogin,
			contextName: ccloudLoginContext,
			cliName:     "ccloud",
		},
		{
			name:        "update ccloud sso context credential",
			inputFile:   inputFileCcloudSSO,
			wantFile:    outputFileCcloudSSO,
			contextName: ccloudSSOContext,
			cliName:     "ccloud",
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
			err = netrcHandler.WriteNetrcCredentials(tt.cliName, tt.isSSO, tt.contextName, netrcUser, netrcPassword)
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
	tests := []struct {
		name          string
		params        GetMatchingNetrcCredentialsParams
		matchNames    []string
		nonMatchNames []string
	}{
		{
			name: "ccloud-sso-regex",
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "ccloud",
				IsSSO:   true,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName("ccloud", true, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName("ccloud", false, "login-csreesangkom@confleunt.io-"+url),
				getNetrcMachineName("ccloud", false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName("confluent", false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "ccloud-all-regex",
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "ccloud",
				IsSSO:   false,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName("ccloud", true, "login-csreesangkom@confleunt.io-"+url),
				getNetrcMachineName("ccloud", false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName("ccloud", false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName("confluent", false, "login-csreesangkom@confleunt.io-"+url),
			},
		},
		{
			name: "confluent-regex",
			params: GetMatchingNetrcCredentialsParams{
				CLIName: "confluent",
				IsSSO:   false,
				URL:     url,
			},
			matchNames: []string{
				getNetrcMachineName("confluent", false, "login-csreesangkom@confleunt.io-"+url),
			},
			nonMatchNames: []string{
				getNetrcMachineName("confluent", false, "login-csreesangkom@confleunt.io-"+"https://wassup"),
				getNetrcMachineName("ccloud", false, "login-csreesangkom@confleunt.io-"+url),
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
						"Params: CLIName=%s IsSSO=%t URL=%s", machineName, regex.String(), tt.params.CLIName, tt.params.IsSSO, tt.params.URL)
				}
			}
			for _, machineName := range tt.nonMatchNames {
				if regex.Match([]byte(machineName)) {
					t.Errorf("Got: regex.Match=true Expect: false\n"+
						"Machine name: %s \n"+
						"Regex String: %s\n"+
						"Params: CLIName=%s IsSSO=%t URL=%s", machineName, regex.String(), tt.params.CLIName, tt.params.IsSSO, tt.params.URL)
				}
			}
		})
	}
}
