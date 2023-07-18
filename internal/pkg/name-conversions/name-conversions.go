package nameconversions

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

// ConvertClusterNameToId attempts to convert from a valid cluster name to its ID, if it fails it returns the input string
func ConvertClusterNameToId(input string, environmentId string, v2Client *ccloudv2.Client, forFlagValue bool) (string, error) {
	if forFlagValue {
		if cluster, _, err := v2Client.DescribeKafkaCluster(input, environmentId); err == nil {
			return cluster.GetId(), err
		}
	}
	clusters, err := v2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return input, err
	}
	clusterPtrs := ConvertToPtrSlice(clusters)
	specPtrs := make([]*cmkv2.CmkV2ClusterSpec, len(clusters))
	for i := range clusters {
		specPtrs[i] = clusterPtrs[i].Spec
	}
	return ConvertSpecNameToId(input, clusterPtrs, specPtrs)
}

// ConvertEnvironmentNameToId attempts to convert from a valid environment name to its ID, if it fails it returns the input string
func ConvertEnvironmentNameToId(input string, v2Client *ccloudv2.Client, forFlagValue bool) (string, error) {
	if forFlagValue {
		if env, err := v2Client.GetOrgEnvironment(input); err == nil {
			return env.GetId(), err
		}
	}
	envs, err := v2Client.ListOrgEnvironments()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(envs))
}

// ConvertIamPoolNameToId attempts to convert from a valid iam pool name to its ID, if it fails it returns the input string
func ConvertIamPoolNameToId(input string, providerId string, v2client *ccloudv2.Client, forFlagValue bool) (string, error) {
	if forFlagValue {
		if pool, err := v2client.GetIdentityPool(input, providerId); err == nil {
			return pool.GetId(), err
		}
	}
	pools, err := v2client.ListIdentityPools(providerId)
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(pools))
}

// ConvertIamProviderNameToId attempts to convert from a valid iam provider name to its ID, if it fails it returns the input string
func ConvertIamProviderNameToId(input string, v2client *ccloudv2.Client, forFlagValue bool) (string, error) {
	if forFlagValue {
		if provider, err := v2client.GetIdentityProvider(input); err == nil {
			return provider.GetId(), err
		}
	}
	providers, err := v2client.ListIdentityProviders()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(providers))
}

// ConvertIamServiceAccountNameToId attempts to convert from a valid iam service account name to its ID, if it fails it returns the input string
func ConvertIamServiceAccountNameToId(input string, v2client *ccloudv2.Client, forFlagValue bool) (string, error) {
	if forFlagValue {
		if serviceAccount, _, err := v2client.GetIamServiceAccount(input); err == nil {
			return serviceAccount.GetId(), err
		}
	}
	serviceAccounts, err := v2client.ListIamServiceAccounts()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(serviceAccounts))
}

// ConvertQuotaNameToId attempts to convert from a valid quota name to its ID, if it fails it returns the input string
func ConvertQuotaNameToId(input string, clusterId string, environmentId string, v2Client *ccloudv2.Client) (string, error) {
	quotas, err := v2Client.ListKafkaQuotas(clusterId, environmentId)
	if err != nil {
		return input, err
	}
	quotaPtrs := ConvertToPtrSlice(quotas)
	specPtrs := make([]*v1.KafkaQuotasV1ClientQuotaSpec, len(quotas))
	for i := range quotas {
		specPtrs[i] = quotaPtrs[i].Spec
	}
	return ConvertSpecNameToId(input, quotaPtrs, specPtrs)
}
