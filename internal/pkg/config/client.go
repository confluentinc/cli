package config

import (
	"context"
	"fmt"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	v1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type contextClient struct {
	context *Context
}

// NewContextClient returns a new contextClient, with the specified context and its client. 
// and an injected CCLoud client, or a dynamically generated client if passed a nil client. 
func NewContextClient(ctx *Context) *contextClient {
	return &contextClient{
		context: ctx,
	}
}

func (c *contextClient) FetchCluster(clusterId string) (*kafkav1.KafkaCluster, error) {
	state, err := c.context.AuthenticatedState()
	if err != nil {
		return nil, err
	}
	req := &kafkav1.KafkaCluster{AccountId: state.Auth.Account.Id, Id: clusterId}
	kc, err := c.context.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		if err != ccloud.ErrNotFound {
			return nil, err
		}
		return nil, errors.ErrNoKafkaContext
	}
	return kc, nil
}

func (c *contextClient) FetchAPIKeyError(apiKey, clusterID string) error {
	state, err := c.context.AuthenticatedState()
	if err != nil {
		return err
	}
	// check if this is API key exists server-side
	key, err := c.context.Client.APIKey.Get(context.Background(), &authv1.ApiKey{AccountId: state.Auth.Account.Id, Key: apiKey})
	if err != nil {
		return err
	}
	// check if the key is for the right cluster
	found := false
	for _, c := range key.LogicalClusters {
		if c.Id == clusterID {
			found = true
			break
		}
	}
	// this means the requested api-key belongs to a different cluster
	if !found {
		return fmt.Errorf("invalid api-key %s for cluster %s", apiKey, clusterID)
	}
	// this means the requested api-key exists, but we just don't have the secret saved locally
	return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterID}
}

func (c *contextClient) FetchSchemaRegistryByAccountId(context context.Context, accountId string) (*v1.SchemaRegistryCluster, error) {
	existingClusters, err := c.context.Client.SchemaRegistry.GetSchemaRegistryClusters(context, &v1.SchemaRegistryCluster{
		AccountId: accountId,
		Name:      "account schema-registry",
	})
	if err != nil {
		return nil, err
	}
	if len(existingClusters) > 0 {
		return existingClusters[0], nil
	}
	return nil, errors.ErrNoSrEnabled
}

func (c *contextClient) FetchSchemaRegistryById(context context.Context, id string, accountId string) (*v1.SchemaRegistryCluster, error) {
	existingCluster, err := c.context.Client.SchemaRegistry.GetSchemaRegistryCluster(context, &v1.SchemaRegistryCluster{
		Id:        id,
		AccountId: accountId,
	})
	if err != nil {
		return nil, err
	}
	//if len(existingClusters) > 0 {
	//	return existingClusters[0], nil
	//}
	if existingCluster == nil {
		return nil, errors.ErrNoSrEnabled
	} else {
		return existingCluster, nil
	}
}
