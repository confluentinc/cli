package apikey

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	kafkaResourceType = "kafka"
	srResourceType    = "schema-registry"
)

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
