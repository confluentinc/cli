package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"os"
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

func getKafkaRestProxyAndLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (*pcmd.KafkaREST, string, error) {
	kafkaREST, err := c.AuthenticatedCLICommand.GetKafkaREST()
	if err != nil {
		return nil, "", err
	}
	if kafkaREST == nil {
		return nil, "", errors.New(errors.RestProxyNotAvailable)
	}
	// Kafka REST is available
	lkc, err := getKafkaClusterLkcId(c, cmd)
	if err != nil {
		return nil, "", err
	}
	return kafkaREST, lkc, nil
}

func getKafkaClusterLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (string, error) {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return "", err
	}
	return kafkaClusterConfig.ID, nil
}
