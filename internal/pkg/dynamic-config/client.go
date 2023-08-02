package dynamicconfig

import (
	"fmt"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (d *DynamicContext) FetchAPIKeyError(apiKey, clusterID string) error {
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
