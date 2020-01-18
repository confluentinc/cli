package apikey

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) resolveResourceId(cmd *cobra.Command, resolver pcmd.FlagResolver) (resourceType string, clusterId string, currentKey string, err error) {
	resourceType, _, err = resolver.ResolveResourceId(cmd)
	if resourceType == pcmd.SrResourceType {
		cluster, err := c.Context.SchemaRegistryCluster(cmd)
		if err != nil {
			return "", "", "", err
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			currentKey = cluster.SrCredentials.Key
		}
	} else {
		resourceType = pcmd.KafkaResourceType
		cluster, err := c.Context.ActiveKafkaCluster(cmd)
		if err != nil {
			return "", "", "", err
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, clusterId, currentKey, nil
}
