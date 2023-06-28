package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	authTokenPlaceholder        = "AUTH_TOKEN_PLACEHOLDER"
	authRefreshTokenPlaceholder = "AUTH_REFRESH_TOKEN_PLACEHOLDER"
	saltPlaceholder             = "SALT_TOKEN_PLACEHOLDER"
	noncePlaceholder            = "NONCE_TOKEN_PLACEHOLDER"
	apiSecretPlaceholder        = "API_SECRET_PLACEHOLDER"
	apiSaltPlaceholder          = "c2FsdHBsYWNlaG9sZGVy"
	apiNoncePlaceholder         = "bm9uY2VwbGFjZWhvbGRlcg=="
)

var (
	apiKeyString      = "abc-key-123"
	apiSecretString   = "def-secret-456"
	apiCredentialName = fmt.Sprintf("api-key-%s", apiKeyString)
	kafkaClusterID    = "anonymous-id"
	contextName       = "my-context"
	environmentId     = "acc-123"
	cloudPlatforms    = []string{
		"devel.cpdev.cloud",
		"stag.cpdev.cloud",
		"confluent.cloud",
		"premium-oryx.gcp.priv.cpdev.cloud",
	}
	regularOrgContextState = &ContextState{
		Auth: &AuthConfig{
			User: &ccloudv1.User{
				Id:    123,
				Email: "test-user@email",
			},
			Organization: testserver.RegularOrg,
		},
		AuthToken:        "eyJ.eyJ.abc",
		AuthRefreshToken: "v1.abc",
	}
	suspendedOrgContextState = func(eventType ccloudv1.SuspensionEventType) *ContextState {
		return &ContextState{Auth: &AuthConfig{Organization: testserver.SuspendedOrg(eventType)}}
	}
)

type TestInputs struct {
	kafkaClusters        map[string]*KafkaClusterConfig
	activeKafka          string
	statefulConfig       *Config
	statelessConfig      *Config
	twoEnvStatefulConfig *Config
	environment          string
}

