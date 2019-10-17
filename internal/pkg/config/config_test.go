package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/stretchr/testify/assert"

	cerrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/confluentinc/cli/internal/pkg/sdk"
	"github.com/confluentinc/cli/internal/pkg/version"
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
		APIKeyPair: &APIKeyPair{
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
		Auth: &AuthConfig{
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
	ver := &version.Version{
		Binary:    "confluent",
		Name:      "Confluent CLI",
		Version:   "",
		Commit:    "",
		BuildDate: "",
		BuildHost: "",
		UserAgent: "Confluent-CLI/ (https://confluent.io; support@confluent.io)",
	}
	client := AuthenticatedConfigMock().Context().Client
	statefulContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     loginCredential,
		CredentialName: loginCredential.Name,
		KafkaClusters: map[string]*KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*APIKeyPair{
					"abc-key-123": {
						Key: "abc-key-123",
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
		State:   state,
		Client:  client,
		Version: ver,
		Logger:  log.New(),
	}
	statefulContext.Resolver = NewResolver(statefulContext, statefulContext.Client)
	statelessContext := &Context{
		Name:           "my-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     apiCredential,
		CredentialName: apiCredential.Name,
		KafkaClusters: map[string]*KafkaClusterConfig{
			"anonymous-id": {
				ID:          "anonymous-id",
				Name:        "anonymous-cluster",
				Bootstrap:   "http://test",
				APIEndpoint: "",
				APIKeys: map[string]*APIKeyPair{
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
		Client:                 client,
		Logger:                 log.New(),
		Version:                ver,
	}
	statelessContext.Resolver = NewResolver(statelessContext, statelessContext.Client)
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
				contents: "{\"platforms\":{\"http://test\":{\"Name\":\"http://test\",\"server\":\"http://test\"}}," +
					"\"credentials\":{\"api-key-abc-key-123\":{\"Name\":\"api-key-abc-key-123\",\"Username\":\"\"," +
					"\"Password\":\"\",\"APIKeyPair\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"def-secret-456\"}," +
					"\"CredentialType\":1}},\"contexts\":{\"my-context\":{\"name\":\"my-context\",\"platform\":\"http://test\"," +
					"\"credential\":\"api-key-abc-key-123\",\"kafka_clusters\":{\"anonymous-id\":{\"id\":\"anonymous-id\",\"name\":\"anonymous-cluster\"," +
					"\"bootstrap_servers\":\"http://test\",\"api_keys\":{\"abc-key-123\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"def-secret-456\"}}," +
					"\"api_key\":\"abc-key-123\"}},\"kafka_cluster\":\"anonymous-id\",\"schema_registry_clusters\":{}}},\"context_states\":{\"my-context\":{" +
					"\"auth\":null,\"auth_token\":\"\"}},\"current_context\":\"my-context\"}",
			},
			want: &Config{
				CLIName: "confluent",
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
				Version:        ver,
				Logger:         log.New(),
			},
			file: "/tmp/TestConfig_Load.json",
		},
		{
			name: "succeed loading config with state from file",
			args: &args{
				contents: "{\"platforms\":{\"http://test\":{\"Name\":\"http://test\",\"server\":\"http://test\"}},\"credentials\":{\"username-test-user\":{\"Name\":\"username-test-user\",\"Username\":\"test-user\",\"Password\":\"\",\"APIKeyPair\":null,\"CredentialType\":0}},\"contexts\":{\"my-context\":{\"name\":\"my-context\",\"platform\":\"http://test\",\"credential\":\"username-test-user\",\"kafka_clusters\":{\"anonymous-id\":{\"id\":\"anonymous-id\",\"name\":\"anonymous-cluster\",\"bootstrap_servers\":\"http://test\",\"api_keys\":{\"abc-key-123\":{\"api_key\":\"abc-key-123\",\"api_secret\":\"\"}},\"api_key\":\"abc-key-123\"}},\"kafka_cluster\":\"anonymous-id\",\"schema_registry_clusters\":{\"acc-123\":{\"id\":\"lsrc-123\",\"schema_registry_endpoint\":\"http://some-lsrc-endpoint\",\"schema_registry_credentials\":null}}}},\"context_states\":{\"my-context\":{\"auth\":{\"user\":{\"id\":123,\"email\":\"test-user@email\"},\"account\":{\"id\":\"acc-123\",\"name\":\"test-env\"},\"accounts\":[{\"id\":\"acc-123\",\"name\":\"test-env\"}]},\"auth_token\":\"abc123\"}},\"current_context\":\"my-context\"}",
			},
			want: &Config{
				CLIName: "confluent",
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
				Version: ver,
				Logger:  log.New(),
			},
			file: "/tmp/TestConfig_Load.json",
		},
	}
	for _, tt := range tests {
		/*
			CLIName              string                   `json:"-" hcl:"-"`
			MetricSink           metric.Sink              `json:"-" hcl:"-"`
			Logger               *log.Logger              `json:"-" hcl:"-"`
			Version              *pversion.Version        `json:"-" hcl:"-"`
			Filename             string                   `json:"-" hcl:"-"`
			Platforms            map[string]*Platform     `json:"platforms" hcl:"platforms"`
			Credentials          map[string]*Credential   `json:"credentials" hcl:"credentials"`
			Contexts             map[string]*Context      `json:"contexts" hcl:"contexts"`
			ContextStates        map[string]*ContextState `json:"context_states" hcl:"context_states"`
			CurrentContext       string                   `json:"current_context" hcl:"current_context"`
			UserSpecifiedContext string                   `json:"-" hcl:"-"`
		*/
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Logger = log.New()
			c.Filename = tt.file
			for _, context := range tt.want.Contexts {
				context.Config = tt.want
			}
			err := ioutil.WriteFile(tt.file, []byte(tt.args.contents), 0644)
			if err != nil {
				t.Errorf("unable to test config to file: %+v", err)
			}
			if err := c.Load("", "", "", "", client); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %+v, wantErr %+v", err, tt.wantErr)
			}
			c.Filename = "" // only for testing
			fmt.Println(tt.args.contents)
			if !t.Failed() && !reflect.DeepEqual(c, tt.want) {
				//t.Errorf("Config.Load() = %+v, want %+v", c, tt.want)
				t.Errorf("Config.Load() = %+v, \nwant %+v", c.Context(),
					tt.want.Context())
			}
			os.Remove(tt.file)
		})
	}
}

