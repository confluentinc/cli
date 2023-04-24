package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	authTokenPlaceholder        = "AUTH_TOKEN_PLACEHOLDER"
	authRefreshTokenPlaceholder = "AUTH_REFRESH_TOKEN_PLACEHOLDER"
	saltPlaceholder             = "SALT_PLACEHOLDER"
	noncePlaceholder            = "NONCE_PLACEHOLDER"
)

var (
	apiKeyString    = "abc-key-123"
	apiSecretString = "def-secret-456"
	kafkaClusterID  = "anonymous-id"
	contextName     = "my-context"
	accountID       = "acc-123"
	cloudPlatforms  = []string{
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
		"confluent.cloud",
		"premium-oryx.gcp.priv.cpdev.cloud",
	}
	account = &orgv1.Account{
		Id:   accountID,
		Name: "test-env",
	}
	regularOrgContextState = &ContextState{
		Auth: &AuthConfig{
			User: &orgv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Account: account,
			Accounts: []*orgv1.Account{
				account,
			},
			Organization: testserver.RegularOrg,
		},
		AuthToken:        "eyJ.eyJ.abc",
		AuthRefreshToken: "v1.abc",
	}
	suspendedOrgContextState = func(eventType orgv1.SuspensionEventType) *ContextState {
		return &ContextState{
			Auth: &AuthConfig{
				Organization: testserver.SuspendedOrg(eventType),
			},
		}
	}
)

type TestInputs struct {
	kafkaClusters        map[string]*KafkaClusterConfig
	activeKafka          string
	statefulConfig       *Config
	statelessConfig      *Config
	twoEnvStatefulConfig *Config
	account              *orgv1.Account
}

func SetupTestInputs(isCloud bool) *TestInputs {
	testInputs := &TestInputs{}
	platform := &Platform{
		Name:   "http://test",
		Server: "http://test",
	}
	if isCloud {
		platform.Name = testserver.TestCloudURL.String()
	}
	apiCredential := &Credential{
		Name:     "api-key-abc-key-123",
		Username: "",
		Password: "",
		APIKeyPair: &APIKeyPair{
			Key:    apiKeyString,
			Secret: apiSecretString,
		},
		CredentialType: 1,
	}
	loginCredential := &Credential{
		Name:           "username-test-user",
		Username:       "test-user",
		Password:       "",
		APIKeyPair:     nil,
		CredentialType: 0,
	}
	account := &orgv1.Account{
		Id:   accountID,
		Name: "test-env",
	savedCredentials := map[string]*LoginCredential{
		contextName: &LoginCredential{
			IsCloud:           isCloud,
			Username:          "test-user",
			EncryptedPassword: "encrypted-password",
		},
	}
	account2 := &orgv1.Account{
		Id:   "env-flag",
		Name: "test-env2",
	}
	testInputs.account = account
	state := &ContextState{
		Auth: &AuthConfig{
			User: &orgv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Account: account,
			Accounts: []*orgv1.Account{
				account,
			},
			Organization: &orgv1.Organization{
				Id:   321,
				Name: "test-org",
			},
		},
		AuthToken: "abc123",
	}
	twoEnvState := &ContextState{
		Auth: &AuthConfig{
			User: &orgv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Account: account,
			Accounts: []*orgv1.Account{
				account,
				account2,
			},
			Organization: &orgv1.Organization{
				Id:   321,
				Name: "test-org",
			},
		},
		AuthToken: "abc123",
	}
	testInputs.kafkaClusters = map[string]*KafkaClusterConfig{
		kafkaClusterID: {
			ID:          kafkaClusterID,
			Name:        "anonymous-cluster",
			Bootstrap:   "http://test",
			APIEndpoint: "",
			APIKeys: map[string]*APIKeyPair{
				apiKeyString: {
					Key:    apiKeyString,
					Secret: apiSecretString,
				},
			},
			APIKey: apiKeyString,
		},
	}
	testInputs.activeKafka = kafkaClusterID
	statefulContext := &Context{
		Name:           contextName,
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			accountID: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State: state,
	}
	statelessContext := &Context{
		Name:                   contextName,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             apiCredential,
		CredentialName:         apiCredential.Name,
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{},
		State:                  &ContextState{},
		Config:                 &Config{SavedCredentials: savedCredentials},
	}
	twoEnvStatefulContext := &Context{
		Name:           contextName,
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			accountID: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State: twoEnvState,
	}
	context := "onprem"
	if isCloud {
		context = "cloud"
	}
	testInputs.statefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params:   &config.Params{},
			Filename: fmt.Sprintf("test_json/stateful_%s.json", context),
			Ver:      config.Version{Version: Version},
		},
		Platforms: map[string]*Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: statefulContext,
		},
		ContextStates: map[string]*ContextState{
			contextName: state,
		},
		CurrentContext:   contextName,
		IsTest:           true,
		SavedCredentials: savedCredentials,
	}
	testInputs.statelessConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params:   &config.Params{},
			Filename: fmt.Sprintf("test_json/stateless_%s.json", context),
			Ver:      config.Version{Version: Version},
		},
		Platforms: map[string]*Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: statelessContext,
		},
		ContextStates: map[string]*ContextState{
			contextName: {},
		},
		CurrentContext:   contextName,
		IsTest:           true,
		SavedCredentials: savedCredentials,
	}
	testInputs.twoEnvStatefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params:   &config.Params{},
			Filename: fmt.Sprintf("test_json/stateful_%s.json", context),
			Ver:      config.Version{Version: Version},
		},
		Platforms: map[string]*Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: twoEnvStatefulContext,
		},
		ContextStates: map[string]*ContextState{
			contextName: twoEnvState,
		},
		CurrentContext: contextName,
		IsTest:         true,
	}

	statefulContext.Config = testInputs.statefulConfig
	statefulContext.KafkaClusterContext = NewKafkaClusterContext(statefulContext, testInputs.activeKafka, testInputs.kafkaClusters)

	statelessContext.Config = testInputs.statelessConfig
	statelessContext.KafkaClusterContext = NewKafkaClusterContext(statelessContext, testInputs.activeKafka, testInputs.kafkaClusters)

	twoEnvStatefulContext.Config = testInputs.twoEnvStatefulConfig
	twoEnvStatefulContext.KafkaClusterContext = NewKafkaClusterContext(twoEnvStatefulContext, testInputs.activeKafka, testInputs.kafkaClusters)

	return testInputs
}

