package name_conversions

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func ConvertClusterNameToId(input string, environmentId string, v2Client *ccloudv2.Client) (string, error) {
	clusters, err := v2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return "", err
	}
	clusterPtrs := resource.ConvertToPtrSlice(clusters)
	specPtrs := make([]*cmkv2.CmkV2ClusterSpec, len(clusters))
	for i := range clusters {
		specPtrs[i] = clusterPtrs[i].Spec
	}
	return resource.ConvertSpecNameToId(input, clusterPtrs, specPtrs)
}
func ConvertEnvironmentNameToId(input string, v2Client *ccloudv2.Client) (string, error) {
	envs, err := v2Client.ListOrgEnvironments()
	if err != nil {
		return "", err
	}
	envPtrs := resource.ConvertToPtrSlice(envs)
	return resource.ConvertV2NameToId(input, envPtrs)
}
