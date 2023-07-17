package name_conversions

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

func ConvertClusterNameToId(input string, environmentId string, v2Client *ccloudv2.Client) (string, error) {
	if cluster, _, err := v2Client.DescribeKafkaCluster(input, environmentId); err == nil {
		return cluster.GetId(), err
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

func ConvertEnvironmentNameToId(input string, v2Client *ccloudv2.Client) (string, error) {
	if env, err := v2Client.GetOrgEnvironment(input); err == nil {
		return env.GetId(), err
	}
	envs, err := v2Client.ListOrgEnvironments()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(envs))
}

func ConvertIamPoolNameToId(input string, providerId string, v2client *ccloudv2.Client) (string, error) {
	if pool, err := v2client.GetIdentityPool(input, providerId); err == nil {
		return pool.GetId(), err
	}
	pools, err := v2client.ListIdentityPools(providerId)
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(pools))
}

func ConvertIamProviderNameToId(input string, v2client *ccloudv2.Client) (string, error) {
	if provider, err := v2client.GetIdentityProvider(input); err == nil {
		return provider.GetId(), err
	}
	providers, err := v2client.ListIdentityProviders()
	if err != nil {
		return input, err
	}
	return ConvertV2NameToId(input, ConvertToPtrSlice(providers))
}
