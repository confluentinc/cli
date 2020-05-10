package cmd

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type contextClient struct {
	context *DynamicContext
}

// NewContextClient returns a new contextClient, with the specified context and a client.
func NewContextClient(ctx *DynamicContext) *contextClient {
	return &contextClient{
		context: ctx,
	}
}

func (c *contextClient) FetchCluster(cmd *cobra.Command, clusterId string) (*schedv1.KafkaCluster, error) {
	envId, err := c.context.AuthenticatedEnvId(cmd)
	if err != nil {
		return nil, err
	}
	req := &schedv1.KafkaCluster{AccountId: envId, Id: clusterId}
	kc, err := c.context.client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return kc, nil
}

func (c *contextClient) FetchAPIKeyError(cmd *cobra.Command, apiKey string, clusterID string) error {
	envId, err := c.context.AuthenticatedEnvId(cmd)
	if err != nil {
		return err
	}
	// check if this is API key exists server-side
	key, err := c.context.client.APIKey.Get(context.Background(), &schedv1.ApiKey{AccountId: envId, Key: apiKey})
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

func (c *contextClient) FetchSchemaRegistryByAccountId(context context.Context, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingClusters, err := c.context.client.SchemaRegistry.GetSchemaRegistryClusters(context, &schedv1.SchemaRegistryCluster{
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

func (c *contextClient) FetchSchemaRegistryById(context context.Context, id string, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingCluster, err := c.context.client.SchemaRegistry.GetSchemaRegistryCluster(context, &schedv1.SchemaRegistryCluster{
		Id:        id,
		AccountId: accountId,
	})
	if err != nil {
		return nil, err
	}
	if existingCluster == nil {
		return nil, errors.ErrNoSrEnabled
	} else {
		return existingCluster, nil
	}
}
