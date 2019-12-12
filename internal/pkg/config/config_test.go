package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	cerrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {
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
			name: "should load auth token from file",
			args: &args{
				contents: "{\"auth_token\": \"abc123\"}",
			},
			want: &Config{
				CLIName:     "confluent",
				AuthToken:   "abc123",
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{},
			},
			file: testConfigFile.Name(),
		},
		{
			name: "should load auth url from file",
			args: &args{
				contents: "{\"auth_url\": \"https://stag.cpdev.cloud\"}",
			},
			want: &Config{
				CLIName:     "confluent",
				AuthURL:     "https://stag.cpdev.cloud",
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{},
			},
			file: testConfigFile.Name(),
		},
		{
			name: "should load disable update checks and disable updates",
			args: &args{
				contents: "{\"disable_update_check\": true, \"disable_updates\": true}",
			},
			want: &Config{
				CLIName:            "confluent",
				DisableUpdates:     true,
				DisableUpdateCheck: true,
				Platforms:          map[string]*Platform{},
				Credentials:        map[string]*Credential{},
				Contexts:           map[string]*Context{},
			},
			file: testConfigFile.Name(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Filename = tt.file
			err := ioutil.WriteFile(tt.file, []byte(tt.args.contents), 0644)
			if err != nil {
				t.Errorf("unable to test config to file: %v", err)
			}
			if err := c.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			c.Filename = "" // only for testing
			// get around automatically assigned anonymous id
			tt.want.AnonymousId = c.AnonymousId
			if !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %v, want %v", c, tt.want)
			}
			os.Remove(tt.file)
		})
	}
}

func TestConfig_Save(t *testing.T) {
	testConfigFile, _ := ioutil.TempFile("", "TestConfig_Save.json")
	type args struct {
		url   string
		token string
	}
	tests := []struct {
		name    string
		args    *args
		want    string
		wantErr bool
		file    string
	}{
		{
			name: "save auth token to file",
			args: &args{
				token: "abc123",
			},
			want: "\"auth_token\": \"abc123\"",
			file: testConfigFile.Name(),
		},
		{
			name: "save auth url to file",
			args: &args{
				url: "https://stag.cpdev.cloud",
			},
			want: "\"auth_url\": \"https://stag.cpdev.cloud\"",
			file: testConfigFile.Name(),
		},
		{
			name: "create parent config dirs",
			args: &args{
				token: "abc123",
			},
			want: "\"auth_token\": \"abc123\"",
			file: testConfigFile.Name(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Filename: tt.file, AuthToken: tt.args.token, AuthURL: tt.args.url}
			if err := c.Save(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := ioutil.ReadFile(tt.file)
			if !strings.Contains(string(got), tt.want) {
				t.Errorf("Config.Save() = %v, want contains %v", string(got), tt.want)
			}
			fd, _ := os.Stat(tt.file)
			if runtime.GOOS != "windows" && fd.Mode() != 0600 {
				t.Errorf("Config.Save() file should only be readable by user")
			}
			os.Remove(testConfigFile.Name())
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
	platform := &Platform{Server: "https://fake-server.com"}
	credential := &Credential{
		APIKeyPair: &APIKeyPair{
			Key: "lock",
		},
		CredentialType: APIKey,
	}
	contextName := "test-context"
	tempContextFile, _ := ioutil.TempFile("", "TestConfig_AddContext.json")
	filename := tempContextFile.Name()
	tests := []struct {
		name                   string
		config                 *Config
		contextName            string
		platform               *Platform
		credential             *Credential
		kafkaClusters          map[string]*KafkaClusterConfig
		kafka                  string
		schemaRegistryClusters map[string]*SchemaRegistryCluster
		filename               string
		want                   *Config
		wantErr                bool
	}{
		{
			name: "add valid context",
			config: &Config{
				Filename:    filename,
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{},
			},
			contextName:            contextName,
			platform:               platform,
			credential:             credential,
			kafkaClusters:          map[string]*KafkaClusterConfig{},
			kafka:                  "akfak",
			schemaRegistryClusters: map[string]*SchemaRegistryCluster{},
			filename:               filename,
			want: &Config{
				Filename:    filename,
				Platforms:   map[string]*Platform{platform.String(): platform},
				Credentials: map[string]*Credential{credential.String(): credential},
				Contexts: map[string]*Context{contextName: {
					Name:                   contextName,
					Platform:               platform.String(),
					Credential:             credential.String(),
					KafkaClusters:          map[string]*KafkaClusterConfig{},
					Kafka:                  "akfak",
					SchemaRegistryClusters: map[string]*SchemaRegistryCluster{},
				}},
				CurrentContext: "",
			},
			wantErr: false,
		},
		{
			name: "add existing context",
			config: &Config{
				Filename:    filename,
				Platforms:   map[string]*Platform{},
				Credentials: map[string]*Credential{},
				Contexts:    map[string]*Context{contextName: {}},
			},
			contextName:            contextName,
			platform:               platform,
			credential:             credential,
			kafkaClusters:          map[string]*KafkaClusterConfig{},
			kafka:                  "akfak",
			schemaRegistryClusters: map[string]*SchemaRegistryCluster{},
			filename:               filename,
			wantErr:                true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.AddContext(tt.contextName, tt.platform, tt.credential, tt.kafkaClusters, tt.kafka, tt.schemaRegistryClusters)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, tt.config)
			err = tt.config.Load()
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, tt.config)
		})
	}
	os.Remove(filename)
}

