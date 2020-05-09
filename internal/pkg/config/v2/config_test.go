package v2

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
	"github.com/confluentinc/cli/internal/pkg/version"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func TestConfig_Load(t *testing.T) {
	platform := &Platform{
		Name:   "http://test",
		Server: "http://test",
	}
	apiCredential := &Credential{
		Name:     "api-key-abc-key-123",
		Username: "",
		Password: "",
		APIKeyPair: &v0.APIKeyPair{
			Key:    "abc-key-123",
			Secret: "def-secret-456",
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
	state := &ContextState{
		Auth: &v1.AuthConfig{
			User: &orgv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Account: &orgv1.Account{
				Id:   "acc-123",
				Name: "test-env",
			},
			Accounts: []*orgv1.Account{
				{
					Id:   "acc-123",
					Name: "test-env",
				},
			},
		},
		AuthToken: "abc123",
	}
	statefulContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		KafkaClusters: map[string]*v1.KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*v0.APIKeyPair{
					"abc-key-123": {
						Key:    "abc-key-123",
						Secret: "def-secret-456",
					},
				},
				APIKey: "abc-key-123",
			},
		},
		Kafka: "anonymous-id",
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			"acc-123": {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State:  state,
		Logger: log.New(),
	}
	statelessContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     apiCredential,
		CredentialName: apiCredential.Name,
		KafkaClusters: map[string]*v1.KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*v0.APIKeyPair{
					"abc-key-123": {
						Key:    "abc-key-123",
						Secret: "def-secret-456",
					},
				},
				APIKey: "abc-key-123",
			},
		},
		Kafka:                  "anonymous-id",
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{},
		State:                  &ContextState{},
		Logger:                 log.New(),
	}
	testConfigFile, _ := ioutil.TempFile("", "TestConfig_Load.json")
	type args struct {
		contents string
	}
	tests := []struct {
		name    string
		args    *args
		want    *Config
		wantErr bool
		file    string
	}{
		{
			name: "succeed loading stateless config from file",
			args: &args{
				contents: "{\"platforms\":{\"http://test\":{\"name\":\"http://test\",\"server\":\"http://test\"}}," +
					"\"credentials\":{\"api-key-abc-key-123\":{\"Name\":\"api-key-abc-key-123\",\"username\":\"\"," +
					"\"password\":\"\",\"api_key_pair\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"def-secret-456\"}," +
					"\"credential_type\":1}},\"contexts\":{\"my-context\":{\"name\":\"my-context\",\"platform\":\"http://test\"," +
					"\"credential\":\"api-key-abc-key-123\",\"kafka_clusters\":{\"anonymous-id\":{\"id\":\"anonymous-id\",\"name\":\"anonymous-cluster\"," +
					"\"bootstrap_servers\":\"http://test\",\"api_keys\":{\"abc-key-123\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"def-secret-456\"}}," +
					"\"api_key\":\"abc-key-123\"}},\"kafka_cluster\":\"anonymous-id\",\"schema_registry_clusters\":{}}},\"context_states\":{\"my-context\":{" +
					"\"auth\":null,\"auth_token\":\"\"}},\"current_context\":\"my-context\"}",
			},
			want: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: testConfigFile.Name(),
					Ver:      Version,
				},
				Platforms: map[string]*Platform{
					platform.Name: platform,
				},
				Credentials: map[string]*Credential{
					apiCredential.Name: apiCredential,
				},
				Contexts: map[string]*Context{
					"my-context": statelessContext,
				},
				ContextStates: map[string]*ContextState{
					"my-context": {},
				},
				CurrentContext: "my-context",
			},
			file: testConfigFile.Name(),
		},
		{
			name: "succeed loading config with state from file",
			args: &args{
				contents: "{\"platforms\":{\"http://test\":{\"name\":\"http://test\",\"server\":\"http://test\"}}," +
					"\"credentials\":{\"username-test-user\":{\"name\":\"username-test-user\",\"username\":\"test-user\"," +
					"\"password\":\"\",\"api_key_pair\":null,\"CredentialType\":0}},\"contexts\":{\"my-context\":{\"name\":\"" +
					"my-context\",\"platform\":\"http://test\",\"credential\":\"username-test-user\",\"kafka_clusters\":{\"" +
					"anonymous-id\":{\"id\":\"anonymous-id\",\"name\":\"anonymous-cluster\",\"bootstrap_servers\"" +
					":\"http://test\",\"api_keys\":{\"abc-key-123\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"def-secret-456\"}}," +
					"\"api_key\":\"abc-key-123\"}},\"kafka_cluster\":\"anonymous-id\",\"schema_registry_clusters\":{\"" +
					"acc-123\":{\"id\":\"lsrc-123\",\"schema_registry_endpoint\":\"http://some-lsrc-endpoint\",\"" +
					"schema_registry_credentials\":null}}}},\"context_states\":{\"my-context\":{\"auth\":{\"user\"" +
					":{\"id\":123,\"email\":\"test-user@email\"},\"account\":{\"id\":\"acc-123\",\"name\":\"test-env\"" +
					"},\"accounts\":[{\"id\":\"acc-123\",\"name\":\"test-env\"}]},\"auth_token\":\"abc123\"}},\"" +
					"current_context\":\"my-context\"}",
			},
			want: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: testConfigFile.Name(),
					Ver:      Version,
				},
				Platforms: map[string]*Platform{
					platform.Name: platform,
				},
				Credentials: map[string]*Credential{
					loginCredential.Name: loginCredential,
				},
				Contexts: map[string]*Context{
					"my-context": statefulContext,
				},
				CurrentContext: "my-context",
				ContextStates: map[string]*ContextState{
					"my-context": state,
				},
			},
			file: testConfigFile.Name(),
		},
		{
			name: "should load disable update checks and disable updates",
			args: &args{
				contents: "{\"disable_update_check\": true, \"disable_updates\": true}",
			},
			want: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: testConfigFile.Name(),
					Ver:      Version,
				},
				DisableUpdates:     true,
				DisableUpdateCheck: true,
				Platforms:          map[string]*Platform{},
				Credentials:        map[string]*Credential{},
				Contexts:           map[string]*Context{},
				ContextStates:      map[string]*ContextState{},
			},
			file: testConfigFile.Name(),
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
			err := ioutil.WriteFile(tt.file, []byte(tt.args.contents), 0644)
			if err != nil {
				t.Errorf("unable to test config to file: %+v", err)
			}
			if err := c.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %+v, wantErr %+v", err, tt.wantErr)
			}
			fmt.Println(tt.args.contents)
			// Get around automatically assigned anonymous id
			tt.want.AnonymousId = c.AnonymousId
			if !t.Failed() && !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %+v, want %+v", c, tt.want)
			}

			os.Remove(tt.file)
		})
	}
}

