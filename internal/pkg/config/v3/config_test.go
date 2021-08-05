package v3

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	apiKeyString    = "abc-key-123"
	apiSecretString = "def-secret-456"
	kafkaClusterID  = "anonymous-id"
	contextName     = "my-context"
	accountID       = "acc-123"
)

type TestInputs struct {
	kafkaClusters        map[string]*v1.KafkaClusterConfig
	activeKafka          string
	statefulConfig       *Config
	statelessConfig      *Config
	twoEnvStatefulConfig *Config
	account              *orgv1.Account
}

func SetupTestInputs(cliName string) *TestInputs {
	testInputs := &TestInputs{}
	platform := &v2.Platform{
		Name:   "http://test",
		Server: "http://test",
	}
	apiCredential := &v2.Credential{
		Name:     "api-key-abc-key-123",
		Username: "",
		Password: "",
		APIKeyPair: &v0.APIKeyPair{
			Key:    apiKeyString,
			Secret: apiSecretString,
		},
		CredentialType: 1,
	}
	loginCredential := &v2.Credential{
		Name:           "username-test-user",
		Username:       "test-user",
		Password:       "",
		APIKeyPair:     nil,
		CredentialType: 0,
	}
	account := &orgv1.Account{
		Id:   accountID,
		Name: "test-env",
	}
	account2 := &orgv1.Account{
		Id:   "env-flag",
		Name: "test-env2",
	}
	testInputs.account = account
	state := &v2.ContextState{
		Auth: &v1.AuthConfig{
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
	twoEnvState := &v2.ContextState{
		Auth: &v1.AuthConfig{
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
	testInputs.kafkaClusters = map[string]*v1.KafkaClusterConfig{
		kafkaClusterID: {
			ID:          kafkaClusterID,
			Name:        "anonymous-cluster",
			Bootstrap:   "http://test",
			APIEndpoint: "",
			APIKeys: map[string]*v0.APIKeyPair{
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
		SchemaRegistryClusters: map[string]*v2.SchemaRegistryCluster{
			accountID: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State:  state,
		Logger: log.New(),
	}
	statelessContext := &Context{
		Name:                   contextName,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             apiCredential,
		CredentialName:         apiCredential.Name,
		SchemaRegistryClusters: map[string]*v2.SchemaRegistryCluster{},
		State:                  &v2.ContextState{},
		Logger:                 log.New(),
	}
	twoEnvStatefulContext := &Context{
		Name:           contextName,
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		SchemaRegistryClusters: map[string]*v2.SchemaRegistryCluster{
			accountID: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State:  twoEnvState,
		Logger: log.New(),
	}
	testInputs.statefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params: &config.Params{
				CLIName:    cliName,
				MetricSink: nil,
				Logger:     log.New(),
			},
			Filename: fmt.Sprintf("test_json/stateful_%s.json", cliName),
			Ver:      Version,
		},
		Platforms: map[string]*v2.Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*v2.Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: statefulContext,
		},
		ContextStates: map[string]*v2.ContextState{
			contextName: state,
		},
		CurrentContext: contextName,
	}
	testInputs.statelessConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params: &config.Params{
				CLIName:    cliName,
				MetricSink: nil,
				Logger:     log.New(),
			},
			Filename: fmt.Sprintf("test_json/stateless_%s.json", cliName),
			Ver:      Version,
		},
		Platforms: map[string]*v2.Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*v2.Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: statelessContext,
		},
		ContextStates: map[string]*v2.ContextState{
			contextName: {},
		},
		CurrentContext: contextName,
	}
	testInputs.twoEnvStatefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params: &config.Params{
				CLIName:    cliName,
				MetricSink: nil,
				Logger:     log.New(),
			},
			Filename: fmt.Sprintf("test_json/stateful_%s.json", cliName),
			Ver:      Version,
		},
		Platforms: map[string]*v2.Platform{
			platform.Name: platform,
		},
		Credentials: map[string]*v2.Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts: map[string]*Context{
			contextName: twoEnvStatefulContext,
		},
		ContextStates: map[string]*v2.ContextState{
			contextName: twoEnvState,
		},
		CurrentContext: contextName,
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
	testConfigsConfluent := SetupTestInputs("confluent")
	testConfigsCcloud := SetupTestInputs("ccloud")
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
		file    string
	}{
		{
			name: "succeed loading stateless confluent config from file",
			want: testConfigsConfluent.statelessConfig,
			file: "test_json/stateless_confluent.json",
		},
		{
			name: "succeed loading confluent config with state from file",
			want: testConfigsConfluent.statefulConfig,
			file: "test_json/stateful_confluent.json",
		},
		{
			name: "succeed loading stateless ccloud config from file",
			want: testConfigsCcloud.statelessConfig,
			file: "test_json/stateless_ccloud.json",
		},
		{
			name: "succeed loading ccloud config with state from file",
			want: testConfigsCcloud.statefulConfig,
			file: "test_json/stateful_ccloud.json",
		},
		{
			name: "should load disable update checks and disable updates",
			want: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: "test_json/load_disable_update.json",
					Ver:      Version,
				},
				DisableUpdates:     true,
				DisableUpdateCheck: true,
				Platforms:          map[string]*v2.Platform{},
				Credentials:        map[string]*v2.Credential{},
				Contexts:           map[string]*Context{},
				ContextStates:      map[string]*v2.ContextState{},
			},
			file: "test_json/load_disable_update.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Params{
				CLIName:    tt.want.CLIName,
				MetricSink: nil,
				Logger:     log.New(),
			})
			c.Filename = tt.file
			for _, context := range tt.want.Contexts {
				context.Config = tt.want
			}
			if err := c.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %+v, wantErr %+v", err, tt.wantErr)
			}
			// Get around automatically assigned anonymous id
			tt.want.AnonymousId = c.AnonymousId
			if !t.Failed() && !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %+v, want %+v", c, tt.want)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	testConfigsConfluent := SetupTestInputs("confluent")
	testConfigsCcloud := SetupTestInputs("ccloud")
	tests := []struct {
		name             string
		config           *Config
		wantFile         string
		wantErr          bool
		kafkaOverwrite   string
		contextOverwrite string
		accountOverwrite *orgv1.Account
	}{
		{
			name:     "save confluent config with state to file",
			config:   testConfigsConfluent.statefulConfig,
			wantFile: "test_json/stateful_confluent.json",
		},
		{
			name:     "save stateless confluent config to file",
			config:   testConfigsConfluent.statelessConfig,
			wantFile: "test_json/stateless_confluent.json",
		},
		{
			name:     "save ccloud config with state to file",
			config:   testConfigsCcloud.statefulConfig,
			wantFile: "test_json/stateful_ccloud.json",
		},
		{
			name:     "save stateless ccloud config to file",
			config:   testConfigsCcloud.statelessConfig,
			wantFile: "test_json/stateless_ccloud.json",
		},
		{
			name:           "save stateless ccloud config with kafka overwrite to file",
			config:         testConfigsCcloud.statefulConfig,
			wantFile:       "test_json/stateful_ccloud.json",
			kafkaOverwrite: "lkc-clusterFlag",
		},
		{
			name:           "save stateless ccloud config with kafka and context overwrite to file",
			config:         testConfigsCcloud.statefulConfig,
			wantFile:       "test_json/stateful_ccloud.json",
			kafkaOverwrite: "lkc-clusterFlag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
			tt.config.Filename = configFile.Name()
			ctx := tt.config.Context()
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
			got, _ := ioutil.ReadFile(configFile.Name())
			want, _ := ioutil.ReadFile(tt.wantFile)
			if utils.NormalizeNewLines(string(got)) != utils.NormalizeNewLines(string(want)) {
				t.Errorf("Config.Save() = %v\n want = %v", utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(string(want)))
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

func TestConfig_SaveWithAccountOverwrite(t *testing.T) {
	testConfigsCcloud := SetupTestInputs("ccloud")
	tests := []struct {
		name             string
		config           *Config
		wantFile         string
		wantErr          bool
		accountOverwrite *orgv1.Account
	}{
		{
			name:             "save ccloud config with state and account overwrite to file",
			config:           testConfigsCcloud.twoEnvStatefulConfig,
			wantFile:         "test_json/account_overwrite.json",
			accountOverwrite: &orgv1.Account{Id: "env-flag"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
			tt.config.Filename = configFile.Name()
			if tt.accountOverwrite != nil {
				tt.config.SetOverwrittenAccount(tt.config.Context().State.Auth.Account)
				tt.config.Context().State.Auth.Account = tt.accountOverwrite
			}
			if err := tt.config.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ioutil.ReadFile(configFile.Name())
			got = append(got, '\n') //account for extra newline at the end of the json file
			want, _ := ioutil.ReadFile(tt.wantFile)
			if utils.NormalizeNewLines(string(got)) != utils.NormalizeNewLines(string(want)) {
				t.Errorf("Config.Save() = %v\n want = %v", utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(string(want)))
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

func TestConfig_OverwrittenKafka(t *testing.T) {
	//	testConfigsConfluent := SetupTestInputs("confluent")
	testConfigsCcloud := SetupTestInputs("ccloud")
	//testConfigsCcloud2 := SetupTestInputs("ccloud")

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string //simulates initial config value overwritten by a cluster flag value
		activeKafka    string //simulates the cluster flag value
	}{
		{
			name:        "test no overwrite value",
			config:      testConfigsCcloud.statefulConfig,
			activeKafka: testConfigsCcloud.activeKafka,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCcloud.statefulConfig,
			overwrittenVal: "lkc-test",
			activeKafka:    testConfigsCcloud.activeKafka,
		},
		{
			name:        "test no overwrite value",
			config:      testConfigsCcloud.statelessConfig,
			activeKafka: testConfigsCcloud.activeKafka,
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
	testConfigsCcloud := SetupTestInputs("ccloud")

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string //simulates initial context value overwritten by a context flag value
		currContext    string //simulates the context flag value
	}{
		{
			name:        "test no overwrite value",
			config:      testConfigsCcloud.statefulConfig,
			currContext: testConfigsCcloud.statefulConfig.CurrentContext,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCcloud.statefulConfig,
			overwrittenVal: "test-context",
			currContext:    testConfigsCcloud.statefulConfig.CurrentContext,
		},
		{
			name:        "test no overwrite value",
			config:      testConfigsCcloud.statelessConfig,
			currContext: testConfigsCcloud.statelessConfig.CurrentContext,
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
	testConfigsCcloud := SetupTestInputs("ccloud")

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal *orgv1.Account //simulates initial environment (account) value overwritten by a environment flag
		activeAccount  string         //simulates the environment (account) flag value
	}{
		{
			name:          "test no overwrite value",
			config:        testConfigsCcloud.statefulConfig,
			activeAccount: testConfigsCcloud.statefulConfig.Context().State.Auth.Account.Id,
		},
		{
			name:           "test with overwrite value",
			config:         testConfigsCcloud.statefulConfig,
			overwrittenVal: &orgv1.Account{Id: "env-test"},
			activeAccount:  testConfigsCcloud.statefulConfig.Context().State.Auth.Account.Id,
		},
		{
			name:   "test no overwrite value",
			config: testConfigsCcloud.statelessConfig,
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
	type fields struct {
		CLIName string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "config filepath is ~/.confluent/config.json",
			want: filepath.FromSlash(os.Getenv("HOME") + "/.confluent/config.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Params{
				CLIName:    tt.fields.CLIName,
				MetricSink: nil,
				Logger:     log.New(),
			})
			got := c.GetFilename()
			if got != tt.want {
				t.Errorf("Config.GetFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_AddContext(t *testing.T) {
	filename := "/tmp/TestConfig_AddContext.json"
	conf := AuthenticatedConfluentConfigMock()
	conf.Filename = filename
	context := conf.Context()
	noContextConf := AuthenticatedConfluentConfigMock()
	noContextConf.Filename = filename
	delete(noContextConf.Contexts, noContextConf.Context().Name)
	noContextConf.CurrentContext = ""
	type testStruct struct {
		name                   string
		config                 *Config
		contextName            string
		platformName           string
		credentialName         string
		kafkaClusters          map[string]*v1.KafkaClusterConfig
		kafka                  string
		schemaRegistryClusters map[string]*v2.SchemaRegistryCluster
		state                  *v2.ContextState
		Version                *version.Version
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
				tt.schemaRegistryClusters, tt.state)
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

func TestConfig_SetContext(t *testing.T) {
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
			if err := c.SetContext(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("SetContext() error = %v, wantErr %v", err, tt.wantErr)
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

func TestConfig_DeleteContext(t *testing.T) {
	const contextName = "test-context"
	type fields struct {
		Contexts       map[string]*Context
		CurrentContext string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantConfig *Config
	}{
		{name: "succeed deleting existing current context",
			fields: fields{
				Contexts:       map[string]*Context{contextName: {Name: contextName}},
				CurrentContext: contextName,
			},
			args:    args{name: contextName},
			wantErr: false,
			wantConfig: &Config{
				Contexts:       map[string]*Context{},
				CurrentContext: "",
			},
		},
		{name: "succeed deleting existing context",
			fields: fields{Contexts: map[string]*Context{
				contextName:     {Name: contextName},
				"other-context": {Name: "other-context"},
			},
				CurrentContext: "other-context",
			},
			args:    args{name: contextName},
			wantErr: false,
			wantConfig: &Config{
				Contexts:       map[string]*Context{"other-context": {Name: "other-context"}},
				CurrentContext: "other-context",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			if err := c.DeleteContext(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("DeleteContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.wantConfig, c)
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
	testInputs := SetupTestInputs("ccloud")
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
	otherKafkaCluster := &v1.KafkaClusterConfig{
		ID:          otherKafkaClusterId,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*v0.APIKeyPair{
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
	testInputs := SetupTestInputs("confluent")
	ctx := testInputs.statefulConfig.Context()
	// temp file so json files in test_json do not get overwritten
	configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
	ctx.Config.Filename = configFile.Name()
	otherKafkaClusterId := "other-kafka"
	otherKafkaCluster := &v1.KafkaClusterConfig{
		ID:          otherKafkaClusterId,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*v0.APIKeyPair{
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

	kcc := &v1.KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*v0.APIKeyPair{
			"akey": {
				Key:    "akey",
				Secret: "asecret",
			},
		},
		APIKey: "akey",
	}
	for _, cliName := range []string{"ccloud", "confluent"} {
		testInputs := SetupTestInputs(cliName)
		kafkaClusterContext := testInputs.statefulConfig.Context().KafkaClusterContext
		kafkaClusterContext.AddKafkaClusterConfig(kcc)
		reflect.DeepEqual(kcc, kafkaClusterContext.GetKafkaClusterConfig(clusterID))
	}
}

func TestKafkaClusterContext_DeleteAPIKey(t *testing.T) {
	clusterID := "lkc-abcdefg"
	apiKey := "akey"
	kcc := &v1.KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*v0.APIKeyPair{
			apiKey: {
				Key:    apiKey,
				Secret: "asecret",
			},
		},
		APIKey: apiKey,
	}
	for _, cliName := range []string{"ccloud", "confluent"} {
		testInputs := SetupTestInputs(cliName)
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
	kcc := &v1.KafkaClusterConfig{
		ID:          clusterID,
		Name:        "lit",
		Bootstrap:   "http://test",
		APIEndpoint: "",
		APIKeys: map[string]*v0.APIKeyPair{
			apiKey: {
				Key:    apiKey,
				Secret: "asecret",
			},
		},
		APIKey: apiKey,
	}
	for _, cliName := range []string{"ccloud", "confluent"} {
		testInputs := SetupTestInputs(cliName)
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
		config := &Config{
			Contexts:       map[string]*Context{"context": {PlatformName: platform}},
			CurrentContext: "context",
		}
		require.True(t, config.IsCloudLogin(), platform+" should be true")
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

	for _, config := range configs {
		require.False(t, config.IsCloudLogin())
	}
}

func TestConfig_IsOnPrem_True(t *testing.T) {
	config := &Config{
		Contexts:       map[string]*Context{"context": {PlatformName: "https://example.com"}},
		CurrentContext: "context",
	}
	require.True(t, config.IsOnPremLogin())
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

	for _, config := range configs {
		require.False(t, config.IsOnPremLogin())
	}
}
