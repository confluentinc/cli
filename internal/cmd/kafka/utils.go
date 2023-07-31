package kafka

import (
	_nethttp "net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
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

func toAlterConfigBatchRequestData(configsMap map[string]string) cckafkarestv3.AlterConfigBatchRequestData {
	kafkaRestConfigs := make([]cckafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		v := val
		kafkaRestConfigs[i] = cckafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: *cckafkarestv3.NewNullableString(&v),
		}
		i++
	}
	return cckafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}
}

func toAlterConfigBatchRequestDataOnPrem(configsMap map[string]string) cpkafkarestv3.AlterConfigBatchRequestData {
	kafkaRestConfigs := make([]cpkafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		v := val
		kafkaRestConfigs[i] = cpkafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: &v,
		}
		i++
	}
	return cpkafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}
}

func handleOpenApiError(httpResp *_nethttp.Response, err error, client *cpkafkarestv3.APIClient) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkarest.NewError(client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func getKafkaRestProxyAndLkcId(c *pcmd.AuthenticatedCLICommand) (*pcmd.KafkaREST, string, error) {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, "", err
	}
	if kafkaREST == nil {
		return nil, "", errors.New(errors.RestProxyNotAvailable)
	}
	// Kafka REST is available
	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, "", err
	}
	return kafkaREST, kafkaClusterConfig.ID, nil
}

func isClusterResizeInProgress(currentCluster *cmkv2.CmkV2Cluster) error {
	if currentCluster.Status.Phase == ccloudv2.StatusProvisioning {
		return errors.New(errors.KafkaClusterStillProvisioningErrorMsg)
	}
	if isExpanding(currentCluster) {
		return errors.New(errors.KafkaClusterExpandingErrorMsg)
	}
	if isShrinking(currentCluster) {
		return errors.New(errors.KafkaClusterShrinkingErrorMsg)
	}
	return nil
}

func getCmkClusterIngressAndEgressMbps(cluster *cmkv2.CmkV2Cluster) (int32, int32) {
	if isDedicated(cluster) {
		return 50 * cluster.Status.GetCku(), 150 * cluster.Status.GetCku()
	}
	return 250, 750
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
	return ccstructs.Sku_name[0] // UNKNOWN
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

func topicNameStrategy(topic string) string {
	return topic + "-value"
}
