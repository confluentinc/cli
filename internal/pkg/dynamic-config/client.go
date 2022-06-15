package dynamicconfig

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type contextClient struct {
	context *DynamicContext
}

// NewContextClient returns a new contextClient, with the specified context and a client.
func NewContextClient(ctx *DynamicContext) *contextClient {
	return &contextClient{context: ctx}
}

func (c *contextClient) FetchCluster(clusterId string) (*schedv1.KafkaCluster, error) {
	envId, err := c.context.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	req := &schedv1.KafkaCluster{AccountId: envId, Id: clusterId}
	cluster, err := c.context.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId)
	}

	return cluster, nil
}

func (c *contextClient) FetchAPIKeyError(apiKey string, clusterID string) error {
	// check if this is API key exists server-side
	key, _, err := c.context.V2Client.GetApiKey(apiKey)
	if err != nil {
		return err
	}
	// check if the key is for the right cluster
	ok := key.Spec.Resource.Id == clusterID
	// this means the requested api-key belongs to a different cluster
	if !ok {
		errorMsg := fmt.Sprintf(errors.InvalidAPIKeyErrorMsg, apiKey, clusterID)
		suggestionsMsg := fmt.Sprintf(errors.InvalidAPIKeySuggestions, clusterID, clusterID, clusterID, clusterID)
		return errors.NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	// this means the requested api-key exists, but we just don't have the secret saved locally
	return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterID}
}

func (c *contextClient) FetchSchemaRegistryByAccountId(context context.Context, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingClusters, err := c.context.Client.SchemaRegistry.GetSchemaRegistryClusters(context, &schedv1.SchemaRegistryCluster{
		AccountId: accountId,
		Name:      "account schema-registry",
	})
	if err != nil {
		return nil, err
	}
	if len(existingClusters) > 0 {
		return existingClusters[0], nil
	}
	return nil, errors.NewSRNotEnabledError()
}

func (c *contextClient) FetchSchemaRegistryById(context context.Context, id string, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingCluster, err := c.context.Client.SchemaRegistry.GetSchemaRegistryCluster(context, &schedv1.SchemaRegistryCluster{
		Id:        id,
		AccountId: accountId,
	})
	if err != nil {
		return nil, err
	}
	if existingCluster == nil {
		return nil, errors.NewSRNotEnabledError()
	} else {
		return existingCluster, nil
	}
}