func TestCredential_String(t *testing.T) {
	keyPair := &APIKeyPair{
		Key:    "lock",
		Secret: "victoria",
	}
	username := "me"
	tests := []struct {
		name       string
		credential *Credential
		want       string
		wantPanic  bool
	}{
		{
			name: "API Key credential stringify",
			credential: &Credential{
				CredentialType: APIKey,
				APIKeyPair:     keyPair,
				Username:       username,
			},
			want:      "api-key-lock",
			wantPanic: false,
		},
		{
			name: "username/password credential stringify",
			credential: &Credential{
				CredentialType: Username,
				APIKeyPair:     keyPair,
				Username:       username,
			},
			want:      "username-me",
			wantPanic: false,
		},
		{
			name: "invalid credential stringify",
			credential: &Credential{
				CredentialType: -1,
				APIKeyPair:     keyPair,
				Username:       username,
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				panicFunc := func() {
					_ = tt.credential.String()
				}
				assert.Panics(t, panicFunc)
			} else {
				assert.Equal(t, tt.want, tt.credential.String())
			}
		})
	}
}

func TestPlatform_String(t *testing.T) {
	platform := &Platform{Server: "alfred"}
	assert.Equal(t, platform.Server, platform.String())
}

func TestConfig_SetContext(t *testing.T) {
	type fields struct {
		CLIName        string
		MetricSink     metric.Sink
		Logger         *log.Logger
		Filename       string
		AuthURL        string
		AuthToken      string
		Auth           *AuthConfig
		Platforms      map[string]*Platform
		Credentials    map[string]*Credential
		Contexts       map[string]*Context
		CurrentContext string
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
				Contexts: map[string]*Context{"some-context": {}},
			},
			args:    args{name: "some-context"},
			wantErr: false,
		},
		{
			name: "fail setting nonexistent context",
			fields: fields{
				Contexts: map[string]*Context{},
			},
			args:    args{name: "some-context"},
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
				AuthURL:        tt.fields.AuthURL,
				AuthToken:      tt.fields.AuthToken,
				Auth:           tt.fields.Auth,
				Platforms:      tt.fields.Platforms,
				Credentials:    tt.fields.Credentials,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			if err := c.SetContext(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("SetContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.args.name, c.CurrentContext)
			}
		})
	}
}

