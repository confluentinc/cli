package config

import (
	"context"
	"strings"
)

type contextResolver struct {
	client  *contextClient
}

func NewResolver(ctx *Context) *contextResolver {
	ctxClient := NewContextClient(ctx)
	return &contextResolver{
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
	srCluster, err := c.client.FetchSchemaRegistryByAccountId(context.Background(), accountId)
	if err != nil {
		return nil, err
	}
	cluster := &SchemaRegistryCluster{
		Id:                     srCluster.Id,
		SchemaRegistryEndpoint: srCluster.Endpoint,
		SrCredentials:          nil, // For now.
	}
	return cluster, nil
}