func TestConfig_Load(t *testing.T) {
	testConfigsOnPrem := SetupTestInputs(false)
	testConfigsCloud := SetupTestInputs(true)
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
		file    string
	}{
		{
			name: "succeed loading stateless on-prem config from file",
			want: testConfigsOnPrem.statelessConfig,
			file: "test_json/stateless_onprem.json",
		},
		{
			name: "succeed loading on-prem config with state from file",
			want: testConfigsOnPrem.statefulConfig,
			file: "test_json/stateful_onprem.json",
		},
		{
			name: "succeed loading stateless cloud config from file",
			want: testConfigsCloud.statelessConfig,
			file: "test_json/stateless_cloud.json",
		},
		{
			name: "succeed loading cloud config with state from file",
			want: testConfigsCloud.statefulConfig,
			file: "test_json/stateful_cloud.json",
		},
		{
			name: "should load disable update checks and disable updates",
			want: &Config{
				BaseConfig: &config.BaseConfig{
					Params:   &config.Params{},
					Filename: "test_json/load_disable_update.json",
					Ver:      config.Version{Version: Version},
				},
				DisableUpdates:     true,
				DisableUpdateCheck: true,
				Platforms:          map[string]*Platform{},
				Credentials:        map[string]*Credential{},
				Contexts:           map[string]*Context{},
				ContextStates:      map[string]*ContextState{},
				SavedCredentials:   map[string]*LoginCredential{},
			},
			file: "test_json/load_disable_update.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Params{})
			c.Filename = tt.file
			for _, context := range tt.want.Contexts {
				context.Config = tt.want
			}
			if err := c.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %+v, wantErr %+v", err, tt.wantErr)
			}

			// Get around automatically assigned anonymous id and IsTest check
			tt.want.AnonymousId = c.AnonymousId
			tt.want.IsTest = c.IsTest

			if !t.Failed() && !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %+v, want %+v", c, tt.want)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	testConfigsOnPrem := SetupTestInputs(false)
	testConfigsCloud := SetupTestInputs(true)
	tests := []struct {
		name             string
		isCloud          bool
		config           *Config
		wantFile         string
		wantErr          bool
		kafkaOverwrite   string
		contextOverwrite string
		accountOverwrite *orgv1.Account
	}{
		{
			name:     "save on-prem config with state to file",
			isCloud:  false,
			config:   testConfigsOnPrem.statefulConfig,
			wantFile: "test_json/stateful_onprem_save.json",
		},
		{
			name:     "save stateless on-prem config to file",
			isCloud:  false,
			config:   testConfigsOnPrem.statelessConfig,
			wantFile: "test_json/stateless_onprem.json",
		},
		{
			name:     "save cloud config with state to file",
			isCloud:  true,
			config:   testConfigsCloud.statefulConfig,
			wantFile: "test_json/stateful_cloud_save.json",
		},
		{
			name:     "save stateless cloud config to file",
			isCloud:  true,
			config:   testConfigsCloud.statelessConfig,
			wantFile: "test_json/stateless_cloud.json",
		},
		{
			name:           "save stateless cloud config with kafka overwrite to file",
			isCloud:        true,
			config:         testConfigsCloud.statefulConfig,
			wantFile:       "test_json/stateful_cloud_save.json",
			kafkaOverwrite: "lkc-clusterFlag",
		},
		{
			name:           "save stateless cloud config with kafka and context overwrite to file",
			isCloud:        true,
			config:         testConfigsCloud.statefulConfig,
			wantFile:       "test_json/stateful_cloud_save.json",
			kafkaOverwrite: "lkc-clusterFlag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
			tt.config.Filename = configFile.Name()
			ctx := tt.config.Context()
			tt.config.SavedCredentials = map[string]*LoginCredential{
				contextName: &LoginCredential{
					IsCloud:           tt.isCloud,
					Username:          "test-user",
					EncryptedPassword: "encrypted-password",
				},
			}
			if tt.kafkaOverwrite != "" {
				tt.config.SetOverwrittenActiveKafka(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
				ctx.KafkaClusterContext.SetActiveKafkaCluster(tt.kafkaOverwrite)
			}
			if tt.contextOverwrite != "" {
				tt.config.SetOverwrittenCurrContext(tt.config.CurrentContext)
				tt.config.CurrentContext = tt.contextOverwrite
			}
			if err := tt.config.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, _ := os.ReadFile(configFile.Name())
			want, _ := os.ReadFile(tt.wantFile)
			wantString := replacePlaceholdersInWant(t, got, want)
			require.Equal(t, utils.NormalizeNewLines(wantString), utils.NormalizeNewLines(string(got)))
			fd, err := os.Stat(configFile.Name())
			require.NoError(t, err)
			if runtime.GOOS != "windows" && fd.Mode() != 0600 {
				t.Errorf("Config.Save() file should only be readable by user")
			}
			os.Remove(configFile.Name())
		})
	}
}

