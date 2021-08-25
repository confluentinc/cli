package v3

import (
	"fmt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/log"
)

var (
	mockUserId             = int32(123)
	MockUserResourceId     = "u-123"
	mockOrganizationId     = int32(123)
	MockOrgResourceId      = "org-resource-id"
	MockEnvironmentId      = "testAccount"
	mockEmail              = "cli-mock-email@confluent.io"
	mockURL                = "http://test"
	usernameCredentialName = fmt.Sprintf("username-%s-%s", mockEmail, mockURL)
	apiKeyCredentialName   = fmt.Sprintf("api-key-%s", kafkaAPIKey)
	mockContextName        = fmt.Sprintf("login-%s-%s", mockEmail, mockURL)
	mockAuthToken          = "some.token.here"

	// kafka cluster
	kafkaClusterId     = "lkc-12345"
	anonymousKafkaId   = "anonymous-id"
	anonymousKafkaName = "anonymous-cluster"
	kafkaClusterName   = "toby-flenderson"
	bootstrapServer    = "https://toby-cluster:9092"
	kafkaApiEndpoint   = "https://is-the-worst:9092"
	kafkaAPIKey        = "costa"
	kafkaAPISecret     = "rica"

	// sr cluster
	srClusterId = "lsrc-test"
	srEndpoint  = "https://sr-test"
	srAPIKey    = "michael"
	srAPISecret = "scott"
)

func MockKafkaClusterId() string {
	return kafkaClusterId
}

func AuthenticatedCloudConfigMock() *Config {
	params := mockConfigParams{
		cliName:        "ccloud",
		contextName:    mockContextName,
		userId:         mockUserId,
		userResourceId: MockUserResourceId,
		username:       mockEmail,
		url:            MockUserResourceId,
		envId:          MockEnvironmentId,
		orgId:          mockOrganizationId,
		orgResourceId:  MockOrgResourceId,
		credentialName: usernameCredentialName,
	}
	return AuthenticatedConfigMock(params)
}

func AuthenticatedConfluentConfigMock() *Config {
	params := mockConfigParams{
		cliName:        "confluent",
		contextName:    mockContextName,
		userId:         mockUserId,
		userResourceId: MockUserResourceId,
		username:       mockEmail,
		url:            MockUserResourceId,
		envId:          MockEnvironmentId,
		orgId:          mockOrganizationId,
		orgResourceId:  MockOrgResourceId,
		credentialName: usernameCredentialName,
	}
	return AuthenticatedConfigMock(params)
}

func AuthenticatedConfigMockWithContextName(cliName string, contextName string) *Config {
	params := mockConfigParams{
		cliName:        cliName,
		contextName:    contextName,
		userId:         mockUserId,
		userResourceId: MockUserResourceId,
		username:       mockEmail,
		url:            MockUserResourceId,
		envId:          MockEnvironmentId,
		orgId:          mockOrganizationId,
		orgResourceId:  MockOrgResourceId,
		credentialName: usernameCredentialName,
	}
	return AuthenticatedConfigMock(params)
}

func APICredentialConfigMock() *Config {
	kafkaAPIKeyPair := createAPIKeyPair(kafkaAPIKey, kafkaAPISecret)

	credential := createAPIKeyCredential(apiKeyCredentialName, kafkaAPIKeyPair)
	contextState := createContextState(nil, "")

	platform := createPlatform(bootstrapServer, bootstrapServer)

	kafkaCluster := createKafkaCluster(anonymousKafkaId, anonymousKafkaName, kafkaAPIKeyPair)
	kafkaClusters := map[string]*v1.KafkaClusterConfig{
		kafkaCluster.ID: kafkaCluster,
	}

	conf := New(&config.Params{
		CLIName:    "ccloud",
		MetricSink: nil,
		Logger:     log.New(),
	})

	ctx, err := newContext(mockContextName, platform, credential, kafkaClusters, kafkaCluster.ID, nil, contextState, conf)
	if err != nil {
		panic(err)
	}
	setUpConfig(conf, ctx, platform, credential, contextState)
	return conf
}

func UnauthenticatedCloudConfigMock() *Config {
	c := AuthenticatedCloudConfigMock()
	c.Contexts = nil
	return c
}

type mockConfigParams struct {
	cliName        string
	contextName    string
	userId         int32
	userResourceId string
	username       string
	url            string
	envId          string
	orgId          int32
	orgResourceId  string
	credentialName string
}

