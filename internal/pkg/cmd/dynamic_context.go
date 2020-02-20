package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"
	srV1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	"github.com/mohae/deepcopy"
	"github.com/spf13/cobra"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type DynamicContext struct {
	*v2.Context
	resolver FlagResolver
	client   *ccloud.Client
}

func NewDynamicContext(context *v2.Context, resolver FlagResolver, client *ccloud.Client) *DynamicContext {
	return &DynamicContext{
		Context:  context,
		resolver: resolver,
		client:   client,
	}
}

func (d *DynamicContext) ActiveKafkaCluster(cmd *cobra.Command) (*v1.KafkaClusterConfig, error) {
	var clusterId string
	resourceType, resourceId, err := d.resolver.ResolveResourceId(cmd)
	if err != nil {
		return nil, err
	}
	if resourceType == KafkaResourceType || resourceType == KSQLResourceType {
		clusterId = resourceId
	}
	if clusterId == "" {
		// Try "cluster" flag.
		clusterId, err = d.resolver.ResolveClusterFlag(cmd)
		if err != nil {
			return nil, err
		}
		if clusterId == "" {
			// No flags provided, just retrieve the current one specified in the config.
			clusterId = d.Kafka
		}
	}
	cluster, err := d.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (d *DynamicContext) FindKafkaCluster(cmd *cobra.Command, clusterId string) (*v1.KafkaClusterConfig, error) {
	if cluster, ok := d.KafkaClusters[clusterId]; ok {
		return cluster, nil
	}
	if d.client == nil {
		return nil, errors.ErrNoKafkaContext
	}
	// Resolve cluster details if not found locally.
	ctxClient := NewContextClient(d)
	kcc, err := ctxClient.FetchCluster(cmd, clusterId)
	if err != nil {
		return nil, err
	}
	cluster := &v1.KafkaClusterConfig{
		ID:          clusterId,
		Name:        kcc.Name,
		Bootstrap:   strings.TrimPrefix(kcc.Endpoint, "SASL_SSL://"),
		APIEndpoint: kcc.ApiEndpoint,
		APIKeys:     make(map[string]*v0.APIKeyPair),
	}
	d.KafkaClusters[clusterId] = cluster
	err = d.Save()
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (d *DynamicContext) SetActiveKafkaCluster(cmd *cobra.Command, clusterId string) error {
	if _, err := d.FindKafkaCluster(cmd, clusterId); err != nil {
		return err
	}
	d.Kafka = clusterId
	return d.Save()
}

func (d *DynamicContext) UseAPIKey(cmd *cobra.Command, apiKey string, clusterId string) error {
	kcc, err := d.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return err
	}
	if _, ok := kcc.APIKeys[apiKey]; !ok {
		// Fetch API key error.
		ctxClient := NewContextClient(d)
		return ctxClient.FetchAPIKeyError(cmd, apiKey, clusterId)
	}
	kcc.APIKey = apiKey
	return d.Save()
}

// SchemaRegistryCluster returns the SchemaRegistryCluster of the Context,
// or an empty SchemaRegistryCluster if there is none set,
// or an ErrNotLoggedIn if the user is not logged in.
func (d *DynamicContext) SchemaRegistryCluster(cmd *cobra.Command) (*v2.SchemaRegistryCluster, error) {
	/*
		1. Get rsrc flag
		2a. If resourceType is SR
			3. Try to find locally by resId
			4a. If found
				5. *Done*
			4b. Else
				5. Fetch remotely. *Done*
		2b. Else
			3. Find locally by envId
			4a. If found
				5. *Done*
			4b. Else
				5. Fetch remotely *Done.
	*/
	resourceType, resourceId, err := d.resolver.ResolveResourceId(cmd)
	if err != nil {
		return nil, err
	}
	envId, err := d.AuthenticatedEnvId(cmd)
	if err != nil {
		return nil, err
	}
	ctxClient := NewContextClient(d)
	var cluster *v2.SchemaRegistryCluster
	var clusterChanged bool
	if resourceType == SrResourceType {
		for _, srCluster := range d.SchemaRegistryClusters {
			if srCluster.Id == resourceId {
				cluster = srCluster
			}
		}
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := ctxClient.FetchSchemaRegistryById(context.Background(), resourceId, envId)
			if err != nil {
				return nil, err
			}
			cluster = makeSRCluster(srCluster)
			clusterChanged = true
		}
	} else {
		cluster = d.SchemaRegistryClusters[envId]
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(context.Background(), envId)
			if err != nil {
				return nil, err
			}
			cluster = makeSRCluster(srCluster)
			clusterChanged = true
		}
	}
	d.SchemaRegistryClusters[envId] = cluster
	if clusterChanged {
		err = d.Save()
		if err != nil {
			return nil, err
		}
	}
	return cluster, nil
}

