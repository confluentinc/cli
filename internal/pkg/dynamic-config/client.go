package dynamicconfig

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (d *DynamicContext) FetchCluster(clusterId string) (*schedv1.KafkaCluster, error) {
	environmentId, err := d.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	cluster := &schedv1.KafkaCluster{
		AccountId: environmentId,
		Id:        clusterId,
	}

	cluster, err = d.Client.Kafka.Describe(context.Background(), cluster)
	return cluster, errors.CatchKafkaNotFoundError(err, clusterId)
}

func (d *DynamicContext) FetchAPIKeyError(apiKey string, clusterID string) error {
	// check if this is API key exists server-side
	key, _, err := d.V2Client.GetApiKey(apiKey)
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

func (d *DynamicContext) FetchSchemaRegistryByAccountId(context context.Context, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingClusters, err := d.Client.SchemaRegistry.GetSchemaRegistryClusters(context, &schedv1.SchemaRegistryCluster{
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

func (d *DynamicContext) FetchSchemaRegistryById(context context.Context, id string, accountId string) (*schedv1.SchemaRegistryCluster, error) {
	existingCluster, err := d.Client.SchemaRegistry.GetSchemaRegistryCluster(context, &schedv1.SchemaRegistryCluster{
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
