package apikey

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *command) resolveResourceId(cmd *cobra.Command, resolver pcmd.FlagResolver, client *ccloud.Client) (resourceType string, clusterId string, currentKey string, err error) {
	resourceType, resourceId, err := resolver.ResolveResourceId(cmd)
	if err != nil || resourceType == "" {
		return "", "", "", err
	}
	if resourceType == pcmd.SrResourceType {
		cluster, err := c.Context.SchemaRegistryCluster(cmd)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			currentKey = cluster.SrCredentials.Key
		}
	} else if resourceType == pcmd.KSQLResourceType {
		ctx := context.Background()
		cluster, err := client.KSQL.Describe(
			ctx, &schedv1.KSQLCluster{
				Id:        resourceId,
				AccountId: c.EnvironmentId(),
			})
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.Id
	} else if resourceType == pcmd.CloudResourceType {
		return resourceType, "", "", nil
	} else {
		// Resource is of KafkaResourceType.
		cluster, err := c.Context.FindKafkaCluster(cmd, resourceId)
		if err != nil {
			return "", "", "", errors.CatchResourceNotFoundError(err, resourceId)
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, clusterId, currentKey, nil
}