func TestConfig_CredentialType(t *testing.T) {
	type fields struct {
		CLIName        string
		MetricSink     metric.Sink
		Logger         *log.Logger
		Filename       string
		AuthURL        string
		AuthToken      string
		Auth           *AuthConfig
		Platforms      map[string]*Platform
		Credentials    map[string]*Credential
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name     string
		fields   fields
		want     CredentialType
		wantErr  bool
		wantExit bool
	}{
		{
			name: "succeed getting CredentialType from existing credential",
			fields: fields{
				Credentials: map[string]*Credential{"some-cred": {
					CredentialType: APIKey,
				}},
				Contexts: map[string]*Context{"textcon": {
					Credential: "some-cred",
				}},
				CurrentContext: "textcon",
			},
			want:    APIKey,
			wantErr: false,
		},
		{
			name: "fail getting CredentialType from nonexistent credential",
			fields: fields{
				Credentials: map[string]*Credential{"some-cred": {
					CredentialType: APIKey,
				}},
				Contexts: map[string]*Context{"textcon": {
					Credential: "another-cred",
				}},
				CurrentContext: "textcon",
			},
			wantErr: true,
		},
		{
			name: "fail getting CredentialType from credential with no current context",
			fields: fields{
				Credentials:    map[string]*Credential{},
				CurrentContext: "",
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
				AuthURL:        tt.fields.AuthURL,
				AuthToken:      tt.fields.AuthToken,
				Auth:           tt.fields.Auth,
				Platforms:      tt.fields.Platforms,
				Credentials:    tt.fields.Credentials,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			got, err := c.CredentialType()
			if (err != nil) != tt.wantErr {
				t.Errorf("CredentialType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CredentialType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_CheckLogin(t *testing.T) {
	type fields struct {
		CLIName        string
		MetricSink     metric.Sink
		Logger         *log.Logger
		Filename       string
		AuthURL        string
		AuthToken      string
		Auth           *AuthConfig
		Platforms      map[string]*Platform
		Credentials    map[string]*Credential
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "succeed checking login of user with auth token",
			fields: fields{
				AuthToken: "nekot",
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: Username,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: "current-cred",
				}},
				CurrentContext: "current-context",
			},
		},
		{
			name: "error when checking login of user without auth token with username creds",
			fields: fields{
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: Username,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: "current-cred",
				}},
				CurrentContext: "current-context",
			},
			wantErr: true,
		},
		{
			name: "error when checking login of user with API key creds",
			fields: fields{
				Credentials: map[string]*Credential{"current-cred": {
					CredentialType: APIKey,
				}},
				Contexts: map[string]*Context{"current-context": {
					Credential: "current-cred",
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
				AuthURL:        tt.fields.AuthURL,
				AuthToken:      tt.fields.AuthToken,
				Auth:           tt.fields.Auth,
				Platforms:      tt.fields.Platforms,
				Credentials:    tt.fields.Credentials,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			if err := c.CheckLogin(); (err != nil) != tt.wantErr {
				t.Errorf("CheckLogin() error = %v, wantErr %v", err, tt.wantErr)
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

func TestConfig_KafkaClusterConfig(t *testing.T) {
	type fields struct {
		Contexts       map[string]*Context
		CurrentContext string
		Filename       string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *KafkaClusterConfig
		wantErr bool
		err     error
	}{
		{
			name: "succeed getting Kafka cluster config",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					KafkaClusters: map[string]*KafkaClusterConfig{"k-id": {
						ID:   "k-id",
						Name: "k-cluster",
					}},
					Kafka: "k-id",
				}},
				CurrentContext: "test-context",
			},
			want: &KafkaClusterConfig{
				ID:   "k-id",
				Name: "k-cluster",
			},
			wantErr: false,
		},
		{
			name: "error getting Kafka cluster config with no current context",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					KafkaClusters: map[string]*KafkaClusterConfig{"k-id": {
						ID:   "k-id",
						Name: "k-cluster",
					}},
					Kafka: "",
				}},
				CurrentContext: "",
			},
			wantErr: true,
		},
		{
			name: "succeed getting Kafka cluster config when it is not set",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					KafkaClusters: map[string]*KafkaClusterConfig{"k-id": {
						ID:   "k-id",
						Name: "k-cluster",
					}},
					Kafka: "",
				}},
				CurrentContext: "test-context",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "error getting set but nonexistent Kafka cluster config",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name:  "test-context",
					Kafka: "nonexistent-cluster",
				}},
				Filename:       "/tmp/TestConfig_KafkaClusterConfig.json",
				CurrentContext: "test-context",
			},
			wantErr: true,
			err: errors.New("the configuration of context \"test-context\" has been corrupted. " +
				"To fix, please remove the config file located at /tmp/TestConfig_KafkaClusterConfig.json," +
				" and run `login` or `init`"),
		},
		{
			name: "error getting set but nonexistent Kafka cluster config in config with bad filepath",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name:  "test-context",
					Kafka: "nonexistent-cluster",
				}},
				Filename:       "~badfilepath",
				CurrentContext: "test-context",
			},
			wantErr: true,
			err: errors.New("an error resolving the config filepath at ~badfilepath has occurred. " +
				"Please try moving the file to a different location"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
				Filename:       tt.fields.Filename,
			}
			got, err := c.KafkaClusterConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("KafkaClusterConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KafkaClusterConfig() got = %v, want %v", got, tt.want)
			}
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestConfig_CheckHasAPIKey(t *testing.T) {
	type fields struct {
		Contexts       map[string]*Context
		CurrentContext string
	}
	type args struct {
		clusterID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		err     interface{}
	}{
		{
			name: "succeed checking existing active API key",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					KafkaClusters: map[string]*KafkaClusterConfig{"k-id": {
						ID:   "k-id",
						Name: "k-cluster",
						APIKeys: map[string]*APIKeyPair{"yek": {
							Key:    "yek",
							Secret: "shhh",
						}},
						APIKey: "yek",
					}},
				}},
				CurrentContext: "test-context",
			},
			args:    args{clusterID: "k-id"},
			wantErr: false,
		},
		{
			name: "error checking API key with no active key",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					KafkaClusters: map[string]*KafkaClusterConfig{"k-id": {
						ID:   "k-id",
						Name: "k-cluster",
						APIKeys: map[string]*APIKeyPair{"yek": {
							Key:    "yek",
							Secret: "shhh",
						}},
						APIKey: "",
					}},
				}},
				CurrentContext: "test-context",
			},
			args:    args{clusterID: "k-id"},
			wantErr: true,
			err:     &cerrors.UnspecifiedAPIKeyError{ClusterID: "k-id"},
		},
		{
			name: "error checking API key with no active context",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
				}},
				CurrentContext: "",
			},
			args:    args{clusterID: "k-id"},
			wantErr: true,
			err:     cerrors.ErrNoContext,
		},
		{
			name: "error checking API key with no matching cluster",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
				}},
				CurrentContext: "test-context",
			},
			args:    args{clusterID: "k-id"},
			wantErr: true,
			err:     errors.New("unknown kafka cluster: k-id"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			err := c.CheckHasAPIKey(tt.args.clusterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHasAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestConfig_CheckSchemaRegistryHasAPIKey(t *testing.T) {
	type fields struct {
		Auth           *AuthConfig
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Check for valid credentials in config",
			fields: fields{
				CurrentContext: "ctx",
				Auth:           &AuthConfig{Account: &orgv1.Account{Id: "me"}, User: new(orgv1.User)},
				Contexts: map[string]*Context{"ctx": {
					SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
						"me": {
							SrCredentials: &APIKeyPair{
								Key:    "Abra",
								Secret: "cadabra",
							},
						},
					},
				},
				}},
			want: true,
		},
		{
			name: "Check for empty Schema Registry API Key credentials",
			fields: fields{
				CurrentContext: "ctx",
				Auth:           &AuthConfig{Account: &orgv1.Account{Id: "me"}},
				Contexts: map[string]*Context{"ctx": {
					SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
						"me": {
							SrCredentials: nil,
						},
					},
				},
				}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				Auth:           tt.fields.Auth,
				CurrentContext: tt.fields.CurrentContext,
			}
			returnVal := c.CheckSchemaRegistryHasAPIKey()
			if returnVal != tt.want {
				t.Errorf("CheckSchemaRegistryHasAPIKey() %s returnVal = %v, wantedReturnVal %v", tt.name, returnVal, tt.want)
			}
		})
	}
}

