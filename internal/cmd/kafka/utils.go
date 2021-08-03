package kafka

import (
	"bufio"
	"fmt"
	"io/ioutil"
	logger "log"
	_nethttp "net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

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

func toMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf(errors.ConfigurationFormErrorMsg)
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}

func toCreateTopicConfigs(topicConfigsMap map[string]string) []kafkarestv3.CreateTopicRequestDataConfigs {
	topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
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

func readConfigsFromFile(configFile string) (map[string]string, error) {
	if configFile == "" {
		return map[string]string{}, nil
	}

	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Create config map from the argument.a
	var configs []string
	for _, s := range strings.Split(string(configContents), "\n") {
		// Filter out blank lines
		spaceTrimmed := strings.TrimSpace(s)
		if s != "" && spaceTrimmed[0] != '#' {
			configs = append(configs, spaceTrimmed)
		}
	}

	return toMap(configs)
}

func getKafkaClusterLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (string, error) {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
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

func handleOpenApiError(httpResp *_nethttp.Response, err error, kafkaREST *pcmd.KafkaREST) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func getKafkaRestProxyAndLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (*pcmd.KafkaREST, string, error) {
	kafkaREST, err := c.AuthenticatedCLICommand.GetKafkaREST()
	if err != nil {
		return nil, "", err
	}
	if kafkaREST == nil {
		return nil, "", errors.New(errors.RestProxyNotAvailable)
	}
	// Kafka REST is available
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return nil, "", err
	}
	return kafkaREST, kafkaClusterConfig.ID, nil
}

func isClusterResizeInProgress(currentCluster *schedv1.KafkaCluster) error {
	if currentCluster.Status == schedv1.ClusterStatus_PROVISIONING {
		return errors.New(errors.KafkaClusterStillProvisioningErrorMsg)
	} else if currentCluster.Status == schedv1.ClusterStatus_EXPANDING {
		return errors.New(errors.KafkaClusterExpandingErrorMsg)
	} else if currentCluster.Status == schedv1.ClusterStatus_SHRINKING {
		return errors.New(errors.KafkaClusterShrinkingErrorMsg)
	} else if currentCluster.Status == schedv1.ClusterStatus_DELETING || currentCluster.Status == schedv1.ClusterStatus_DELETED {
		return errors.New(errors.KafkaClusterDeletingErrorMsg)
	}
	return nil
}

func camelToSnake(camels []string) []string {
	var ret []string
	for _, camel := range camels {
		snake := matchFirstCap.ReplaceAllString(camel, "${1}_${2}")
		snake  = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
		ret = append(ret, strings.ToLower(snake))
	}
	return ret
}

func camelToSpaced(camels []string) []string {
	var ret []string
	for _, camel := range camels {
		snake := matchFirstCap.ReplaceAllString(camel, "${1} ${2}")
		snake  = matchAllCap.ReplaceAllString(snake, "${1} ${2}")
		ret = append(ret, snake)
	}
	return ret
}