func SetupTestInputs(isCloud bool) *TestInputs {
	testInputs := &TestInputs{}
	platform := &Platform{
		Name:   "http://test",
		Server: "http://test",
	}
	if isCloud {
		platform.Name = testserver.TestCloudUrl.String()
	}
	apiCredential := &Credential{
		Name: "api-key-abc-key-123",
		APIKeyPair: &APIKeyPair{
			Key:    apiKeyString,
			Secret: apiSecretString,
		},
		CredentialType: 1,
	}
	loginCredential := &Credential{
		Name:           "username-test-user",
		Username:       "test-user",
		CredentialType: 0,
	}
	savedCredentials := map[string]*LoginCredential{contextName: {
		IsCloud:           isCloud,
		Username:          "test-user",
		EncryptedPassword: "encrypted-password",
	}}
	testInputs.environment = environmentId
	testInputs.kafkaClusters = map[string]*KafkaClusterConfig{
		kafkaClusterID: {
			ID:        kafkaClusterID,
			Name:      "anonymous-cluster",
			Bootstrap: "http://test",
			APIKeys:   map[string]*APIKeyPair{apiKeyString: apiCredential.APIKeyPair},
			APIKey:    apiKeyString,
		},
	}
	testInputs.activeKafka = kafkaClusterID
	statefulContext := &Context{
		Name:               contextName,
		Platform:           platform,
		PlatformName:       platform.Name,
		Credential:         loginCredential,
		CredentialName:     loginCredential.Name,
		CurrentEnvironment: environmentId,
		Environments:       map[string]*EnvironmentContext{environmentId: {}},
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			environmentId: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
				SrCredentials:          nil,
			},
		},
		State: regularOrgContextState,
	}
	statelessContext := &Context{
		Name:                   contextName,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             apiCredential,
		CredentialName:         apiCredential.Name,
		Environments:           map[string]*EnvironmentContext{},
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{},
		State:                  &ContextState{},
		Config:                 &Config{SavedCredentials: savedCredentials},
	}
	twoEnvStatefulContext := &Context{
		Name:               contextName,
		Platform:           platform,
		PlatformName:       platform.Name,
		Credential:         loginCredential,
		CredentialName:     loginCredential.Name,
		CurrentEnvironment: "acc-123",
		Environments: map[string]*EnvironmentContext{
			"acc-123":  {},
			"env-flag": {},
		},
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			environmentId: {
				Id:                     "lsrc-123",
				SchemaRegistryEndpoint: "http://some-lsrc-endpoint",
			},
		},
		State: regularOrgContextState,
	}
	context := "onprem"
	if isCloud {
		context = "cloud"
	}
	testInputs.statefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Filename: fmt.Sprintf("test_json/stateful_%s.json", context),
			Ver:      config.Version{Version: ver},
		},
		Platforms: map[string]*Platform{platform.Name: platform},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		CurrentContext:   contextName,
		Contexts:         map[string]*Context{contextName: statefulContext},
		ContextStates:    map[string]*ContextState{contextName: regularOrgContextState},
		IsTest:           true,
		SavedCredentials: savedCredentials,
	}
	testInputs.statelessConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Filename: fmt.Sprintf("test_json/stateless_%s.json", context),
			Ver:      config.Version{Version: ver},
		},
		Platforms: map[string]*Platform{platform.Name: platform},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		Contexts:         map[string]*Context{contextName: statelessContext},
		ContextStates:    map[string]*ContextState{contextName: {}},
		CurrentContext:   contextName,
		IsTest:           true,
		SavedCredentials: savedCredentials,
	}
	testInputs.twoEnvStatefulConfig = &Config{
		BaseConfig: &config.BaseConfig{
			Filename: fmt.Sprintf("test_json/stateful_%s.json", context),
			Ver:      config.Version{Version: ver},
		},
		Platforms: map[string]*Platform{platform.Name: platform},
		Credentials: map[string]*Credential{
			apiCredential.Name:   apiCredential,
			loginCredential.Name: loginCredential,
		},
		CurrentContext: contextName,
		Contexts:       map[string]*Context{contextName: twoEnvStatefulContext},
		ContextStates:  map[string]*ContextState{contextName: regularOrgContextState},
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
					Filename: "test_json/load_disable_update.json",
					Ver:      config.Version{Version: ver},
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := New()
			cfg.Filename = test.file
			for _, context := range test.want.Contexts {
				context.Config = test.want
			}
			if err := cfg.Load(); (err != nil) != test.wantErr {
				t.Errorf("Config.Load() error = %+v, wantErr %+v", err, test.wantErr)
			}

			// Get around automatically assigned anonymous id and IsTest check
			test.want.AnonymousId = cfg.AnonymousId
			test.want.IsTest = cfg.IsTest
			test.want.Version = cfg.Version
			test.want.Credentials = cfg.Credentials
			if ctx := test.want.Contexts[contextName]; ctx != nil {
				ctx.Credential = cfg.Contexts[contextName].Credential
				ctx.KafkaClusterContext.KafkaClusterConfigs = cfg.Contexts[contextName].KafkaClusterContext.KafkaClusterConfigs
			}

			if !t.Failed() && !reflect.DeepEqual(cfg, test.want) {
				t.Errorf("Config.Load() =\n%+v, want \n%+v", cfg, test.want)
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configFile, _ := os.CreateTemp("", "TestConfig_Save.json")
			test.config.Filename = configFile.Name()
			ctx := test.config.Context()
			test.config.SavedCredentials = map[string]*LoginCredential{
				contextName: {
					IsCloud:           test.isCloud,
					Username:          "test-user",
					EncryptedPassword: "encrypted-password",
				},
			}
			if test.kafkaOverwrite != "" {
				test.config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
				ctx.KafkaClusterContext.SetActiveKafkaCluster(test.kafkaOverwrite)
			}
			if test.contextOverwrite != "" {
				test.config.SetOverwrittenCurrentContext(test.config.CurrentContext)
				test.config.CurrentContext = test.contextOverwrite
			}
			if err := test.config.Save(); (err != nil) != test.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, test.wantErr)
			}

			got, _ := os.ReadFile(configFile.Name())
			want, _ := os.ReadFile(test.wantFile)
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