func TestConfig_SaveWithAccountOverwrite(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)
	tests := []struct {
		name             string
		config           *Config
		wantFile         string
		wantErr          bool
		accountOverwrite *orgv1.Account
	}{
		{
			name:             "save cloud config with state and account overwrite to file",
			config:           testConfigsCloud.twoEnvStatefulConfig,
			wantFile:         "test_json/account_overwrite.json",
			accountOverwrite: &orgv1.Account{Id: "env-flag"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
			tt.config.Filename = configFile.Name()
			tt.config.SavedCredentials = map[string]*LoginCredential{
				contextName: &LoginCredential{
					IsCloud:           true,
					Username:          "test-user",
					EncryptedPassword: "encrypted-password",
				},
			}
			if tt.accountOverwrite != nil {
				tt.config.SetOverwrittenAccount(tt.config.Context().State.Auth.Account)
				tt.config.Context().State.Auth.Account = tt.accountOverwrite
			}
			if err := tt.config.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ioutil.ReadFile(configFile.Name())
			got = append(got, '\n') //account for extra newline at the end of the json file
			want, _ := os.ReadFile(tt.wantFile)
			wantString := replacePlaceholdersInWant(t, got, want)
			if utils.NormalizeNewLines(string(got)) != utils.NormalizeNewLines(wantString) {
				t.Errorf("Config.Save() = %v\n want = %v", utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(wantString))
			}
			fd, err := os.Stat(configFile.Name())
			require.NoError(t, err)
			if runtime.GOOS != "windows" && fd.Mode() != 0600 {
				t.Errorf("Config.Save() file should only be readable by user")
			}
			os.Remove(configFile.Name())
		})
	}
}

