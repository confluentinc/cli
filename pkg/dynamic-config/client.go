package dynamicconfig

import (
	"fmt"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func (d *DynamicContext) FetchAPIKeyError(apiKey, clusterId string) error {
	// check if this is API key exists server-side
	key, httpResp, err := d.V2Client.GetApiKey(apiKey)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	// check if the key is for the right cluster
	ok := key.Spec.Resource.Id == clusterId
	// this means the requested api-key belongs to a different cluster
	if !ok {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`invalid API key "%s" for resource "%s"`, apiKey, clusterId),
			fmt.Sprintf("To list API key that belongs to resource \"%[1]s\", use `confluent api-key list --resource %[1]s`.\nTo create new API key for resource \"%[1]s\", use `confluent api-key create --resource %[1]s`.", clusterId),
		)
	}
	// the requested api-key exists, but the secret is not saved locally
	return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterId}
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