func TestConfig_SaveWithEnvironmentOverwrite(t *testing.T) {
	configFile, err := os.CreateTemp("", "TestConfig_Save.json")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	testConfigsCloud := SetupTestInputs(true)
	config := testConfigsCloud.twoEnvStatefulConfig
	config.Filename = configFile.Name()
	config.SavedCredentials = map[string]*LoginCredential{contextName: {
		IsCloud:           true,
		Username:          "test-user",
		EncryptedPassword: "encrypted-password",
	}}
	config.SetOverwrittenCurrentEnvironment(config.Context().GetCurrentEnvironment())
	config.Context().SetCurrentEnvironment("env-flag")
	err = config.Save()
	require.NoError(t, err)

	got, _ := os.ReadFile(configFile.Name())
	want, _ := os.ReadFile("test_json/account_overwrite.json")
	wantString := replacePlaceholdersInWant(t, got, want)
	require.Equal(t, utils.NormalizeNewLines(wantString), utils.NormalizeNewLines(string(got)))

	fd, err := os.Stat(configFile.Name())
	require.NoError(t, err)
	if runtime.GOOS != "windows" && fd.Mode() != 0600 {
		t.Errorf("Config.Save() file should only be readable by user")
	}
}

func replacePlaceholdersInWant(t *testing.T, got []byte, want []byte) string {
	data := &Config{}
	err := json.Unmarshal(got, data)
	require.NoError(t, err)

	wantString := strings.ReplaceAll(string(want), authTokenPlaceholder, data.ContextStates[contextName].AuthToken)
	wantString = strings.ReplaceAll(wantString, authRefreshTokenPlaceholder, data.ContextStates[contextName].AuthRefreshToken)

	saltString := base64.RawStdEncoding.EncodeToString(data.ContextStates[contextName].Salt)
	wantString = strings.ReplaceAll(wantString, saltPlaceholder, saltString)

	nonceString := base64.RawStdEncoding.EncodeToString(data.ContextStates[contextName].Nonce)
	wantString = strings.ReplaceAll(wantString, noncePlaceholder, nonceString)

	wantString = strings.ReplaceAll(wantString, apiSecretPlaceholder, data.Credentials[apiCredentialName].APIKeyPair.Secret)

	apiSaltString := base64.RawStdEncoding.EncodeToString(data.Credentials[apiCredentialName].APIKeyPair.Salt)
	wantString = strings.ReplaceAll(wantString, apiSaltPlaceholder, apiSaltString)

	apiNonceString := base64.RawStdEncoding.EncodeToString(data.Credentials[apiCredentialName].APIKeyPair.Nonce)
	wantString = strings.ReplaceAll(wantString, apiNoncePlaceholder, apiNonceString)

	return wantString
}

func TestConfig_OverwrittenKafka(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string // simulates initial config value overwritten by a cluster flag value
		activeKafka    string // simulates the cluster flag value
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
	for _, test := range tests {
		ctx := test.config.Context()
		test.config.SetOverwrittenCurrentKafkaCluster(test.overwrittenVal)
		// resolve should reset the active kafka to be the overwritten value and return the flag value to be used in restore
		tempKafka := test.config.resolveOverwrittenKafka()
		require.Equal(t, test.activeKafka, tempKafka)
		if ctx.KafkaClusterContext.EnvContext && ctx.KafkaClusterContext.GetCurrentKafkaEnvContext() != nil {
			require.Equal(t, test.overwrittenVal, ctx.KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster)
		} else {
			require.Equal(t, test.overwrittenVal, ctx.KafkaClusterContext.ActiveKafkaCluster)
		}
		// restore should reset the active kafka to be the flag value
		test.config.restoreOverwrittenKafka(tempKafka)
		if ctx.KafkaClusterContext.EnvContext && ctx.KafkaClusterContext.GetCurrentKafkaEnvContext() != nil {
			require.Equal(t, tempKafka, ctx.KafkaClusterContext.GetCurrentKafkaEnvContext().ActiveKafkaCluster)
		} else {
			require.Equal(t, tempKafka, ctx.KafkaClusterContext.ActiveKafkaCluster)
		}
		test.config.overwrittenCurrentKafkaCluster = ""
	}
}

