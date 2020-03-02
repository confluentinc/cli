package v3

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
	testUtils "github.com/confluentinc/cli/test"
)

var (
	apiKeyString    = "abc-key-123"
	apiSecretString = "def-secret-456"
	kafkaClusterID  = "anonymous-id"
	contextName     = "my-context"
	accountID       = "acc-123"
)

type TestConfigs struct {
	kafkaClusters   map[string]*v1.KafkaClusterConfig
	activeKafka     string
	statefulConfig  *Config
	statelessConfig *Config
}

func SetupConfigs() *TestConfigs {
	testConfigs := &TestConfigs{}
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
		},
		AuthToken: "abc123",
	}
	testConfigs.kafkaClusters = map[string]*v1.KafkaClusterConfig{
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
	testConfigs.activeKafka = kafkaClusterID
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
	testConfigs.statefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params: &config.Params{
				CLIName:    "confluent",
				MetricSink: nil,
				Logger:     log.New(),
			},
			Filename: "test_json/stateful.json",
			Ver:      &Version,
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
	testConfigs.statelessConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Params: &config.Params{
				CLIName:    "confluent",
				MetricSink: nil,
				Logger:     log.New(),
			},
			Filename: "test_json/stateless.json",
			Ver:      &Version,
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

	statefulContext.Config = testConfigs.statefulConfig
	statefulContext.KafkaClusterContext = NewKafkaClusterContext(statefulContext, testConfigs.activeKafka, testConfigs.kafkaClusters)

	statelessContext.Config = testConfigs.statelessConfig
	statelessContext.KafkaClusterContext = NewKafkaClusterContext(statelessContext, testConfigs.activeKafka, testConfigs.kafkaClusters)
	return testConfigs
}

func TestConfig_Load(t *testing.T) {
	testConfigs := SetupConfigs()
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
		file    string
	}{
		{
			name: "succeed loading stateless config from file",
			want: testConfigs.statelessConfig,
			file: "test_json/stateless.json",
		},
		{
			name: "succeed loading config with state from file",
			want: testConfigs.statefulConfig,
			file: "test_json/stateful.json",
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
					Ver:      &Version,
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
				CLIName:    "confluent",
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
	testConfigs := SetupConfigs()
	tests := []struct {
		name     string
		config   *Config
		wantFile string
		wantErr  bool
	}{
		{
			name:     "save config with state to file",
			config:   testConfigs.statefulConfig,
			wantFile: "test_json/stateful.json",
		},
		{
			name:     "save stateless config to file",
			config:   testConfigs.statelessConfig,
			wantFile: "test_json/stateless.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
			tt.config.Filename = configFile.Name()
			if err := tt.config.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ioutil.ReadFile(configFile.Name())
			want, _ := ioutil.ReadFile(tt.wantFile)
			if testUtils.NormalizeNewLines(string(got)) != testUtils.NormalizeNewLines(string(want)) {
				t.Errorf("Config.Save() = %v\n want = %v", testUtils.NormalizeNewLines(string(got)), testUtils.NormalizeNewLines(string(want)))
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
			name: "config file for ccloud binary",
			fields: fields{
				CLIName: "ccloud",
			},
			want: filepath.FromSlash(os.Getenv("HOME") + "/.ccloud/config.json"),
		},
		{
			name: "config file for confluent binary",
			fields: fields{
				CLIName: "confluent",
			},
			want: filepath.FromSlash(os.Getenv("HOME") + "/.confluent/config.json"),
		},
		{
			name:   "should default to ~/.confluent if CLIName isn't provided",
			fields: fields{},
			want:   filepath.FromSlash(os.Getenv("HOME") + "/.confluent/config.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&config.Params{
				CLIName:    tt.fields.CLIName,
				MetricSink: nil,
				Logger:     log.New(),
			})
			got, err := c.getFilename()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.getFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Config.getFilename() = %v, want %v", got, tt.want)
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
	type testSturct struct {
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

	test := testSturct{
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

	tests := []testSturct{
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
	config := AuthenticatedCloudConfigMock()
	contextName := config.Context().Name
	config.CurrentContext = ""
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
				Config: config,
			},
			args:    args{name: contextName},
			wantErr: false,
		},
		{
			name: "fail setting nonexistent context",
			fields: fields{
				Config: config,
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