func replacePlaceholdersInWant(t *testing.T, got []byte, want []byte) string {
	data := Config{}
	err := json.Unmarshal(got, &data)
	require.NoError(t, err)
	wantString := strings.Replace(string(want), authTokenPlaceholder, data.ContextStates[contextName].AuthToken, -1)
	wantString = strings.Replace(wantString, authRefreshTokenPlaceholder, data.ContextStates[contextName].AuthRefreshToken, -1)
	saltString := base64.RawStdEncoding.EncodeToString(data.ContextStates[contextName].Salt)
	wantString = strings.Replace(wantString, saltPlaceholder, saltString, -1)
	nonceString := base64.RawStdEncoding.EncodeToString(data.ContextStates[contextName].Nonce)
	return strings.Replace(wantString, noncePlaceholder, nonceString, -1)
}

func TestConfig_OverwrittenKafka(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string //simulates initial config value overwritten by a cluster flag value
		activeKafka    string //simulates the cluster flag value
	}{
		{
			name:        "test no overwrite value",
			config:      testConfigsCloud.statefulConfig,
			activeKafka: testConfigsCloud.activeKafka,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCloud.statefulConfig,
			overwrittenVal: "lkc-test",
			activeKafka:    testConfigsCloud.activeKafka,
		},
		{
			name:        "test no overwrite value",
			config:      testConfigsCloud.statelessConfig,
			activeKafka: testConfigsCloud.activeKafka,
		},
	}
	for _, tt := range tests {
		ctx := tt.config.Context()
		tt.config.SetOverwrittenActiveKafka(tt.overwrittenVal)
		//resolve should reset the active kafka to be the overwritten value and return the flag value to be used in restore
		tempKafka := tt.config.resolveOverwrittenKafka()
		require.Equal(t, tt.activeKafka, tempKafka)
		if ctx.KafkaClusterContext.EnvContext && ctx.KafkaClusterContext.GetCurrentKafkaEnvContext() != nil {
			require.Equal(t, tt.overwrittenVal, ctx.KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster)
		} else {
			require.Equal(t, tt.overwrittenVal, ctx.KafkaClusterContext.ActiveKafkaCluster)
		}
		//restore should reset the active kafka to be the flag value
		tt.config.restoreOverwrittenKafka(tempKafka)
		if ctx.KafkaClusterContext.EnvContext && ctx.KafkaClusterContext.GetCurrentKafkaEnvContext() != nil {
			require.Equal(t, tempKafka, ctx.KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster)
		} else {
			require.Equal(t, tempKafka, ctx.KafkaClusterContext.ActiveKafkaCluster)
		}
		tt.config.overwrittenActiveKafka = ""
	}
}

func TestConfig_OverwrittenContext(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string //simulates initial context value overwritten by a context flag value
		currContext    string //simulates the context flag value
	}{
		{
			name:        "test no overwrite value",
			config:      testConfigsCloud.statefulConfig,
			currContext: testConfigsCloud.statefulConfig.CurrentContext,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCloud.statefulConfig,
			overwrittenVal: "test-context",
			currContext:    testConfigsCloud.statefulConfig.CurrentContext,
		},
		{
			name:        "test no overwrite value",
			config:      testConfigsCloud.statelessConfig,
			currContext: testConfigsCloud.statelessConfig.CurrentContext,
		},
	}
	for _, tt := range tests {
		tt.config.SetOverwrittenCurrContext(tt.overwrittenVal)
		//resolve should reset the current context to be the overwritten value and return the flag value to be used in restore
		tempContext := tt.config.resolveOverwrittenContext()
		require.Equal(t, tt.overwrittenVal, tt.config.CurrentContext)
		require.Equal(t, tt.currContext, tempContext)
		//restore should reset the current context to be the flag value
		tt.config.restoreOverwrittenContext(tempContext)
		require.Equal(t, tt.currContext, tt.config.CurrentContext)
		tt.config.overwrittenCurrContext = ""
	}
}

