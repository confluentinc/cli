package config

import (
	"context"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"
)

type contextResolver struct {
	context *Context
	client  *contextClient
}

func NewResolver(context *Context, client *ccloud.Client) *contextResolver {
	ctxClient := NewContextClient(context, client)
	return &contextResolver{
		context: context,
		client:  ctxClient,
	}
}

func (c *contextResolver) ResolveCluster(clusterId string) (*KafkaClusterConfig, error) {
	kc, err := c.client.FetchCluster(clusterId)
	if err != nil {
		return nil, err
	}
	cluster := &KafkaClusterConfig{
		ID:          clusterId,
		Name:        kc.Name,
		Bootstrap:   strings.TrimPrefix(kc.Endpoint, "SASL_SSL://"),
		APIEndpoint: kc.ApiEndpoint,
		APIKeys:     make(map[string]*APIKeyPair),
	}
	return cluster, nil
}

func (c *contextResolver) ResolveSchemaRegistryByAccountId(accountId string) (*SchemaRegistryCluster, error) {
	srCluster, err := c.client.FetchSchemaRegistryByAccountId(accountId, context.Background())
	if err != nil {
		return nil, err
	}
	cluster := &SchemaRegistryCluster{
		SchemaRegistryEndpoint: srCluster.Endpoint,
		SrCredentials:          nil, // For now.
	}
	return cluster, nil
}
