package cmd

import (
	"context"
	"fmt"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

type DynamicContext struct {
	*v1.Context
	resolver FlagResolver
	client   *ccloud.Client
	v2Client *ccloudv2.Client
}

func NewDynamicContext(context *v1.Context, resolver FlagResolver, client *ccloud.Client, v2Client *ccloudv2.Client) *DynamicContext {
	return &DynamicContext{
		Context:  context,
		resolver: resolver,
		client:   client,
		v2Client: v2Client,
	}
}

// Parse "--environment" and "--cluster" flag values into config struct
func (d *DynamicContext) ParseFlagsIntoContext(cmd *cobra.Command, client *ccloud.Client) error {
	if d.resolver == nil {
		return nil
	}

	if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
		if d.Credential.CredentialType == v1.APIKey {
			return errors.New(errors.EnvironmentFlagWithApiLoginErrorMsg)
		}

		// If environment ID is not found in config, make api call and check against those accounts
		if !d.verifyEnvironmentId(environment, d.State.Auth.Accounts) {
			if client == nil {
				return fmt.Errorf(errors.EnvironmentNotFoundErrorMsg, environment, d.Name)
			}

			accounts, err := client.Account.List(context.Background(), &orgv1.Account{})
			if err != nil {
				return err
			}

			if d.verifyEnvironmentId(environment, accounts) {
				d.State.Auth.Accounts = accounts
				_ = d.Save()
			} else {
				return fmt.Errorf(errors.EnvironmentNotFoundErrorMsg, environment, d.Name)
			}
		}
	}

	if cluster, _ := cmd.Flags().GetString("cluster"); cluster != "" {
		if d.Credential.CredentialType == v1.APIKey {
			return errors.New(errors.ClusterFlagWithApiLoginErrorMsg)
		}
		ctx := d.Config.Context()
		d.Config.SetOverwrittenActiveKafka(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		ctx.KafkaClusterContext.SetActiveKafkaCluster(cluster)
	}

	return nil
}

func (d *DynamicContext) verifyEnvironmentId(envId string, environments []*orgv1.Account) bool {
	for _, env := range environments {
		if env.Id == envId {
			d.Config.SetOverwrittenAccount(d.State.Auth.Account)
			d.State.Auth.Account = env
			return true
		}
	}
	return false
}

func (d *DynamicContext) GetKafkaClusterForCommand() (*v1.KafkaClusterConfig, error) {
	clusterId := d.KafkaClusterContext.GetActiveKafkaClusterId()
	if clusterId == "" {
		return nil, errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	cluster, err := d.FindKafkaCluster(clusterId)
	return cluster, errors.CatchKafkaNotFoundError(err, clusterId)
}

func (d *DynamicContext) FindKafkaCluster(clusterId string) (*v1.KafkaClusterConfig, error) {
	if cluster := d.KafkaClusterContext.GetKafkaClusterConfig(clusterId); cluster != nil {
		return cluster, nil
	}

	if d.client == nil {
		return nil, errors.Errorf(errors.FindKafkaNoClientErrorMsg, clusterId)
	}

	// Resolve cluster details if not found locally.
	kcc, err := NewContextClient(d).FetchCluster(clusterId)
	if err != nil {
		return nil, err
	}

	cluster := kafkaClusterToKafkaClusterConfig(kcc, make(map[string]*v1.APIKeyPair))
	d.KafkaClusterContext.AddKafkaClusterConfig(cluster)
	err = d.Save()

	return cluster, err
}

func kafkaClusterToKafkaClusterConfig(kcc *schedv1.KafkaCluster, apiKeys map[string]*v1.APIKeyPair) *v1.KafkaClusterConfig {
	return &v1.KafkaClusterConfig{
		ID:           kcc.Id,
		Name:         kcc.Name,
		Bootstrap:    strings.TrimPrefix(kcc.Endpoint, "SASL_SSL://"),
		APIEndpoint:  kcc.ApiEndpoint,
		APIKeys:      apiKeys,
		RestEndpoint: kcc.RestEndpoint,
	}
}

func (d *DynamicContext) SetActiveKafkaCluster(clusterId string) error {
	if _, err := d.FindKafkaCluster(clusterId); err != nil {
		return err
	}
	d.KafkaClusterContext.SetActiveKafkaCluster(clusterId)
	return d.Save()
}

func (d *DynamicContext) RemoveKafkaClusterConfig(clusterId string) error {
	d.KafkaClusterContext.RemoveKafkaCluster(clusterId)
	return d.Save()
}

func (d *DynamicContext) UseAPIKey(apiKey string, clusterId string) error {
	kcc, err := d.FindKafkaCluster(clusterId)
	if err != nil {
		return err
	}
	if _, ok := kcc.APIKeys[apiKey]; !ok {
		// Fetch API key error.
		ctxClient := NewContextClient(d)
		return ctxClient.FetchAPIKeyError(apiKey, clusterId)
	}
	kcc.APIKey = apiKey
	return d.Save()
}

// SchemaRegistryCluster returns the SchemaRegistryCluster of the Context,
// or an empty SchemaRegistryCluster if there is none set,
// or an ErrNotLoggedIn if the user is not logged in.
func (d *DynamicContext) SchemaRegistryCluster(cmd *cobra.Command) (*v1.SchemaRegistryCluster, error) {
	resourceId, _ := cmd.Flags().GetString("resource")
	resourceType := resource.LookupType(resourceId)

	envId, err := d.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	ctxClient := NewContextClient(d)
	var cluster *v1.SchemaRegistryCluster
	var clusterChanged bool
	if resourceType == resource.SchemaRegistry {
		for _, srCluster := range d.SchemaRegistryClusters {
			if srCluster.Id == resourceId {
				cluster = srCluster
			}
		}
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := ctxClient.FetchSchemaRegistryById(context.Background(), resourceId, envId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resourceId)
			}
			cluster = makeSRCluster(srCluster)
			clusterChanged = true
		}
	} else {
		cluster = d.SchemaRegistryClusters[envId]
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(context.Background(), envId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resourceId)
			}
			cluster = makeSRCluster(srCluster)
			clusterChanged = true
		}
	}
	d.SchemaRegistryClusters[envId] = cluster
	if clusterChanged {
		if err := d.Save(); err != nil {
			return nil, err
		}
	}
	return cluster, nil
}