func TestConfig_Save(t *testing.T) {
	platform := &Platform{
		Name:   "http://test",
		Server: "http://test",
	}
	apiCredential := &Credential{
		Name:     "api-key-abc-key-123",
		Username: "",
		Password: "",
		APIKeyPair: &v0.APIKeyPair{
			Key:    "abc-key-123",
			Secret: "def-secret-456",
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
	state := &ContextState{
		Auth: &v1.AuthConfig{
			User: &orgv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Account: &orgv1.Account{
				Id:   "acc-123",
				Name: "test-env",
			},
			Accounts: []*orgv1.Account{
				{
					Id:   "acc-123",
					Name: "test-env",
				},
			},
		},
		AuthToken: "abc123",
	}
	statefulContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		KafkaClusters: map[string]*v1.KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*v0.APIKeyPair{
					"abc-key-123": {
						Key:    "abc-key-123",
						Secret: "def-secret-456",
					},
				},
				APIKey: "abc-key-123",
			},
		},
		Kafka: "anonymous-id",
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			"acc-123": {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State:  state,
		Logger: log.New(),
	}
	statelessContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     apiCredential,
		CredentialName: apiCredential.Name,
		KafkaClusters: map[string]*v1.KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*v0.APIKeyPair{
					"abc-key-123": {
						Key:    "abc-key-123",
						Secret: "def-secret-456",
					},
				},
				APIKey: "abc-key-123",
			},
		},
		Kafka:                  "anonymous-id",
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{},
		State:                  &ContextState{},
		Logger:                 log.New(),
	}
	tests := []struct {
		name    string
		config  *Config
		want    string
		wantErr bool
	}{
		{
			name: "save config with state to file",
			config: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: "",
					Ver:      Version,
				},
				Platforms: map[string]*Platform{
					platform.Name: platform,
				},
				Credentials: map[string]*Credential{
					apiCredential.Name:   apiCredential,
					loginCredential.Name: loginCredential,
				},
				Contexts: map[string]*Context{
					"my-context": statefulContext,
				},
				ContextStates: map[string]*ContextState{
					"my-context": state,
				},
				CurrentContext: "my-context",
			},
			want: "{\n  \"version\": \"2.0.0\",\n  \"disable_update_check\": false,\n  \"disable_updates\": false,\n  \"no_browser\": false,\n  \"platforms\": {\n    \"http://test\": {\n      \"name\": \"http://test\",\n      \"server\": \"http://test\"\n    }\n  },\n  \"credentials\": {\n    \"api-key-abc-key-123\": {\n      \"name\": \"api-key-abc-key-123\",\n      \"username\": \"\",\n      \"password\": \"\",\n      \"api_key_pair\": {\n        \"api_key\": \"abc-key-123\",\n        \"api_secret\": \"def-secret-456\"\n      },\n      \"credential_type\": 1\n    },\n    \"username-test-user\": {\n      \"name\": \"username-test-user\",\n      \"username\": \"test-user\",\n      \"password\": \"\",\n      \"api_key_pair\": null,\n      \"credential_type\": 0\n    }\n  },\n  \"contexts\": {\n    \"my-context\": {\n      \"name\": \"my-context\",\n      \"platform\": \"http://test\",\n      \"credential\": \"username-test-user\",\n      \"kafka_clusters\": {\n        \"anonymous-id\": {\n          \"id\": \"anonymous-id\",\n          \"name\": \"anonymous-cluster\",\n          \"bootstrap_servers\": \"http://test\",\n          \"api_keys\": {\n            \"abc-key-123\": {\n              \"api_key\": \"abc-key-123\",\n              \"api_secret\": \"def-secret-456\"\n            }\n          },\n          \"api_key\": \"abc-key-123\"\n        }\n      },\n      \"kafka_cluster\": \"anonymous-id\",\n      \"schema_registry_clusters\": {\n        \"acc-123\": {\n          \"id\": \"lsrc-123\",\n          \"schema_registry_endpoint\": \"http://some-lsrc-endpoint\",\n          \"schema_registry_credentials\": null\n        }\n      }\n    }\n  },\n  \"context_states\": {\n    \"my-context\": {\n      \"auth\": {\n        \"user\": {\n          \"id\": 123,\n          \"email\": \"test-user@email\"\n        },\n        \"account\": {\n          \"id\": \"acc-123\",\n          \"name\": \"test-env\"\n        },\n        \"accounts\": [\n          {\n            \"id\": \"acc-123\",\n            \"name\": \"test-env\"\n          }\n        ]\n      },\n      \"auth_token\": \"abc123\"\n    }\n  },\n  \"current_context\": \"my-context\"\n}",
		},
		{
			name: "save stateless config to file",
			config: &Config{
				BaseConfig: &config.BaseConfig{
					Params: &config.Params{
						CLIName:    "confluent",
						MetricSink: nil,
						Logger:     log.New(),
					},
					Filename: "",
					Ver:      Version,
				},
				Platforms: map[string]*Platform{
					platform.Name: platform,
				},
				Credentials: map[string]*Credential{
					apiCredential.Name: apiCredential,
				},
				Contexts: map[string]*Context{
					"my-context": statelessContext,
				},
				ContextStates: map[string]*ContextState{
					"my-context": {},
				},
				CurrentContext: "my-context",
			},
			want: "{\n  \"version\": \"2.0.0\",\n  \"disable_update_check\": false,\n  \"disable_updates\": false,\n  \"no_browser\": false,\n  \"platforms\": {\n    \"http://test\": {\n      \"name\": \"http://test\",\n      \"server\": \"http://test\"\n    }\n  },\n  \"credentials\": {\n    \"api-key-abc-key-123\": {\n      \"name\": \"api-key-abc-key-123\",\n      \"username\": \"\",\n      \"password\": \"\",\n      \"api_key_pair\": {\n        \"api_key\": \"abc-key-123\",\n        \"api_secret\": \"def-secret-456\"\n      },\n      \"credential_type\": 1\n    }\n  },\n  \"contexts\": {\n    \"my-context\": {\n      \"name\": \"my-context\",\n      \"platform\": \"http://test\",\n      \"credential\": \"api-key-abc-key-123\",\n      \"kafka_clusters\": {\n        \"anonymous-id\": {\n          \"id\": \"anonymous-id\",\n          \"name\": \"anonymous-cluster\",\n          \"bootstrap_servers\": \"http://test\",\n          \"api_keys\": {\n            \"abc-key-123\": {\n              \"api_key\": \"abc-key-123\",\n              \"api_secret\": \"def-secret-456\"\n            }\n          },\n          \"api_key\": \"abc-key-123\"\n        }\n      },\n      \"kafka_cluster\": \"anonymous-id\",\n      \"schema_registry_clusters\": {}\n    }\n  },\n  \"context_states\": {\n    \"my-context\": {\n      \"auth\": null,\n      \"auth_token\": \"\"\n    }\n  },\n  \"current_context\": \"my-context\"\n}",
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
			if string(got) != tt.want {
				t.Errorf("Config.Save() = %v, want contains %v", string(got), tt.want)
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
	conf := AuthenticatedConfigMock()
	conf.Filename = filename
	context := conf.Context()
	noContextConf := AuthenticatedConfigMock()
	noContextConf.Filename = filename
	delete(noContextConf.Contexts, noContextConf.Context().Name)
	noContextConf.CurrentContext = ""
	tests := []struct {
		name                   string
		config                 *Config
		contextName            string
		platform               *Platform
		platformName           string
		credentialName         string
		credential             *Credential
		kafkaClusters          map[string]*v1.KafkaClusterConfig
		kafka                  string
		schemaRegistryClusters map[string]*SchemaRegistryCluster
		state                  *ContextState
		Version                *version.Version
		filename               string
		want                   *Config
		wantErr                bool
	}{
		{
			name:                   "add valid context",
			config:                 noContextConf,
			contextName:            context.Name,
			platformName:           context.PlatformName,
			credentialName:         context.CredentialName,
			kafkaClusters:          context.KafkaClusters,
			kafka:                  context.Kafka,
			schemaRegistryClusters: context.SchemaRegistryClusters,
			state:                  context.State,
			filename:               filename,
			want:                   conf,
			wantErr:                false,
		},
		{
			name:                   "fail adding existing context",
			config:                 conf,
			contextName:            context.Name,
			platformName:           context.PlatformName,
			credentialName:         context.CredentialName,
			kafkaClusters:          context.KafkaClusters,
			kafka:                  context.Kafka,
			schemaRegistryClusters: context.SchemaRegistryClusters,
			state:                  context.State,
			filename:               filename,
			want:                   nil,
			wantErr:                true,
		},
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
	config := AuthenticatedConfigMock()
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

//func TestConfig_AuthenticatedState(t *testing.T) {
//	type fields struct {
//		CLIName        string
//		MetricSink     metric.Sink
//		Logger         *log.Logger
//		Filename       string
//		Platforms      map[string]*Platform
//		Credentials    map[string]*Credential
//		Contexts       map[string]*Context
//		CurrentContext string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//		want    *ContextState
//	}{
//		{
//			name: "succeed checking authenticated state of user with auth token",
//			fields: fields{
//				Credentials: map[string]*Credential{"current-cred": {
//					CredentialType: Username,
//				}},
//				Contexts: map[string]*Context{"current-context": {
//					Credential: &Credential{
//						CredentialType: Username,
//					},
//					CredentialName: "current-cred",
//					State: &ContextState{
//						Auth: &AuthConfig{
//							Account: &orgv1.Account{
//								Id: "abc123",
//							},
//							Accounts: nil,
//						},
//						AuthToken: "nekot",
//					},
//				}},
//				CurrentContext: "current-context",
//			},
//			wantErr: false,
//			want: &ContextState{
//				Auth: &AuthConfig{
//					Account: &orgv1.Account{
//						Id: "abc123",
//					},
//					Accounts: nil,
//				},
//				AuthToken: "nekot",
//			},
//		},
//		{
//			name: "error when authenticated state of user without auth token with username creds",
//			fields: fields{
//				Credentials: map[string]*Credential{"current-cred": {
//					CredentialType: Username,
//				}},
//				Contexts: map[string]*Context{"current-context": {
//					Credential: &Credential{
//						CredentialType: Username,
//					},
//					CredentialName: "current-cred",
//				}},
//				CurrentContext: "current-context",
//			},
//			wantErr: true,
//		},
//		{
//			name: "error when checking authenticated state of user with API key creds",
//			fields: fields{
//				Credentials: map[string]*Credential{"current-cred": {
//					CredentialType: APIKey,
//				}},
//				Contexts: map[string]*Context{"current-context": {
//					Credential: &Credential{
//						CredentialType: APIKey,
//					},
//					CredentialName: "current-cred",
//				}},
//				CurrentContext: "current-context",
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &Config{
//				CLIName:        tt.fields.CLIName,
//				MetricSink:     tt.fields.MetricSink,
//				Logger:         tt.fields.Logger,
//				Filename:       tt.fields.Filename,
//				Platforms:      tt.fields.Platforms,
//				Credentials:    tt.fields.Credentials,
//				Contexts:       tt.fields.Contexts,
//				CurrentContext: tt.fields.CurrentContext,
//			}
//			dc := &cmd.DynamicConfig{
//				Config:              c,
//				Resolver:            nil,
//				Client:              nil,
//			}
//			got, err := dc.AuthenticatedState()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("AuthenticatedState() error = %v, wantErr %v", err, tt.wantErr)
//			}
//			if !reflect.DeepEqual(got, c.Context().State) {
//				t.Errorf("AuthenticatedState() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

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

//func TestConfig_SchemaRegistryCluster(t *testing.T) {
//	conf := AuthenticatedConfigMock()
//	context := conf.Context()
//	srCluster := context.SchemaRegistryClusters[context.State.Auth.Account.Id]
//	noAuthConf := AuthenticatedConfigMock()
//	noAuthConf.Context().State = new(ContextState)
//	noAuthConf.ContextStates[noAuthConf.Context().Name] = new(ContextState)
//	tests := []struct {
//		name    string
//		config  *Config
//		want    *SchemaRegistryCluster
//		wantErr bool
//		err     error
//	}{
//		{
//			name:    "succeed getting existing schema registry cluster",
//			config:  conf,
//			want:    srCluster,
//			wantErr: false,
//		},
//		{
//			name:    "error getting schema registry cluster without current context",
//			config:  New(),
//			wantErr: true,
//			err:     cerrors.ErrNoContext,
//		},
//		{
//			name:    "error getting schema registry cluster when not logged in",
//			config:  noAuthConf,
//			wantErr: true,
//			err:     cerrors.ErrNotLoggedIn,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := tt.config.SchemaRegistryCluster(mock.NewClientMock())
//			if (err != nil) != tt.wantErr {
//				t.Errorf("SchemaRegistryCluster() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("SchemaRegistryCluster() got = %v, want %v", got, tt.want)
//			}
//			if tt.err != nil {
//				assert.Equal(t, tt.err, err)
//			}
//		})
//	}
//}

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
