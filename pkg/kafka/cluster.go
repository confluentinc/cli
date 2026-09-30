package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/resource"
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
		return nil, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterId), "You can set the active Kafka cluster with `confluent kafka cluster use`.")
	}

	cluster, httpResp, err := client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return nil, errors.CatchKafkaNotFoundError(err, clusterId, httpResp)
	}

	bootstrap := cluster.Spec.GetKafkaBootstrapEndpoint()
	if active_endpoint := ctx.KafkaClusterContext.GetActiveKafkaClusterEndpoint(); active_endpoint != "" {
		clusterConfigs, _, err := client.DescribeKafkaCluster(clusterId, ctx.GetCurrentEnvironment())
		if err != nil {
			log.CliLogger.Debugf("Error describing Kafka Cluster: %v", err)
			return nil, fmt.Errorf("error retrieving configs for cluster %q", clusterId)
		}

		clusterEndpoints := clusterConfigs.Spec.GetEndpoints()

		for _, attributes := range clusterEndpoints {
			if attributes.GetHttpEndpoint() == active_endpoint {
				bootstrap = attributes.GetKafkaBootstrapEndpoint()
				break
			}
		}
	}

	config := &config.KafkaClusterConfig{
		ID:           cluster.GetId(),
		Name:         cluster.Spec.GetDisplayName(),
		Bootstrap:    strings.TrimPrefix(bootstrap, "SASL_SSL://"),
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
