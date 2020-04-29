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
	mockEmail       = "cli-mock-email@confluent.io"
	mockURL         = "http://test"
	MockContextName = fmt.Sprintf("login-%s-%s", mockEmail, mockURL)
)

func AuthenticatedConfigMock(cliName string) *Config {
	conf := New(&config.Params{
		CLIName:    cliName,
		MetricSink: nil,
		Logger:     log.New(),
	})
	conf.Logger = log.New()
	auth := &v1.AuthConfig{
		User: &orgv1.User{
			Id:    123,
			Email: mockEmail,
		},
		Account: &orgv1.Account{Id: "testAccount"},
	}
	credName := fmt.Sprintf("username-%s-%s", auth.User.Email, mockURL)
	platform := &v2.Platform{
		Name:   mockURL,
		Server: mockURL,
	}
	conf.Platforms[platform.Name] = platform
	credential := &v2.Credential{
		Name:           credName,
		Username:       auth.User.Email,
		CredentialType: v2.Username,
	}
	state := &v2.ContextState{
		Auth:      auth,
		AuthToken: "some.token.here",
	}
	conf.Credentials[credential.Name] = credential
	kafkaClusters := map[string]*v1.KafkaClusterConfig{
		"lkc-0000": {
			ID:          "lkc-0000",
			Name:        "toby-flenderson",
			Bootstrap:   "http://toby-cluster",
			APIEndpoint: "http://is-the-worst",
			APIKeys: map[string]*v0.APIKeyPair{
				"costa": {
					Key:    "costa",
					Secret: "rica",
				},
			},
			APIKey: "costa",
		},
	}
	srClusters := map[string]*v2.SchemaRegistryCluster{
		state.Auth.Account.Id: {
			Id:                     "lsrc-test",
			SchemaRegistryEndpoint: "https://sr-test",
			SrCredentials: &v0.APIKeyPair{
				Key:    "michael",
				Secret: "scott",
			},
		},
	}
	ctxName := MockContextName
	ctx, err := newContext(ctxName, platform, credential, kafkaClusters, "lkc-0000", srClusters, state, conf)
	if err != nil {
		panic(err)
	}
	conf.ContextStates[ctx.Name] = state
	conf.Contexts[ctx.Name] = ctx
	conf.Contexts[ctx.Name].Config = conf
	conf.CurrentContext = ctx.Name
	if err := conf.Validate(); err != nil {
		panic(err)
	}
	return conf
}

func AuthenticatedCloudConfigMock() *Config {
	return AuthenticatedConfigMock("ccloud")
}

func AuthenticatedConfluentConfigMock() *Config {
	return AuthenticatedConfigMock("")
}