func TestConfig_SchemaRegistryCluster(t *testing.T) {
	type fields struct {
		Auth           *AuthConfig
		Contexts       map[string]*Context
		CurrentContext string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *SchemaRegistryCluster
		wantErr bool
		err     error
	}{
		{
			name: "succeed getting existing schema registry cluster",
			fields: fields{
				Auth: &AuthConfig{
					Account: &orgv1.Account{
						Id: "test-acct-id",
					},
				},
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					SchemaRegistryClusters: map[string]*SchemaRegistryCluster{"test-acct-id": {
						SchemaRegistryEndpoint: "test-sr",
					},
					},
				}},
				CurrentContext: "test-context",
			},
			want: &SchemaRegistryCluster{
				SchemaRegistryEndpoint: "test-sr",
			},
			wantErr: false,
		},
		{
			name: "succeed getting nonexistent schema registry cluster without current context",
			fields: fields{
				Auth: &AuthConfig{
					Account: &orgv1.Account{
						Id: "test-acct-id",
					},
				},
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
					SchemaRegistryClusters: map[string]*SchemaRegistryCluster{"another-acct": {
						SchemaRegistryEndpoint: "test-sr",
					},
					},
				}},
				CurrentContext: "test-context",
			},
			want:    &SchemaRegistryCluster{},
			wantErr: false,
		},
		{
			name: "error getting schema registry cluster without current context",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
				}},
				CurrentContext: "",
			},
			wantErr: true,
			err:     cerrors.ErrNoContext,
		},
		{
			name: "error getting schema registry cluster when not logged in",
			fields: fields{
				Contexts: map[string]*Context{"test-context": {
					Name: "test-context",
				}},
				CurrentContext: "test-context",
			},
			wantErr: true,
			err:     cerrors.ErrNotLoggedIn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Auth:           tt.fields.Auth,
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			got, err := c.SchemaRegistryCluster()
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
		name    string
		fields  fields
		want    *Context
		wantErr bool
		err     error
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
			wantErr: false,
		},
		{
			name: "error getting current context when not set",
			fields: fields{
				Contexts: map[string]*Context{},
			},
			wantErr: true,
			err:     cerrors.ErrNoContext,
		},
		{
			name: "error getting current context with corrupted config",
			fields: fields{
				Contexts:       map[string]*Context{},
				CurrentContext: "test-context",
			},
			wantErr: true,
			err:     errors.New("context \"test-context\" does not exist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Contexts:       tt.fields.Contexts,
				CurrentContext: tt.fields.CurrentContext,
			}
			got, err := c.Context()
			if (err != nil) != tt.wantErr {
				t.Errorf("Context() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Context() got = %v, want %v", got, tt.want)
			}
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}
