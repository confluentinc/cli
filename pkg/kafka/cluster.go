package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func GetClusterForCommand(client *ccloudv2.Client, ctx *config.Context) (*config.KafkaClusterConfig, error) {
	if ctx.KafkaClusterContext == nil {
		return nil, errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	clusterId := ctx.KafkaClusterContext.GetActiveKafkaClusterId()
	if clusterId == "" {
		return nil, errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	if resource.LookupType(clusterId) != resource.KafkaCluster && clusterId != "anonymous-id" {
		return nil, fmt.Errorf(errors.KafkaClusterMissingPrefixErrorMsg, clusterId)
	}

	cluster, err := FindCluster(client, ctx, clusterId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, nil)
	}

	return cluster, nil
}

func FindCluster(client *ccloudv2.Client, ctx *config.Context, clusterId string) (*config.KafkaClusterConfig, error) {
	if config := ctx.KafkaClusterContext.GetKafkaClusterConfig(clusterId); config != nil && config.Bootstrap != "" {
		if clusterId == "anonymous-id" {
			return config, nil
		}
		const week = 7 * 24 * time.Hour
		if time.Now().Before(config.LastUpdate.Add(week)) {
			return config, nil
		}
	}

	// Resolve cluster details if not found locally.
	environmentId, err := ctx.EnvironmentId()
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, fmt.Errorf(errors.KafkaClusterNotFoundErrorMsg, clusterId)
	}

	cluster, httpResp, err := client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, httpResp)
	}

	config := &config.KafkaClusterConfig{
		ID:           cluster.GetId(),
		Name:         cluster.Spec.GetDisplayName(),
		Bootstrap:    strings.TrimPrefix(cluster.Spec.GetKafkaBootstrapEndpoint(), "SASL_SSL://"),
		RestEndpoint: cluster.Spec.GetHttpEndpoint(),
		APIKeys:      make(map[string]*config.APIKeyPair),
		LastUpdate:   time.Now(),
	}

	ctx.KafkaClusterContext.AddKafkaClusterConfig(config)
	if err := ctx.Save(); err != nil {
		return nil, err
	}

	return config, nil
}
