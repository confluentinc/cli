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
		return "", err
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
		return "", err
	}
	envPtrs := ConvertToPtrSlice(envs)
	return ConvertV2NameToId(input, envPtrs)
}

//func ConvertIamPoolNameToId(input string, providerId string, v2client *ccloudv2.Client) (string, error) {
//	pools, err := v2client.ListIdentityPools()
//}
