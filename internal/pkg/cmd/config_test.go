package cmd

import (
	"context"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/confluentinc/ccloudapis/org/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	"testing"
)

const (
	currentContext = "current_ctx"
)

func getConfig(endpoint string) config.Config {
	return config.Config{
		CurrentContext: currentContext,
		Contexts: map[string]*config.Context{
			currentContext: {
				SchemaRegistryClusters: map[string]*config.SchemaRegistryCluster{
					"env": {
						SchemaRegistryEndpoint: endpoint,
					},
				},
			},
		},
		Auth: &config.AuthConfig{Account: &v1.Account{Id: "env"}},
	}
}

func getClient() *ccloud.Client {
	return &ccloud.Client{
		SchemaRegistry: &mock.SchemaRegistry{
			GetSchemaRegistryClustersFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) ([]*srv1.SchemaRegistryCluster, error) {
				return []*srv1.SchemaRegistryCluster{&srv1.SchemaRegistryCluster{Endpoint: "remotehost"}}, nil
			},
		},
	}
}

func TestSchemaRegistryURL(t *testing.T) {
	// Found locally
	cfg := getConfig("localhost")
	client := getClient()
	ch := ConfigHelper{
		Config: &cfg,
		Client: client,
	}
	if found, _ := ch.SchemaRegistryURL(nil); found != "localhost" {
		t.Errorf("expected %v, but found %v", "localhost", found)
		t.Fail()
	}

	// Not found locally
	cfg = getConfig("")
	ch = ConfigHelper{
		Config: &cfg,
		Client: client,
	}
	if found, _ := ch.SchemaRegistryURL(nil); found != "remotehost" {
		t.Errorf("expected %v, but found %v", "remotehost", found)
		t.Fail()
	}

}
