package kafka

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	logger "log"
	_nethttp "net/http"
	"os"
	"regexp"
	"strings"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func copyMap(inputMap map[string]string) map[string]string {
	newMap := make(map[string]string)
	for key, val := range inputMap {
		newMap[key] = val
	}
	return newMap
}

func toCreateTopicConfigs(topicConfigsMap map[string]string) []kafkarestv3.ConfigData {
	topicConfigs := make([]kafkarestv3.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = kafkarestv3.ConfigData{
			Name:  k,
			Value: &val,
		}
		i++
	}
	return topicConfigs
}

func toAlterConfigBatchRequestData(configsMap map[string]string) []kafkarestv3.AlterConfigBatchRequestDataData {
	kafkaRestConfigs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for k, v := range configsMap {
		val := v
		kafkaRestConfigs[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:      k,
			Value:     &val,
			Operation: nil,
		}
		i++
	}
	return kafkaRestConfigs
}

func getKafkaClusterLkcId(c *pcmd.AuthenticatedStateFlagCommand) (string, error) {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return "", err
	}
	return kafkaClusterConfig.ID, nil
}

func createTestConfigFile(name string, configs map[string]string) (string, error) {
	dir, _ := os.Getwd()
	logger.Println("Test config file dir:", dir)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return dir, err
	}

	write := bufio.NewWriter(file)
	for key, val := range configs {
		if _, err = write.WriteString(key + "=" + val + "\n"); err != nil {
			file.Close()
			return dir, err
		}
	}

	if err = write.Flush(); err != nil {
		return dir, err
	}

	return dir, file.Close()
}

func parseProducerConfigFile(path string) (*producerConfigs, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	configs := &producerConfigs{}
	configBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(configBytes, configs)

	return configs, err
}

func handleOpenApiError(httpResp *_nethttp.Response, err error, client *kafkarestv3.APIClient) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkaRestError(client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func getKafkaRestProxyAndLkcId(c *pcmd.AuthenticatedStateFlagCommand) (*pcmd.KafkaREST, string, error) {
	kafkaREST, err := c.AuthenticatedCLICommand.GetKafkaREST()
	if err != nil {
		return nil, "", err
	}
	if kafkaREST == nil {
		return nil, "", errors.New(errors.RestProxyNotAvailable)
	}
	// Kafka REST is available
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, "", err
	}
	return kafkaREST, kafkaClusterConfig.ID, nil
}

func isClusterResizeInProgress(currentCluster *cmkv2.CmkV2Cluster) error {
	if currentCluster.Status.Phase == "PROVISIONING" {
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

func getCmkClusterIngressAndEgress(cluster *cmkv2.CmkV2Cluster) (int32, int32) {
	if isDedicated(cluster) {
		return 50 * (*cluster.Status.Cku), 150 * (*cluster.Status.Cku)
	}
	return 100, 100
}

func getCmkClusterType(cluster *cmkv2.CmkV2Cluster) string {
	if isBasic(cluster) {
		return productv1.Sku_name[2]
	}
	if isStandard(cluster) {
		return productv1.Sku_name[3]
	}
	return productv1.Sku_name[4]
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

func camelToSnake(camels []string) []string {
	var ret []string
	for _, camel := range camels {
		snake := matchFirstCap.ReplaceAllString(camel, "${1}_${2}")
		snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
		ret = append(ret, strings.ToLower(snake))
	}
	return ret
}

func camelToSpaced(camels []string) []string {
	var ret []string
	for _, camel := range camels {
		snake := matchFirstCap.ReplaceAllString(camel, "${1} ${2}")
		snake = matchAllCap.ReplaceAllString(snake, "${1} ${2}")
		ret = append(ret, snake)
	}
	return ret
}

func topicNameStrategy(topic string) string {
	return topic + "-value"
}