//func TestConfig_Save(t *testing.T) {
//	type args struct {
//		url   string
//		token string
//	}
//	tests := []struct {
//		name    string
//		args    *args
//		want    string
//		wantErr bool
//		file    string
//	}{
//		{
//			name: "save auth token to file",
//			args: &args{
//				token: "abc123",
//			},
//			want: "\"auth_token\": \"abc123\"",
//			file: "/tmp/TestConfig_Save.json",
//		},
//		{
//			name: "save auth url to file",
//			args: &args{
//				url: "https://stag.cpdev.cloud",
//			},
//			want: "\"auth_url\": \"https://stag.cpdev.cloud\"",
//			file: "/tmp/TestConfig_Save.json",
//		},
//		{
//			name: "create parent config dirs",
//			args: &args{
//				token: "abc123",
//			},
//			want: "\"auth_token\": \"abc123\"",
//			file: "/tmp/xyz987/TestConfig_Save.json",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &Config{Filename: tt.file, AuthToken: tt.args.token, AuthURL: tt.args.url}
//			if err := c.Save(); (err != nil) != tt.wantErr {
//				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
//			}
//			got, _ := ioutil.ReadFile(tt.file)
//			if !strings.Contains(string(got), tt.want) {
//				t.Errorf("Config.Save() = %v, want contains %v", string(got), tt.want)
//			}
//			fd, _ := os.Stat(tt.file)
//			if fd.Mode() != 0600 {
//				t.Errorf("Config.Save() file should only be readable by user")
//			}
//			os.RemoveAll("/tmp/xyz987")
//		})
//	}
//}

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
			want: os.Getenv("HOME") + "/.ccloud/config.json",
		},
		{
			name: "config file for confluent binary",
			fields: fields{
				CLIName: "confluent",
			},
			want: os.Getenv("HOME") + "/.confluent/config.json",
		},
		{
			name:   "should default to ~/.confluent if CLIName isn't provided",
			fields: fields{},
			want:   os.Getenv("HOME") + "/.confluent/config.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(&Config{
				CLIName: tt.fields.CLIName,
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
	//platform := &Platform{
	//	Name:   "https://fake-server.com",
	//	Server: "https://fake-server.com",
	//}
	//credential := &Credential{
	//	Name: "api-key-lock",
	//	APIKeyPair: &APIKeyPair{
	//		Key:    "lock",
	//		Secret: "shhh",
	//	},
	//	CredentialType: APIKey,
	//}
	//contextName := "test-context"
	//state := &ContextState{
	//	AuthToken: "abc123",
	//}
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
		kafkaClusters          map[string]*KafkaClusterConfig
		kafka                  string
		schemaRegistryClusters map[string]*SchemaRegistryCluster
		state                  *ContextState
		client                 *sdk.Client
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
			client:                 context.Client,
			filename:               filename,
			want:                   conf,
			wantErr:                false,
		},
		{
			name:                   "add existing context",
			config:                 conf,
			contextName:            context.Name,
			platformName:           context.PlatformName,
			credentialName:         context.CredentialName,
			kafkaClusters:          context.KafkaClusters,
			kafka:                  context.Kafka,
			schemaRegistryClusters: context.SchemaRegistryClusters,
			state:                  context.State,
			client:                 context.Client,
			filename:               filename,
			want:                   nil,
			wantErr:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.AddContext(tt.contextName, tt.platformName, tt.credentialName, tt.kafkaClusters, tt.kafka,
				tt.schemaRegistryClusters, tt.state, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddContext() error = %v, wantErr %v", err, tt.wantErr)
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

func TestConfig_AuthenticatedState(t *testing.T) {
	type fields struct {
		CLIName        string
		MetricSink     metric.Sink
		Logger         *log.Logger
		Filename       string
		Platforms      map[string]*Platform
		Credentials    map[string]*Credential
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    *ContextState
	}{
		{
			name: "succeed checking authenticated state of user with auth token",
			fields: fields{
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: Username,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: &Credential{
						CredentialType: Username,
					},
					CredentialName: "current-cred",
					State: &ContextState{
						Auth: &AuthConfig{
							Account: &orgv1.Account{
								Id: "abc123",
							},
							Accounts: nil,
						},
						AuthToken: "nekot",
					},
				}},
				CurrentContext: "current-context",
			},
			wantErr: false,
			want: &ContextState{
				Auth: &AuthConfig{
					Account: &orgv1.Account{
						Id: "abc123",
					},
					Accounts: nil,
				},
				AuthToken: "nekot",
			},
		},
		{
			name: "error when authenticated state of user without auth token with username creds",
			fields: fields{
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: Username,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: &Credential{
						CredentialType: Username,
					},
					CredentialName: "current-cred",
				}},
				CurrentContext: "current-context",
			},
			wantErr: true,
		},
		{
			name: "error when checking authenticated state of user with API key creds",
			fields: fields{
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: APIKey,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: &Credential{
						CredentialType: APIKey,
					},
					CredentialName: "current-cred",
				}},
				CurrentContext: "current-context",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				CLIName:        tt.fields.CLIName,
				MetricSink:     tt.fields.MetricSink,
				Logger:         tt.fields.Logger,
				Filename:       tt.fields.Filename,
				Platforms:      tt.fields.Platforms,
				Credentials:    tt.fields.Credentials,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			got, err := c.AuthenticatedState()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthenticatedState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, c.Context().State) {
				t.Errorf("AuthenticatedState() got = %v, want %v", got, tt.want)
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
				contextName:     {Name: contextName,},
				"other-context": {Name: "other-context"},
			},
				CurrentContext: "other-context",
			},
			args:    args{name: contextName},
			wantErr: false,
			wantConfig: &Config{
				Contexts:       map[string]*Context{"other-context": {Name: "other-context",}},
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

func TestConfig_SchemaRegistryCluster(t *testing.T) {
	conf := AuthenticatedConfigMock()
	context := conf.Context()
	srCluster := context.SchemaRegistryClusters[context.State.Auth.Account.Id]
	noAuthConf := AuthenticatedConfigMock()
	noAuthConf.Context().State = new(ContextState)
	noAuthConf.ContextStates[noAuthConf.Context().Name] = new(ContextState)
	tests := []struct {
		name    string
		config  *Config
		want    *SchemaRegistryCluster
		wantErr bool
		err     error
	}{
		{
			name: "succeed getting existing schema registry cluster",
			config: conf,
			want: srCluster,
			wantErr: false,
		},
		{
			name: "error getting schema registry cluster without current context",
			config: New(),
			wantErr: true,
			err:     cerrors.ErrNoContext,
		},
		{
			name: "error getting schema registry cluster when not logged in",
			config: noAuthConf,
			wantErr: true,
			err:     cerrors.ErrNotLoggedIn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.SchemaRegistryCluster()
			if (err != nil) != tt.wantErr {
				t.Errorf("SchemaRegistryCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SchemaRegistryCluster() got = %v, want %v", got, tt.want)
			}
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
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
