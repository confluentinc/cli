package kafka

import (
	"fmt"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/ccstructs"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/kafkausagelimits"
)

func toCreateTopicConfigs(topicConfigsMap map[string]string) []cckafkarestv3.ConfigData {
	topicConfigs := make([]cckafkarestv3.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = cckafkarestv3.ConfigData{
			Name:  k,
			Value: *cckafkarestv3.NewNullableString(&val),
		}
		i++
	}
	return topicConfigs
}

func toCreateTopicConfigsOnPrem(topicConfigsMap map[string]string) []cpkafkarestv3.ConfigData {
	topicConfigs := make([]cpkafkarestv3.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = cpkafkarestv3.ConfigData{
			Name:  k,
			Value: &val,
		}
		i++
	}
	return topicConfigs
}

func toAlterConfigBatchRequestData(configsMap map[string]string) []cckafkarestv3.AlterConfigBatchRequestDataData {
	configs := make([]cckafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		val := val
		configs[i] = cckafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: *cckafkarestv3.NewNullableString(&val),
		}
		i++
	}
	return configs
}

func handleOpenApiError(httpResp *http.Response, err error, client *cpkafkarestv3.APIClient) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkarest.NewError(client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func isClusterResizeInProgress(currentCluster *cmkv2.CmkV2Cluster) error {
	if currentCluster.Status.Phase == ccloudv2.StatusProvisioning {
		return fmt.Errorf("your cluster is still provisioning, so it can't be updated yet; please retry in a few minutes")
	}
	if isExpanding(currentCluster) {
		return fmt.Errorf("your cluster is expanding; please wait for that operation to complete before updating again")
	}
	if isShrinking(currentCluster) {
		return fmt.Errorf("your cluster is shrinking; please wait for that operation to complete before updating again")
	}
	return nil
}

func getCmkClusterIngressAndEgressMbps(currentMaxEcku int32, limits *kafkausagelimits.Limits) (int32, int32) {
	ingress, egress := limits.GetIngress(), limits.GetEgress()

	if limits.GetMaxEcku() == nil {
		return ingress, egress
	}

	// Scale limits by cluster's max eCKU when limits are set per eCKU
	// Use default max ecku when currentMaxEcku is not set
	if currentMaxEcku == 0 && limits.GetMaxEcku() != nil {
		currentMaxEcku = limits.GetMaxEcku().Value
	}

	return ingress * currentMaxEcku, egress * currentMaxEcku
}

func getCmkClusterType(cluster *cmkv2.CmkV2Cluster) string {
	if isBasic(cluster) {
		return ccstructs.Sku_name[2]
	}
	if isStandard(cluster) {
		return ccstructs.Sku_name[3]
	}
	if isDedicated(cluster) {
		return ccstructs.Sku_name[4]
	}
	if isEnterprise(cluster) {
		return ccstructs.Sku_name[6]
	}
	if isFreight(cluster) {
		return ccstructs.Sku_name[7]
	}
	return ccstructs.Sku_name[0] // UNKNOWN
}

func getCmkMaxEcku(cluster *cmkv2.CmkV2Cluster) int32 {
	if isBasic(cluster) {
		return cluster.Spec.Config.CmkV2Basic.GetMaxEcku()
	}
	if isStandard(cluster) {
		return cluster.Spec.Config.CmkV2Standard.GetMaxEcku()
	}
	if isEnterprise(cluster) {
		return cluster.Spec.Config.CmkV2Enterprise.GetMaxEcku()
	}
	if isFreight(cluster) {
		return cluster.Spec.Config.CmkV2Freight.GetMaxEcku()
	}

	return -1
}

func getCmkClusterSize(cluster *cmkv2.CmkV2Cluster) int32 {
	if isDedicated(cluster) {
		return *cluster.Status.Cku
	}
	return -1
}

func getCmkClusterPendingSize(cluster *cmkv2.CmkV2Cluster) int32 {
	if isDedicated(cluster) {
		return cluster.Spec.Config.CmkV2Dedicated.Cku
	}
	return -1
}

func getCmkByokId(cluster *cmkv2.CmkV2Cluster) string {
	if isDedicated(cluster) && cluster.Spec.Byok != nil {
		return cluster.Spec.Byok.Id
	}
	return ""
}

func getCmkEncryptionKey(cluster *cmkv2.CmkV2Cluster) string {
	if isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.EncryptionKey != nil {
		return *cluster.Spec.Config.CmkV2Dedicated.EncryptionKey
	}
	return ""
}

func isBasic(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Basic != nil
}

func isStandard(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Standard != nil
}

func isEnterprise(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Enterprise != nil
}

func isFreight(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Freight != nil
}

func isDedicated(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Dedicated != nil
}

func isExpanding(cluster *cmkv2.CmkV2Cluster) bool {
	return isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.Cku > *cluster.Status.Cku
}

func isShrinking(cluster *cmkv2.CmkV2Cluster) bool {
	return isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.Cku < *cluster.Status.Cku
}

func getCmkClusterStatus(cluster *cmkv2.CmkV2Cluster) string {
	if isExpanding(cluster) {
		return "EXPANDING"
	}
	if isShrinking(cluster) {
		return "SHRINKING"
	}
	if cluster.Status.Phase == "PROVISIONED" {
		return "UP"
	}
	return cluster.Status.Phase
}

func topicNameStrategy(topic, mode string) string {
	return fmt.Sprintf("%s-%s", topic, mode)
}

func getLimitsForSku(cluster *cmkv2.CmkV2Cluster, usageLimits *kafkausagelimits.UsageLimits) *kafkausagelimits.Limits {
	if isDedicated(cluster) {
		return usageLimits.GetCkuLimit(cluster.Status.GetCku())
	}

	sku := getCmkClusterType(cluster)
	return usageLimits.GetTierLimit(sku).GetClusterLimits()
}
