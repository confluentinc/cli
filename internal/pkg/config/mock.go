package config

import (
	"fmt"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/confluentinc/ccloudapis/org/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/sdk"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func AuthenticatedConfigMock() *Config {
	conf := New()
	conf.Logger = log.New()
	conf.Version = version.NewVersion("", "", "", "", "", "", "")
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
	baseClient := &ccloud.Client{
		Params:         nil,
		Auth:           &mock.Auth{},
		Account:        &mock.Account{},
		Kafka:          &mock.Kafka{},
		SchemaRegistry: &mock.SchemaRegistry{},
		Connect:        &mock.Connect{},
		User:           &mock.User{},
		APIKey:         &mock.APIKey{},
		KSQL:           &mock.MockKSQL{},
		Metrics:        &mock.Metrics{},
	}
	client := sdk.NewClient(baseClient, conf.Logger)
	kafkaClusters := map[string]*KafkaClusterConfig{
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
	}
	srClusters := map[string]*SchemaRegistryCluster{
		state.Auth.Account.Id: {
			Id: "lsrc-test",
			SchemaRegistryEndpoint: "https://sr-test",
			SrCredentials: &APIKeyPair{
				Key:    "michael",
				Secret: "scott",
			},
		},
	}
	ctx, err := newContext("test-context", platform, credential, kafkaClusters, "lkc-0000", srClusters, state, client, conf)
	if err != nil {
		panic(err)
	}
	conf.ContextStates[ctx.Name] = state
	conf.Contexts[ctx.Name] = ctx
	conf.CurrentContext = ctx.Name
	if err := conf.validate(); err != nil {
		panic(err)
	}
	return conf
}