func TestConfig_OverwrittenAccount(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal *orgv1.Account //simulates initial environment (account) value overwritten by a environment flag
		activeAccount  string         //simulates the environment (account) flag value
	}{
		{
			name:          "test no overwrite value",
			config:        testConfigsCloud.statefulConfig,
			activeAccount: testConfigsCloud.statefulConfig.Context().State.Auth.Account.Id,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCloud.statefulConfig,
			overwrittenVal: &orgv1.Account{Id: "env-test"},
			activeAccount:  testConfigsCloud.statefulConfig.Context().State.Auth.Account.Id,
		},
		{
			name:   "test no overwrite value",
			config: testConfigsCloud.statelessConfig,
		},
	}
	for _, tt := range tests {
		tt.config.SetOverwrittenAccount(tt.overwrittenVal)
		if tt.config.Context().State.Auth == nil {
			tempAccount := tt.config.resolveOverwrittenAccount()
			require.Nil(t, tempAccount)
			tt.config.restoreOverwrittenAccount(tempAccount)
			require.Nil(t, tt.config.Context().State.Auth)
		} else {
			//resolve should reset the current context to be the overwritten value and return the flag value to be used in restore
			tempAccount := tt.config.resolveOverwrittenAccount()
			if tt.overwrittenVal != nil {
				require.Equal(t, tt.overwrittenVal, tt.config.Context().State.Auth.Account)
				require.Equal(t, tt.activeAccount, tempAccount.Id)
			}
			//restore should reset the current context to be the flag value
			tt.config.restoreOverwrittenAccount(tempAccount)
			require.Equal(t, tt.activeAccount, tt.config.Context().State.Auth.Account.Id)
		}
		tt.config.overwrittenAccount = nil
	}
}

func TestConfig_getFilename(t *testing.T) {
	c := New(&config.Params{})
	got := c.GetFilename()
	want := filepath.FromSlash(os.Getenv("HOME") + "/.confluent/config.json")
	if got != want {
		t.Errorf("Config.GetFilename() = %v, want %v", got, want)
	}
}

func TestConfig_AddContext(t *testing.T) {
	filename := "/tmp/TestConfig_AddContext.json"
	conf := AuthenticatedOnPremConfigMock()
	conf.Filename = filename
	context := conf.Context()
	conf.CurrentContext = ""
	noContextConf := AuthenticatedOnPremConfigMock()
	noContextConf.Filename = filename
	delete(noContextConf.Contexts, noContextConf.Context().Name)
	noContextConf.CurrentContext = ""

	type testStruct struct {
		name                   string
		config                 *Config
		contextName            string
		platformName           string
		credentialName         string
		kafkaClusters          map[string]*KafkaClusterConfig
		kafka                  string
		schemaRegistryClusters map[string]*SchemaRegistryCluster
		state                  *ContextState
		Version                *pversion.Version
		filename               string
		want                   *Config
		wantErr                bool
	}

	test := testStruct{
		name:                   "",
		config:                 noContextConf,
		contextName:            context.Name,
		platformName:           context.PlatformName,
		credentialName:         context.CredentialName,
		kafkaClusters:          context.KafkaClusterContext.KafkaClusterConfigs,
		kafka:                  context.KafkaClusterContext.ActiveKafkaCluster,
		schemaRegistryClusters: context.SchemaRegistryClusters,
		state:                  context.State,
		filename:               filename,
		want:                   nil,
		wantErr:                false,
	}

	addValidContextTest := test
	addValidContextTest.name = "add valid context"
	addValidContextTest.want = conf
	addValidContextTest.wantErr = false

	failAddingExistingContextTest := test
	failAddingExistingContextTest.name = "add valid context"
	failAddingExistingContextTest.want = nil
	failAddingExistingContextTest.wantErr = true

	tests := []testStruct{
		addValidContextTest,
		failAddingExistingContextTest,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.AddContext(tt.contextName, tt.platformName, tt.credentialName, tt.kafkaClusters, tt.kafka,
				tt.schemaRegistryClusters, tt.state, MockOrgResourceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				tt.want.AnonymousId = tt.config.AnonymousId
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.want, tt.config) {
				t.Errorf("AddContext() got = %v, want %v", tt.config, tt.want)
			}
		})
	}
	os.Remove(filename)
}

func TestConfig_CreateContext(t *testing.T) {
	cfg := &Config{
		BaseConfig:    &config.BaseConfig{Params: new(config.Params), Ver: config.Version{Version: new(version.Version)}},
		ContextStates: make(map[string]*ContextState),
		Contexts:      make(map[string]*Context),
		Credentials:   make(map[string]*Credential),
		Platforms:     make(map[string]*Platform),
	}

	err := cfg.CreateContext("context", "https://example.com", "api-key", "api-secret")
	require.NoError(t, err)

	ctx := cfg.Contexts["context"]
	require.Equal(t, "context", ctx.Name)
	require.Equal(t, "example.com", ctx.PlatformName)
	require.Equal(t, "api-key", ctx.Credential.APIKeyPair.Key)
	require.Equal(t, "api-secret", ctx.Credential.APIKeyPair.Secret)
}

