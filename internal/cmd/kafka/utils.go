package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

func getKafkaClusterLkcId(c *pcmd.AuthenticatedStateFlagCommand, cmd *cobra.Command) (string, error) {
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return "", err
	}
	return kafkaClusterConfig.ID, nil
}
