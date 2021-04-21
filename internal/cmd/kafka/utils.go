package kafka

import (
	"bufio"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"io/ioutil"
	logger "log"
	_nethttp "net/http"
	"os"
	"strings"
)



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
	var linkConfigs []string
	for _, s := range strings.Split(string(configContents), "\n") {
		// Filter out blank lines
		if s != "" {
			linkConfigs = append(linkConfigs, s)
		}
	}

	return toMap(linkConfigs)
}

func getKafkaClusterLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (string, error) {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return "", err
	}
	return kafkaClusterConfig.ID, nil
}

func createTestConfigFile(name string, configs map[string]string) (string, error) {
	dir,_ := os.Getwd()
	logger.Println("Test config file dir:",dir)
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

	if httpResp != nil{
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return err
}
