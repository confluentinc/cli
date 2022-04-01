package kafka

import (
	"bufio"
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	logger "log"
	_nethttp "net/http"
	"os"
	"regexp"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func toCPCreateTopicConfigs(topicConfigsMap map[string]string) []kafkarestv3.ConfigData {
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

func toCloudCreateTopicConfigs(topicConfigsMap map[string]string) *[]cloudkafkarest.ConfigData {
	topicConfigs := make([]cloudkafkarest.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = cloudkafkarest.ConfigData{
			Name:  k,
			Value: *cloudkafkarest.NewNullableString(&val),
		}
		i++
	}
	return &topicConfigs
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

func toCloudAlterConfigBatchRequestData(configsMap map[string]string) []cloudkafkarest.AlterConfigBatchRequestDataData {
	kafkaRestConfigs := make([]cloudkafkarest.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for k, v := range configsMap {
		val := v
		kafkaRestConfigs[i] = cloudkafkarest.AlterConfigBatchRequestDataData{
			Name:      k,
			Value:     *cloudkafkarest.NewNullableString(&val),
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

func handleOpenApiError(httpResp *_nethttp.Response, err error, client *kafkarestv3.APIClient) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkaRestError(client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func getKafkaRestProxyAndLkcId(c *pcmd.AuthenticatedStateFlagCommand) (*pcmd.CloudKafkaREST, string, error) {
	kafkaREST, err := c.AuthenticatedCLICommand.GetCloudKafkaREST()
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

func isClusterResizeInProgress(currentCluster *schedv1.KafkaCluster) error {
	switch currentCluster.Status {
	case schedv1.ClusterStatus_PROVISIONING:
		return errors.New(errors.KafkaClusterStillProvisioningErrorMsg)
	case schedv1.ClusterStatus_EXPANDING:
		return errors.New(errors.KafkaClusterExpandingErrorMsg)
	case schedv1.ClusterStatus_SHRINKING:
		return errors.New(errors.KafkaClusterShrinkingErrorMsg)
	case schedv1.ClusterStatus_DELETING:
		return errors.New(errors.KafkaClusterDeletingErrorMsg)
	case schedv1.ClusterStatus_DELETED:
		return errors.New(errors.KafkaClusterDeletingErrorMsg)
	}
	return nil
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
