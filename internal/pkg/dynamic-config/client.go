package dynamicconfig

import (
	"context"
	"fmt"
	"strings"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (d *DynamicContext) FetchCluster(clusterId string) (*v1.KafkaClusterConfig, error) {
	environmentId, err := d.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	cluster, httpResp, err := d.V2Client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchCCloudV2Error(err, httpResp)
	}

	apiEndpoint, err := getKafkaApiEndpoint(d.Client, clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, nil)
	}

	config := &v1.KafkaClusterConfig{
		ID:           *cluster.Id,
		Name:         *cluster.Spec.DisplayName,
		Bootstrap:    strings.TrimPrefix(*cluster.Spec.KafkaBootstrapEndpoint, "SASL_SSL://"),
		APIEndpoint:  apiEndpoint,
		RestEndpoint: *cluster.Spec.HttpEndpoint,
		APIKeys:      make(map[string]*v1.APIKeyPair),
		LastUpdate:   time.Now(),
	}

	return config, nil
}

func getKafkaApiEndpoint(client *ccloud.Client, clusterId, environmentId string) (string, error) {
	cluster := &schedv1.KafkaCluster{
		Id:        clusterId,
		AccountId: environmentId,
	}

	cluster, err := client.Kafka.Describe(context.Background(), cluster)
	if err != nil {
		return "", err
	}

	return cluster.ApiEndpoint, nil
}

func (d *DynamicContext) FetchAPIKeyError(apiKey string, clusterID string) error {
	// check if this is API key exists server-side
	key, httpResp, err := d.V2Client.GetApiKey(apiKey)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	// check if the key is for the right cluster
	ok := key.Spec.Resource.Id == clusterID
	// this means the requested api-key belongs to a different cluster
	if !ok {
		errorMsg := fmt.Sprintf(errors.InvalidAPIKeyErrorMsg, apiKey, clusterID)
		suggestionsMsg := fmt.Sprintf(errors.InvalidAPIKeySuggestions, clusterID, clusterID, clusterID, clusterID)
		return errors.NewErrorWithSuggestions(errorMsg, suggestionsMsg)
	}
	// the requested api-key exists, but the secret is not saved locally
	return errors.NewUnconfiguredAPISecretError(apiKey, clusterID)
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
