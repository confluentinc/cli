package dynamicconfig

import (
	"fmt"
	"strings"
	"time"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (d *DynamicContext) FetchCluster(clusterId string) (*v1.KafkaClusterConfig, error) {
	environmentId, err := d.EnvironmentId()
	if err != nil {
		return nil, err
	}

	cluster, httpResp, err := d.V2Client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, httpResp)
	}

	config := &v1.KafkaClusterConfig{
		ID:           *cluster.Id,
		Name:         *cluster.Spec.DisplayName,
		Bootstrap:    strings.TrimPrefix(*cluster.Spec.KafkaBootstrapEndpoint, "SASL_SSL://"),
		RestEndpoint: *cluster.Spec.HttpEndpoint,
		APIKeys:      make(map[string]*v1.APIKeyPair),
		LastUpdate:   time.Now(),
	}

	return config, nil
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
	return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterID}
}

func (d *DynamicContext) FetchSchemaRegistryByEnvironmentId(accountId string) (srcmv2.SrcmV2Cluster, error) {
	existingClusters, err := d.V2Client.GetSchemaRegistryClustersByEnvironment(accountId)
	if err != nil {
		return srcmv2.SrcmV2Cluster{}, err
	}
	if len(existingClusters) > 0 {
		return existingClusters[0], nil
	}
	return srcmv2.SrcmV2Cluster{}, errors.NewSRNotEnabledError()
}