func (d *DynamicContext) HasLogin(cmd *cobra.Command) (bool, error) {
	credType := d.Credential.CredentialType
	switch credType {
	case v2.Username:
		_, err := d.resolveEnvironmentId(cmd)
		if err != nil {
			return false, err
		}
		return d.State.AuthToken != "", nil
	case v2.APIKey:
		return false, nil
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, d.Name))
	}
}

func (d *DynamicContext) AuthenticatedEnvId(cmd *cobra.Command) (string, error) {
	state, err := d.AuthenticatedState(cmd)
	if err != nil {
		return "", err
	}
	return state.Auth.Account.Id, nil
}

// AuthenticatedState returns the context's state if authenticated, and an error otherwise.
// A view of the state is returned, rather than a pointer to the actual state. Changing the state
// should be done by accessing the state field directly.
func (d *DynamicContext) AuthenticatedState(cmd *cobra.Command) (*v2.ContextState, error) {
	hasLogin, err := d.HasLogin(cmd)
	if err != nil {
		return nil, err
	}
	if !hasLogin {
		return nil, errors.ErrNotLoggedIn
	}
	envId, err := d.resolveEnvironmentId(cmd)
	if err != nil {
		return nil, err
	}
	if envId == "" {
		return d.State, nil
	}
	state := deepcopy.Copy(d.State).(*v2.ContextState)
	for _, account := range d.State.Auth.Accounts {
		if account.Id == envId {
			state.Auth.Account = account
		}
	}
	return state, nil
}

func (d *DynamicContext) HasAPIKey(cmd *cobra.Command, clusterId string) (bool, error) {
	cluster, err := d.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return false, err
	}
	return cluster.APIKey != "", nil
}

func (d *DynamicContext) CheckSchemaRegistryHasAPIKey(cmd *cobra.Command) (bool, error) {
	srCluster, err := d.SchemaRegistryCluster(cmd)
	if err != nil {
		return false, nil
	}
	return !(srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0), nil
}

func (d *DynamicContext) resolveEnvironmentId(cmd *cobra.Command) (string, error) {
	envId, err := d.resolver.ResolveEnvironmentFlag(cmd)
	if err != nil {
		return "", err
	}
	if d.State == nil || d.State.Auth == nil {
		return "", errors.ErrNotLoggedIn
	}
	if envId == "" {
		// Environment flag not set.
		if d.State.Auth.Account == nil || d.State.Auth.Account.Id == "" {
			return "", errors.ErrNotLoggedIn
		}
		return d.State.Auth.Account.Id, nil
	}
	// Environment flag is set.
	if d.State.Auth.Accounts == nil {
		return "", errors.ErrNotLoggedIn
	}
	for _, account := range d.State.Auth.Accounts {
		if account.Id == envId {
			return envId, nil
		}
	}
	return "", fmt.Errorf("environment with id '%s' not found in context '%s'", envId, d.Name)
}

func missingDetails(cluster *v2.SchemaRegistryCluster) bool {
	return cluster.SchemaRegistryEndpoint == "" || cluster.Id == ""
}

func makeSRCluster(cluster *srV1.SchemaRegistryCluster) *v2.SchemaRegistryCluster {
	return &v2.SchemaRegistryCluster{
		Id:                     cluster.Id,
		SchemaRegistryEndpoint: cluster.Endpoint,
		SrCredentials:          nil, // For now.
	}
}
