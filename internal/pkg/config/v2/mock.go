package v2

import (
	"fmt"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func AuthenticatedConfigMock() *Config {
	conf := New(&config.Params{
		CLIName:    "",
		MetricSink: nil,
		Logger:     log.New(),
	})
	conf.Logger = log.New()
	auth := &v1.AuthConfig{
		User: &orgv1.User{
			Id:    123,
			Email: "cli-mock-email@confluent.io",
		},
		Account: &orgv1.Account{Id: "testAccount"},
	}
	url := "http://test"
	credName := fmt.Sprintf("username-%s-%s", auth.User.Email, url)
	platform := &Platform{
		Name:   url,
		Server: url,
	}
	conf.Platforms[platform.Name] = platform
	credential := &Credential{
		Name:           credName,
		Username:       auth.User.Email,
		CredentialType: Username,
	}
	state := &ContextState{
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
	srClusters := map[string]*SchemaRegistryCluster{
		state.Auth.Account.Id: {
			Id:                     "lsrc-test",
			SchemaRegistryEndpoint: "https://sr-test",
			SrCredentials: &v0.APIKeyPair{
				Key:    "michael",
				Secret: "scott",
			},
		},
	}
	ctxName := fmt.Sprintf("login-%s-%s", auth.User.Email, url)
	ctx, err := newContext(ctxName, platform, credential, kafkaClusters, "lkc-0000", srClusters, state, conf)
	if err != nil {
		panic(err)
	}
	conf.ContextStates[ctx.Name] = state
	conf.Contexts[ctx.Name] = ctx
	conf.CurrentContext = ctx.Name
	if err := conf.Validate(); err != nil {
		panic(err)
	}
	return conf
}