func TestConfig_UseContext(t *testing.T) {
	cfg := AuthenticatedCloudConfigMock()
	contextName := cfg.Context().Name
	cfg.CurrentContext = ""
	type fields struct {
		Config *Config
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "succeed setting valid context",
			fields: fields{
				Config: cfg,
			},
			args:    args{name: contextName},
			wantErr: false,
		},
		{
			name: "fail setting nonexistent context",
			fields: fields{
				Config: cfg,
			},
			args:    args{name: "some-context"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.Config
			if err := c.UseContext(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("UseContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.args.name, c.CurrentContext)
			}
		})
	}
}

func TestConfig_FindContext(t *testing.T) {
	type fields struct {
		Contexts map[string]*Context
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Context
		wantErr bool
	}{
		{name: "success finding existing context",
			fields:  fields{Contexts: map[string]*Context{"test-context": {Name: "test-context"}}},
			args:    args{name: "test-context"},
			want:    &Context{Name: "test-context"},
			wantErr: false,
		},
		{name: "error finding nonexistent context",
			fields:  fields{Contexts: map[string]*Context{}},
			args:    args{name: "test-context"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts: tt.fields.Contexts,
			}
			got, err := c.FindContext(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Context(t *testing.T) {
	type fields struct {
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name   string
		fields fields
		want   *Context
	}{
		{
			name: "succeed getting current context",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
				}},
				CurrentContext: "test-context",
			},
			want: &Context{
				Name: "test-context",
			},
		},
		{
			name: "error getting current context when not set",
			fields: fields{
				Contexts: map[string]*Context{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			got := c.Context()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Context() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKafkaClusterContext_SetAndGetActiveKafkaCluster_Env(t *testing.T) {
	testInputs := SetupTestInputs(true)
	ctx := testInputs.statefulConfig.Context()
	// temp file so json files in test_json do not get overwritten
	configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
	ctx.Config.Filename = configFile.Name()

	// Creating another environment with another kafka cluster
	otherAccountId := "other-abc"
	otherAccount := &orgv1.Account{
		Id:   otherAccountId,
		Name: "other-account",
	}
	otherKafkaClusterId := "other-kafka"
	otherKafkaCluster := &KafkaClusterConfig{
		ID:          otherKafkaClusterId,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*APIKeyPair{
			"akey": {
				Key:    "akey",
				Secret: "asecret",
			},
		},
		APIKey: "akey",
	}

	ctx.State.Auth.Accounts = append(ctx.State.Auth.Accounts, otherAccount)
	var activeKafka string

	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeKafka != testInputs.activeKafka {
		t.Errorf("GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)

	// switch environment add the kafka cluster, and set it as active cluster
	ctx.State.Auth.Account = otherAccount
	ctx.KafkaClusterContext.AddKafkaClusterConfig(otherKafkaCluster)
	ctx.KafkaClusterContext.SetActiveKafkaCluster(otherKafkaClusterId)
	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeKafka != otherKafkaClusterId {
		t.Errorf("After settting active kafka in new environment, GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)

	// switch environment back
	ctx.State.Auth.Account = testInputs.account
	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeKafka != testInputs.activeKafka {
		t.Errorf("After switching to back to first environment, GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)
	_ = os.Remove(configFile.Name())
}

func TestKafkaClusterContext_SetAndGetActiveKafkaCluster_NonEnv(t *testing.T) {
	testInputs := SetupTestInputs(false)
	ctx := testInputs.statefulConfig.Context()
	// temp file so json files in test_json do not get overwritten
	configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
	ctx.Config.Filename = configFile.Name()
	otherKafkaClusterId := "other-kafka"
	otherKafkaCluster := &KafkaClusterConfig{
		ID:          otherKafkaClusterId,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*APIKeyPair{
			"akey": {
				Key:    "akey",
				Secret: "asecret",
			},
		},
		APIKey: "akey",
	}

	activeKafka := ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeKafka != testInputs.activeKafka {
		t.Errorf("GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)

	// Add another kafka cluster and set it as active cluster
	ctx.KafkaClusterContext.AddKafkaClusterConfig(otherKafkaCluster)
	ctx.KafkaClusterContext.SetActiveKafkaCluster(otherKafkaClusterId)

	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if activeKafka != otherKafkaClusterId {
		t.Errorf("After settting active kafka, GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)
	_ = os.Remove(configFile.Name())
}

func TestKafkaClusterContext_AddAndGetKafkaClusterConfig(t *testing.T) {
	clusterID := "lkc-abcdefg"

	kcc := &KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*APIKeyPair{
			"akey": {
				Key:    "akey",
				Secret: "asecret",
			},
		},
		APIKey: "akey",
	}
	for _, isCloud := range []bool{true, false} {
		testInputs := SetupTestInputs(isCloud)
		kafkaClusterContext := testInputs.statefulConfig.Context().KafkaClusterContext
		kafkaClusterContext.AddKafkaClusterConfig(kcc)
		reflect.DeepEqual(kcc, kafkaClusterContext.GetKafkaClusterConfig(clusterID))
	}
}

func TestKafkaClusterContext_DeleteAPIKey(t *testing.T) {
	clusterID := "lkc-abcdefg"
	apiKey := "akey"
	kcc := &KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*APIKeyPair{
			apiKey: {
				Key:    apiKey,
				Secret: "asecret",
			},
		},
		APIKey: apiKey,
	}
	for _, isCloud := range []bool{true, false} {
		testInputs := SetupTestInputs(isCloud)
		kafkaClusterContext := testInputs.statefulConfig.Context().KafkaClusterContext
		kafkaClusterContext.AddKafkaClusterConfig(kcc)

		kafkaClusterContext.DeleteAPIKey(apiKey)
		kcc := kafkaClusterContext.GetKafkaClusterConfig(clusterID)
		if _, ok := kcc.APIKeys[apiKey]; ok {
			t.Errorf("DeleteAPIKey did not delete the API key.")
		}
		if kcc.APIKey != "" {
			t.Errorf("DeleteAPIKey did not remove deleted active API key.")
		}
	}
}

func TestKafkaClusterContext_RemoveKafkaCluster(t *testing.T) {
	clusterID := "lkc-abcdefg"
	apiKey := "akey"
	kcc := &KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*APIKeyPair{
			apiKey: {
				Key:    apiKey,
				Secret: "asecret",
			},
		},
		APIKey: apiKey,
	}
	for _, isCloud := range []bool{true, false} {
		testInputs := SetupTestInputs(isCloud)
		kafkaClusterContext := testInputs.statefulConfig.Context().KafkaClusterContext
		kafkaClusterContext.AddKafkaClusterConfig(kcc)
		kafkaClusterContext.SetActiveKafkaCluster(clusterID)
		require.Equal(t, clusterID, kafkaClusterContext.GetActiveKafkaClusterId())

		kafkaClusterContext.RemoveKafkaCluster(clusterID)
		_, ok := kafkaClusterContext.KafkaClusterConfigs[clusterID]
		require.False(t, ok)
		require.Empty(t, kafkaClusterContext.GetActiveKafkaClusterId())
	}
}

func TestConfig_IsCloud_True(t *testing.T) {
	platforms := []string{
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
		"confluent.cloud",
		"premium-oryx.gcp.priv.cpdev.cloud",
	}

	for _, platform := range platforms {
		cfg := &Config{
			Contexts:       map[string]*Context{"context": {PlatformName: platform}},
			CurrentContext: "context",
		}
		require.True(t, cfg.IsCloudLogin(), platform+" should be true")
	}
}

func TestConfig_IsCloud_False(t *testing.T) {
	configs := []*Config{
		nil,
		{
			Contexts:       map[string]*Context{"context": {PlatformName: "https://example.com"}},
			CurrentContext: "context",
		},
	}

	for _, cfg := range configs {
		require.False(t, cfg.IsCloudLogin())
	}
}

func TestConfig_IsOnPrem_True(t *testing.T) {
	cfg := &Config{
		Contexts:       map[string]*Context{"context": {PlatformName: "https://example.com"}},
		CurrentContext: "context",
	}
	require.True(t, cfg.IsOnPremLogin())
}

func TestConfig_IsOnPrem_False(t *testing.T) {
	configs := []*Config{
		nil,
		{
			Contexts:       map[string]*Context{"context": new(Context)},
			CurrentContext: "context",
		},
		{
			Contexts:       map[string]*Context{"context": {PlatformName: "confluent.cloud"}},
			CurrentContext: "context",
		},
	}

	for _, cfg := range configs {
		require.False(t, cfg.IsOnPremLogin())
	}
}