func (d *DynamicContext) HasLogin() (bool, error) {
	credType := d.Credential.CredentialType
	switch credType {
	case v1.Username:
		_, err := d.resolveEnvironmentId()
		if err != nil {
			return false, err
		}
		return d.State.AuthToken != "", nil
	case v1.APIKey:
		return false, nil
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, d.Name))
	}
}

func (d *DynamicContext) AuthenticatedEnvId() (string, error) {
	state, err := d.AuthenticatedState()
	if err != nil {
		return "", err
	}
	return state.Auth.Account.Id, nil
}

// AuthenticatedState returns the context's state if authenticated, and an error otherwise.
// A view of the state is returned, rather than a pointer to the actual state. Changing the state
// should be done by accessing the state field directly.
func (d *DynamicContext) AuthenticatedState() (*v1.ContextState, error) {
	hasLogin, err := d.HasLogin()
	if err != nil {
		return nil, err
	}
	if !hasLogin {
		return nil, new(errors.NotLoggedInError)
	}
	return d.State, nil
}

func (d *DynamicContext) HasAPIKey(clusterId string) (bool, error) {
	cluster, err := d.FindKafkaCluster(clusterId)
	return cluster.APIKey != "", err
}

func (d *DynamicContext) CheckSchemaRegistryHasAPIKey(cmd *cobra.Command) (bool, error) {
	srCluster, err := d.SchemaRegistryCluster(cmd)
	if err != nil {
		return false, nil
	}
	key, secret, err := d.KeyAndSecretFlags(cmd)
	if err != nil {
		return false, err
	}
	if key != "" {
		if srCluster.SrCredentials == nil {
			srCluster.SrCredentials = &v1.APIKeyPair{}
		}
		srCluster.SrCredentials.Key = key
	}
	if secret != "" {
		srCluster.SrCredentials.Secret = secret
	}
	return !(srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0), nil
}

func (d *DynamicContext) KeyAndSecretFlags(cmd *cobra.Command) (string, string, error) {
	key, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return "", "", err
	}

	secret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return "", "", err
	}

	if key == "" && secret != "" {
		return "", "", errors.NewErrorWithSuggestions(errors.PassedSecretButNotKeyErrorMsg, errors.PassedSecretButNotKeySuggestions)
	}

	return key, secret, nil
}

func (d *DynamicContext) resolveEnvironmentId() (string, error) {
	if d.State == nil || d.State.Auth == nil {
		return "", new(errors.NotLoggedInError)
	}
	if d.State.Auth.Account == nil || d.State.Auth.Account.Id == "" {
		return "", new(errors.NotLoggedInError)
	}
	return d.State.Auth.Account.Id, nil
}

func missingDetails(cluster *v1.SchemaRegistryCluster) bool {
	return cluster.SchemaRegistryEndpoint == "" || cluster.Id == ""
}

func makeSRCluster(cluster *schedv1.SchemaRegistryCluster) *v1.SchemaRegistryCluster {
	return &v1.SchemaRegistryCluster{
		Id:                     cluster.Id,
		SchemaRegistryEndpoint: cluster.Endpoint,
		SrCredentials:          nil, // For now.
	}
}
