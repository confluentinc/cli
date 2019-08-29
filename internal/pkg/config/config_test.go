package config

import (
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/metric"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestConfig_Load(t *testing.T) {
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
			file: "/tmp/TestConfig_Load.json",
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
			file: "/tmp/TestConfig_Load.json",
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
			if !reflect.DeepEqual(c, tt.want) {
				t.Errorf("Config.Load() = %v, want %v", c, tt.want)
			}
			os.Remove(tt.file)
		})
	}
}

func TestConfig_Save(t *testing.T) {
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
			file: "/tmp/TestConfig_Save.json",
		},
		{
			name: "save auth url to file",
			args: &args{
				url: "https://stag.cpdev.cloud",
			},
			want: "\"auth_url\": \"https://stag.cpdev.cloud\"",
			file: "/tmp/TestConfig_Save.json",
		},
		{
			name: "create parent config dirs",
			args: &args{
				token: "abc123",
			},
			want: "\"auth_token\": \"abc123\"",
			file: "/tmp/xyz987/TestConfig_Save.json",
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
			if fd.Mode() != 0600 {
				t.Errorf("Config.Save() file should only be readable by user")
			}
			os.RemoveAll("/tmp/xyz987")
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
	platform := &Platform{Server: "fake-server.com"}
	credential := &Credential{
		APIKeyPair: &APIKeyPair{
			Key: "lock",
		},
		CredentialType: APIKey,
	}
	contextName := "test-context"
	filename := "/tmp/TestConfig_AddContext.json"
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
		wantPanic bool
	}{
		{
			name: "API Key credential stringify",
			credential: &Credential{
				CredentialType: APIKey,
				APIKeyPair:     keyPair,
				Username:       username,
			},
			want: "api-key-lock",
			wantPanic: false,
		},
		{
			name: "username/password credential stringify",
			credential: &Credential{
				CredentialType: Username,
				APIKeyPair:     keyPair,
				Username:       username,
			},
			want: "username-me",
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
		name    string
		fields  fields
		want    CredentialType
		wantErr bool
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
