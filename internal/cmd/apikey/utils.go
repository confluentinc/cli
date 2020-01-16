package apikey

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func resolveResourceId(cmd *pcmd.AuthenticatedCLICommand, resolver pcmd.FlagResolver) (resourceType string, accId string, clusterId string, currentKey string, err error) {
	accId = cmd.EnvironmentId()
	resourceType, _, err = resolver.ResolveResourceId(cmd.Command)
	if resourceType == pcmd.SrResourceType {
		cluster, err := cmd.Context.SchemaRegistryCluster(cmd.Command)
		if err != nil {
			return "", "", "", "", err
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			currentKey = cluster.SrCredentials.Key
		}
	} else {
		cluster, err := cmd.Context.ActiveKafkaCluster(cmd.Command)
		if err != nil {
			return "", "", "", "", err
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, accId, clusterId, currentKey, nil
}
