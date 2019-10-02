package config

import (
	"fmt"

	"github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func AuthenticatedConfigMock() *Config {
	conf := New()
	conf.Logger = log.New()
	auth := &AuthConfig{
		User: &v1.User{
			Id:    123,
			Email: "cli-mock-email@confluent.io",
		},
		Account: &v1.Account{Id: "testAccount"},
	}
	url := "http://test"
	name := fmt.Sprintf("login-%s-%s", auth.User.Email, url)
	platform := &Platform{
		Name:   name,
		Server: url,
	}
	conf.Platforms[platform.Name] = platform
	credential := &Credential{
		Name:           name,
		Username:       auth.User.Email,
		CredentialType: Username,
	}
	state := &ContextState{
		Auth:      auth,
		AuthToken: "some.token.here",
	}
	conf.Credentials[credential.Name] = credential
	ctx := &Context{
		Name:           "test-context",
		Platform:       platform,
		PlatformName:   platform.Name,
		Credential:     credential,
		CredentialName: credential.Name,
		KafkaClusters: map[string]*KafkaClusterConfig{
			"lkc-0000": {
				ID:          "lkc-0000",
				Name:        "toby-flenderson",
				Bootstrap:   "http://toby-cluster",
				APIEndpoint: "http://is-the-worst",
				APIKeys: map[string]*APIKeyPair{
					"costa": {
						Key:    "costa",
						Secret: "rica",
					},
				},
				APIKey: "costa",
			},
		},
		Kafka: "lkc-0000",
		SchemaRegistryClusters: map[string]*SchemaRegistryCluster{
			state.Auth.Account.Id: {
				SchemaRegistryEndpoint: "https://sr-test",
				SrCredentials: &APIKeyPair{
					Key:    "michael",
					Secret: "scott",
				},
			},
		},
		State: state,
	}
	if err := ctx.Validate(); err != nil {
		panic(err)
	}
	conf.Contexts[ctx.Name] = ctx
	conf.CurrentContext = ctx.Name
	if err := conf.Validate(); err != nil {
		panic(err)
	}
	return conf
}