func TestConfig_OverwrittenContext(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name           string
		config         *Config
		overwrittenVal string // simulates initial context value overwritten by a context flag value
		currContext    string // simulates the context flag value
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
	for _, test := range tests {
		test.config.SetOverwrittenCurrentContext(test.overwrittenVal)
		// resolve should reset the current context to be the overwritten value and return the flag value to be used in restore
		tempContext := test.config.resolveOverwrittenContext()
		require.Equal(t, test.overwrittenVal, test.config.CurrentContext)
		require.Equal(t, test.currContext, tempContext)
		// restore should reset the current context to be the flag value
		test.config.restoreOverwrittenContext(tempContext)
		require.Equal(t, test.currContext, test.config.CurrentContext)
		test.config.overwrittenCurrentContext = ""
	}
}

func TestConfig_OverwrittenEnvironment(t *testing.T) {
	testConfigsCloud := SetupTestInputs(true)

	tests := []struct {
		name                          string
		config                        *Config
		currentEnvironment            string // simulates the environment flag value
		overwrittenCurrentEnvironment string // simulates initial environment value overwritten by a environment flag
	}{
		{
			name:               "test no overwrite value",
			config:             testConfigsCloud.statefulConfig,
			currentEnvironment: testConfigsCloud.statefulConfig.Context().GetCurrentEnvironment(),
		},
		{
			name:                          "test with overwrite value",
			config:                        testConfigsCloud.statefulConfig,
			currentEnvironment:            testConfigsCloud.statefulConfig.Context().GetCurrentEnvironment(),
			overwrittenCurrentEnvironment: "env-test",
		},
		{
			name:   "test no overwrite value",
			config: testConfigsCloud.statelessConfig,
		},
	}
	for _, test := range tests {
		test.config.SetOverwrittenCurrentEnvironment(test.overwrittenCurrentEnvironment)
		if test.config.Context().CurrentEnvironment == "" {
			tempAccount := test.config.resolveOverwrittenCurrentEnvironment()
			require.Empty(t, tempAccount)
			test.config.restoreOverwrittenEnvironment(tempAccount)
			require.Empty(t, test.config.Context().CurrentEnvironment)
		} else {
			// resolve should reset the current context to be the overwritten value and return the flag value to be used in restore
			tempAccount := test.config.resolveOverwrittenCurrentEnvironment()
			if test.overwrittenCurrentEnvironment != "" {
				require.Equal(t, test.overwrittenCurrentEnvironment, test.config.Context().GetCurrentEnvironment())
				require.Equal(t, test.currentEnvironment, tempAccount)
			}
			// restore should reset the current context to be the flag value
			test.config.restoreOverwrittenEnvironment(tempAccount)
			require.Equal(t, test.currentEnvironment, test.config.Context().GetCurrentEnvironment())
		}
		test.config.overwrittenCurrentEnvironment = ""
	}
}

func TestConfig_getFilename(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	path := filepath.Join(home, ".confluent", "config.json")
	require.Equal(t, path, New().GetFilename())
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
		currentEnvironment     string
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
		currentEnvironment:     context.CurrentEnvironment,
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.AddContext(test.contextName, test.platformName, test.credentialName, test.kafkaClusters, test.kafka, test.schemaRegistryClusters, test.state, MockOrgResourceId, test.currentEnvironment)
			if (err != nil) != test.wantErr {
				t.Errorf("AddContext() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.want != nil {
				test.want.AnonymousId = test.config.AnonymousId
			}
			if !test.wantErr && !reflect.DeepEqual(test.want, test.config) {
				t.Errorf("AddContext() got = %v, want %v", test.config, test.want)
			}
		})
	}
	os.Remove(filename)
}

