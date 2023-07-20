package ccloudv2

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	v1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
)

// KafkaClusterNameToId attempts to convert from a valid kafka cluster name to its ID, if it fails it returns the input string
func KafkaClusterNameToId(input string, environmentId string, client *Client) (string, error) {
	clusters, err := client.ListKafkaClusters(environmentId)
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

// EnvironmentNameToId attempts to convert from a valid environment name to its ID, if it fails it returns an error
func EnvironmentNameToId(input string, client *Client, tryGetFirst bool) (string, error) {
	if tryGetFirst {
		if env, err := client.GetOrgEnvironment(input); err == nil {
			return env.GetId(), err
		}
	}
	envs, err := client.ListOrgEnvironments()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(envs))
}

// IamPoolNameToId attempts to convert from a valid iam pool name to its ID, if it fails it returns an error
func IamPoolNameToId(input string, providerId string, client *Client) (string, error) {
	pools, err := client.ListIdentityPools(providerId)
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(pools))
}

// IamProviderNameToId attempts to convert from a valid iam provider name to its ID, if it fails it returns an error
func IamProviderNameToId(input string, client *Client, tryGetFirst bool) (string, error) {
	if tryGetFirst {
		if provider, err := client.GetIdentityProvider(input); err == nil {
			return provider.GetId(), err
		}
	}
	providers, err := client.ListIdentityProviders()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(providers))
}

// IamServiceAccountNameToId attempts to convert from a valid iam service account name to its ID, if it fails it returns an error
func IamServiceAccountNameToId(input string, client *Client, tryGetFirst bool) (string, error) {
	if tryGetFirst {
		if serviceAccount, _, err := client.GetIamServiceAccount(input); err == nil {
			return serviceAccount.GetId(), err
		}
	}
	serviceAccounts, err := client.ListIamServiceAccounts()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(serviceAccounts))
}

// QuotaNameToId attempts to convert from a valid quota name to its ID, if it fails it returns an error
func QuotaNameToId(input string, clusterId string, environmentId string, client *Client) (string, error) {
	quotas, err := client.ListKafkaQuotas(clusterId, environmentId)
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
