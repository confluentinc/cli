package apikey

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	kafkaResourceType = "kafka"
	srResourceType    = "schema-registry"
)

//func (c *command) resolveResourceID(cmd *cobra.Command, args []string) (resourceType string, accId string, clusterId string, currentKey string, err error) {
//	resource, err := cmd.Flags().GetString("resource")
//	if err != nil {
//		return "", "", "", "", err
//	}
//	// If resource is schema registry
//	if strings.HasPrefix(resource, "lsrc-") {
//		src, err := pcmd.GetSchemaRegistry(cmd, c.ch)
//		if err != nil {
//			return "", "", "", "", err
//		}
//		if src == nil {
//			return "", "", "", "", errors.ErrNoSrEnabled
//		}
//		clusterInContext, _ := c.config.SchemaRegistryCluster()
//		if clusterInContext == nil || clusterInContext.SrCredentials == nil {
//			currentKey = ""
//		} else {
//			currentKey = clusterInContext.SrCredentials.Key
//		}
//		return srResourceType, src.AccountId, src.Id, currentKey, nil
//
//	} else {
//		kcc, err := pcmd.GetKafkaClusterConfig(cmd, c.ch, "resource")
//		if err != nil {
//			return "", "", "", "", err
//		}
//		state, err := c.config.AuthenticatedState()
//		if err != nil {
//			return "", "", "", "", err
//		}
//		return kafkaResourceType, state.Auth.Account.Id, kcc.ID, kcc.APIKey, nil
//	}
//}

func resolveResourceId(cfg *config.Config) (resourceType string, accId string, clusterId string, currentKey string, err error) {
	resolutionError := func(err error) (string, string, string, string, error) {
		return "", "", "", "", err
	}
	ctx := cfg.Context()
	state, err := ctx.AuthenticatedState()
	if err != nil {
		return resolutionError(err)
	}
	if ctx.UserSpecifiedSchemaRegistryEnvId != "" {
		resourceType = srResourceType
		cluster, err := cfg.SchemaRegistryCluster()
		if err != nil {
			return resolutionError(err)
		}
		if cluster == nil {
			return resolutionError(errors.ErrNoSrEnabled)
		}
		clusterId = cluster.Id
		if cluster.SrCredentials != nil {
			currentKey = cluster.SrCredentials.Key
		}
	} else {
		resourceType = kafkaResourceType
		cluster, err := ctx.ActiveKafkaCluster()
		if err != nil {
			return resolutionError(err)
		}
		clusterId = cluster.ID
		currentKey = cluster.APIKey
	}
	return resourceType, state.Auth.Account.Id, clusterId, currentKey, nil

}