func TestConfig_CreateContext(t *testing.T) {
	cfg := &Config{
		BaseConfig:    &config.BaseConfig{Ver: config.Version{Version: version.Must(version.NewVersion("1.0.0"))}},
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
			name:   "succeed setting valid context",
			fields: fields{Config: cfg},
			args:   args{name: contextName},
		},
		{
			name:    "fail setting nonexistent context",
			fields:  fields{Config: cfg},
			args:    args{name: "some-context"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := test.fields.Config
			if err := cfg.UseContext(test.args.name); (err != nil) != test.wantErr {
				t.Errorf("UseContext() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr {
				assert.Equal(t, test.args.name, cfg.CurrentContext)
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
		{
			name:   "success finding existing context",
			fields: fields{Contexts: map[string]*Context{"test-context": {Name: "test-context"}}},
			args:   args{name: "test-context"},
			want:   &Context{Name: "test-context"},
		},
		{
			name:    "error finding nonexistent context",
			fields:  fields{Contexts: map[string]*Context{}},
			args:    args{name: "test-context"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &Config{Contexts: test.fields.Contexts}
			got, err := cfg.FindContext(test.args.name)
			if (err != nil) != test.wantErr {
				t.Errorf("FindContext() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("FindContext() got = %v, want %v", got, test.want)
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
				Contexts:       map[string]*Context{"test-context": {Name: "test-context"}},
				CurrentContext: "test-context",
			},
			want: &Context{Name: "test-context"},
		},
		{
			name:   "error getting current context when not set",
			fields: fields{Contexts: map[string]*Context{}},
			want:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &Config{
				Contexts:       test.fields.Contexts,
				CurrentContext: test.fields.CurrentContext,
			}
			got := cfg.Context()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Context() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestKafkaClusterContext_SetAndGetActiveKafkaCluster_Env(t *testing.T) {
	testInputs := SetupTestInputs(true)
	ctx := testInputs.statefulConfig.Context()
	// temp file so json files in test_json do not get overwritten
	configFile, _ := os.CreateTemp("", "TestConfig_Save.json")
	ctx.Config.Filename = configFile.Name()

	// Creating another environment with another kafka cluster
	otherEnvironmentId := "other-abc"
	ctx.Environments[otherEnvironmentId] = &EnvironmentContext{}

	otherKafkaCluster := &KafkaClusterConfig{
		ID:        "other-kafka",
		Name:      "lit",
		Bootstrap: "http://test",
		APIKeys: map[string]*APIKeyPair{"akey": {
			Key:    "akey",
			Secret: "asecret",
		}},
		APIKey: "akey",
	}

	activeKafka := ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	require.Equal(t, testInputs.activeKafka, activeKafka)
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)

	// switch environment add the kafka cluster, and set it as active cluster
	ctx.SetCurrentEnvironment(otherEnvironmentId)
	ctx.KafkaClusterContext.AddKafkaClusterConfig(otherKafkaCluster)
	ctx.KafkaClusterContext.SetActiveKafkaCluster(otherKafkaCluster.ID)
	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	require.Equal(t, otherKafkaCluster.ID, activeKafka)
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)

	// switch environment back
	ctx.SetCurrentEnvironment(testInputs.environment)
	activeKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	require.Equal(t, testInputs.activeKafka, activeKafka)
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)
	_ = os.Remove(configFile.Name())
}

func TestKafkaClusterContext_SetAndGetActiveKafkaCluster_NonEnv(t *testing.T) {
	testInputs := SetupTestInputs(false)
	ctx := testInputs.statefulConfig.Context()
	// temp file so json files in test_json do not get overwritten
	configFile, _ := os.CreateTemp("", "TestConfig_Save.json")
	ctx.Config.Filename = configFile.Name()
	otherKafkaClusterId := "other-kafka"
	otherKafkaCluster := &KafkaClusterConfig{
		ID:        otherKafkaClusterId,
		Name:      "lit",
		Bootstrap: "http://test",
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
		t.Errorf("After setting active kafka, GetActiveKafkaClusterId() got %s, want %s.", activeKafka, testInputs.activeKafka)
	}
	require.Equal(t, ctx.KafkaClusterContext.GetActiveKafkaClusterConfig().ID, activeKafka)
	_ = os.Remove(configFile.Name())
}

func TestKafkaClusterContext_AddAndGetKafkaClusterConfig(t *testing.T) {
	clusterID := "lkc-abcdefg"

	kcc := &KafkaClusterConfig{
		ID:        clusterID,
		Name:      "lit",
		Bootstrap: "http://test",
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
		ID:        clusterID,
		Name:      "lit",
		Bootstrap: "http://test",
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
		ID:        clusterID,
		Name:      "lit",
		Bootstrap: "http://test",
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
	for _, platform := range cloudPlatforms {
		// test case: org not suspended
		cfg := &Config{
			Contexts: map[string]*Context{"context": {
				State:        regularOrgContextState,
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		require.True(t, cfg.IsCloudLogin(), platform+" should be true")
	}
}

func TestConfig_IsCloud_False(t *testing.T) {
	// test case: platform name not cloud
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

	for _, platform := range cloudPlatforms {
		// test case: org suspended due to normal reason
		cfg := &Config{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		require.False(t, cfg.IsCloudLogin(), platform+" should be false")

		// test case: org suspended due to end of free trial
		cfg = &Config{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		require.False(t, cfg.IsCloudLogin(), platform+" should be false")
	}
}

func TestConfig_IsCloudLoginAllowFreeTrialEnded_True(t *testing.T) {
	for _, platform := range cloudPlatforms {
		// test case: org not suspended
		cfg := &Config{
			Contexts: map[string]*Context{"context": {
				State:        regularOrgContextState,
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		isCloudLoginAllowFreeTrialEnded := cfg.CheckIsCloudLoginAllowFreeTrialEnded()
		require.True(t, isCloudLoginAllowFreeTrialEnded == nil, platform+" should be true")

		// test case: org suspended due to end of free trial
		cfg = &Config{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		isCloudLoginAllowFreeTrialEnded = cfg.CheckIsCloudLoginAllowFreeTrialEnded()
		require.True(t, isCloudLoginAllowFreeTrialEnded == nil, platform+" should be true")
	}
}

func TestConfig_IsCloudLoginAllowFreeTrialEnded_False(t *testing.T) {
	// test case: platform name not cloud
	configs := []*Config{
		nil,
		{
			Contexts:       map[string]*Context{"context": {PlatformName: "https://example.com"}},
			CurrentContext: "context",
		},
	}

	for _, cfg := range configs {
		isCloudLoginAllowFreeTrialEnded := cfg.CheckIsCloudLoginAllowFreeTrialEnded()
		require.True(t, isCloudLoginAllowFreeTrialEnded != nil)
	}

	for _, platform := range cloudPlatforms {
		// test case: org suspended due to normal reason
		cfg := &Config{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
				PlatformName: platform,
			}},
			CurrentContext: "context",
		}
		isCloudLoginAllowFreeTrialEnded := cfg.CheckIsCloudLoginAllowFreeTrialEnded()
		require.True(t, isCloudLoginAllowFreeTrialEnded != nil, platform+" should be false")
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
			Contexts: map[string]*Context{"context": {
				State:        regularOrgContextState,
				PlatformName: "confluent.cloud",
			}},
			CurrentContext: "context",
		},
		{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
				PlatformName: "confluent.cloud",
			}},
			CurrentContext: "context",
		},
		{
			Contexts: map[string]*Context{"context": {
				State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
				PlatformName: "confluent.cloud",
			}},
			CurrentContext: "context",
		},
	}

	for _, cfg := range configs {
		require.False(t, cfg.IsOnPremLogin())
	}
}
