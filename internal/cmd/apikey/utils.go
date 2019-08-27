package apikey

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
	"strings"
)

func (c *command) resolveResourceID(cmd *cobra.Command, args []string) (resourceType string, accId string, clusterId string, currentKey string, err error) {
	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return "", "", "", "", err
	}
	// If resource is schema registry
	if strings.HasPrefix(resource, "lsrc-") {
		src, err := pcmd.GetSchemaRegistry(cmd, c.ch)
		resourceType = "schema-registry"
		if err != nil {
			return "", "", "", "", err
		}
		clusterInContext, _ := c.config.SchemaRegistryCluster()
		if clusterInContext == nil || clusterInContext.SrCredentials == nil {
			currentKey = ""
		} else {
			currentKey = clusterInContext.SrCredentials.Key
		}
		return resourceType, src.AccountId, src.Id, currentKey, nil

	} else {
		kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, "resource")
		resourceType = "kafka"
		if err != nil {
			return "", "", "", "", err
		}
		return resourceType, c.config.Auth.Account.Id, kcc.ID, kcc.APIKey, nil
	}
}
