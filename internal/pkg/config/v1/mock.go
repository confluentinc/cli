package v1

import (
	"fmt"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	testserver "github.com/confluentinc/cli/test/test-server"
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
	bootstrapServer    = "SASL_SSL://pkc-abc123.us-west2.gcp.confluent.cloud:9092"
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
	return AuthenticatedToOrgCloudConfigMock(mockOrganizationId, MockOrgResourceId)
}

func AuthenticatedToOrgCloudConfigMock(orgId int32, orgResourceId string) *Config {
	params := mockConfigParams{
		contextName:    mockContextName,
		userId:         mockUserId,
		userResourceId: MockUserResourceId,
		username:       mockEmail,
		url:            testserver.TestCloudUrl.String(),
		envId:          MockEnvironmentId,
		orgId:          orgId,
		orgResourceId:  orgResourceId,
		credentialName: usernameCredentialName,
	}
	return AuthenticatedConfigMock(params)
}

func AuthenticatedOnPremConfigMock() *Config {
	params := mockConfigParams{
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

func AuthenticatedConfigMockWithContextName(contextName string) *Config {
	params := mockConfigParams{
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
	kafkaClusters := map[string]*KafkaClusterConfig{
		kafkaCluster.ID: kafkaCluster,
	}

	cfg := New()

	ctx, err := newContext(mockContextName, platform, credential, kafkaClusters, kafkaCluster.ID, nil, contextState, cfg, "")
	if err != nil {
		panic(err)
	}
	setUpConfig(cfg, ctx, platform, credential, contextState)
	return cfg
}

func UnauthenticatedCloudConfigMock() *Config {
	c := AuthenticatedCloudConfigMock()
	c.Contexts = nil
	return c
}

type mockConfigParams struct {
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
	kafkaClusters := map[string]*KafkaClusterConfig{
		kafkaCluster.ID: kafkaCluster,
	}

	srAPIKeyPair := createAPIKeyPair(srAPIKey, srAPISecret)
	srCluster := createSRCluster(srAPIKeyPair)
	srClusters := map[string]*SchemaRegistryCluster{
		MockEnvironmentId: srCluster,
	}

	cfg := New()
	cfg.IsTest = true

	ctx, err := newContext(params.contextName, platform, credential, kafkaClusters, kafkaCluster.ID, srClusters, contextState, cfg, params.orgResourceId)
	if err != nil {
		panic(err)
	}
	setUpConfig(cfg, ctx, platform, credential, contextState)
	return cfg
}

func createUsernameCredential(credentialName string, auth *AuthConfig) *Credential {
	return &Credential{
		Name:           credentialName,
		Username:       auth.User.Email,
		CredentialType: Username,
	}
}

func createAPIKeyCredential(credentialName string, apiKeyPair *APIKeyPair) *Credential {
	return &Credential{
		Name:           credentialName,
		APIKeyPair:     apiKeyPair,
		CredentialType: APIKey,
	}
}

func createPlatform(name, server string) *Platform {
	return &Platform{
		Name:   name,
		Server: server,
	}
}

func createAuthConfig(userId int32, email, userResourceId, envId string, organizationId int32, orgResourceId string) *AuthConfig {
	return &AuthConfig{
		User: &ccloudv1.User{
			Id:         userId,
			Email:      email,
			ResourceId: userResourceId,
		},
		Account: &ccloudv1.Account{Id: envId},
		Organization: &ccloudv1.Organization{
			Id:         organizationId,
			ResourceId: orgResourceId,
		},
		Accounts: []*ccloudv1.Account{{Id: envId}},
	}
}

func createContextState(authConfig *AuthConfig, authToken string) *ContextState {
	return &ContextState{
		Auth:      authConfig,
		AuthToken: authToken,
	}
}

func createAPIKeyPair(apiKey, apiSecret string) *APIKeyPair {
	return &APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
}

func createKafkaCluster(clusterID, clusterName string, apiKeyPair *APIKeyPair) *KafkaClusterConfig {
	return &KafkaClusterConfig{
		ID:         clusterID,
		Name:       clusterName,
		Bootstrap:  bootstrapServer,
		APIKeys:    map[string]*APIKeyPair{apiKeyPair.Key: apiKeyPair},
		APIKey:     apiKeyPair.Key,
		LastUpdate: time.Now(),
	}
}

func createSRCluster(apiKeyPair *APIKeyPair) *SchemaRegistryCluster {
	return &SchemaRegistryCluster{
		Id:                     srClusterId,
		SchemaRegistryEndpoint: srEndpoint,
		SrCredentials:          apiKeyPair,
	}
}

func setUpConfig(conf *Config, ctx *Context, platform *Platform, credential *Credential, contextState *ContextState) {
	conf.Platforms[platform.Name] = platform
	conf.Credentials[credential.Name] = credential
	conf.ContextStates[ctx.Name] = contextState
	conf.Contexts[ctx.Name] = ctx
	conf.Contexts[ctx.Name].Config = conf
	conf.CurrentContext = ctx.Name
	conf.IsTest = true
	if err := conf.Validate(); err != nil {
		panic(err)
	}
}

func AddEnvironmentToConfigMock(cfg *Config, id, name string) {
	ctx := cfg.Context()
	ctx.State.Auth.Accounts = append(ctx.GetEnvironments(), &ccloudv1.Account{
		Id:   id,
		Name: name,
	})
}