func AuthenticatedConfigMock(params mockConfigParams) *Config {
	authConfig := createAuthConfig(params.userId, params.username, params.userResourceId, params.envId, params.orgId, params.orgResourceId)
	credential := createUsernameCredential(params.credentialName, authConfig)
	contextState := createContextState(authConfig, mockAuthToken)

	platform := createPlatform(params.url, params.url)

	kafkaAPIKeyPair := createAPIKeyPair(kafkaAPIKey, kafkaAPISecret)
	kafkaCluster := createKafkaCluster(kafkaClusterId, kafkaClusterName, kafkaAPIKeyPair)
	kafkaClusters := map[string]*v1.KafkaClusterConfig{
		kafkaCluster.ID: kafkaCluster,
	}

	srAPIKeyPair := createAPIKeyPair(srAPIKey, srAPISecret)
	srCluster := createSRCluster(srAPIKeyPair)
	srClusters := map[string]*v2.SchemaRegistryCluster{
		MockEnvironmentId: srCluster,
	}

	conf := New(&config.Params{
		CLIName:    params.cliName,
		MetricSink: nil,
		Logger:     log.New(),
	})

	ctx, err := newContext(params.contextName, platform, credential, kafkaClusters, kafkaCluster.ID, srClusters, contextState, conf)
	if err != nil {
		panic(err)
	}
	setUpConfig(conf, ctx, platform, credential, contextState)
	return conf
}

func createUsernameCredential(credentialName string, auth *v1.AuthConfig) *v2.Credential {
	credential := &v2.Credential{
		Name:           credentialName,
		Username:       auth.User.Email,
		CredentialType: v2.Username,
	}
	return credential
}

func createAPIKeyCredential(credentialName string, apiKeyPair *v0.APIKeyPair) *v2.Credential {
	credential := &v2.Credential{
		Name:           credentialName,
		APIKeyPair:     apiKeyPair,
		CredentialType: v2.APIKey,
	}
	return credential
}
func createPlatform(name, server string) *v2.Platform {
	platform := &v2.Platform{
		Name:   name,
		Server: server,
	}
	return platform
}

func createAuthConfig(userId int32, email string, userResourceId string, envId string, organizationId int32, orgResourceId string) *v1.AuthConfig {
	auth := &v1.AuthConfig{
		User: &orgv1.User{
			Id:         userId,
			Email:      email,
			ResourceId: userResourceId,
		},
		Account: &orgv1.Account{Id: envId},
		Organization: &orgv1.Organization{
			Id:         organizationId,
			ResourceId: orgResourceId,
		},
		Accounts: []*orgv1.Account{
			{Id: envId},
		},
	}
	return auth
}

func createContextState(authConfig *v1.AuthConfig, authToken string) *v2.ContextState {
	contextState := &v2.ContextState{
		Auth:      authConfig,
		AuthToken: authToken,
	}
	return contextState
}

func createAPIKeyPair(apiKey, apiSecret string) *v0.APIKeyPair {
	keyPair := &v0.APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	return keyPair
}

func createKafkaCluster(clusterID string, clusterName string, apiKeyPair *v0.APIKeyPair) *v1.KafkaClusterConfig {
	cluster := &v1.KafkaClusterConfig{
		ID:          clusterID,
		Name:        clusterName,
		Bootstrap:   bootstrapServer,
		APIEndpoint: kafkaApiEndpoint,
		APIKeys: map[string]*v0.APIKeyPair{
			apiKeyPair.Key: apiKeyPair,
		},
		APIKey: apiKeyPair.Key,
	}
	return cluster
}

func createSRCluster(apiKeyPair *v0.APIKeyPair) *v2.SchemaRegistryCluster {
	cluster := &v2.SchemaRegistryCluster{
		Id:                     srClusterId,
		SchemaRegistryEndpoint: srEndpoint,
		SrCredentials:          apiKeyPair,
	}
	return cluster
}

func setUpConfig(conf *Config, ctx *Context, platform *v2.Platform, credential *v2.Credential, contextState *v2.ContextState) {
	conf.Platforms[platform.Name] = platform
	conf.Credentials[credential.Name] = credential
	conf.ContextStates[ctx.Name] = contextState
	conf.Contexts[ctx.Name] = ctx
	conf.Contexts[ctx.Name].Config = conf
	conf.CurrentContext = ctx.Name
	if err := conf.Validate(); err != nil {
		panic(err)
	}
}

func AddEnvironmentToConfigMock(config *Config, envId string, envName string) {
	accounts := config.Context().State.Auth.Accounts
	config.Context().State.Auth.Accounts = append(accounts, &orgv1.Account{
		Id:   envId,
		Name: envName,
	})
}
