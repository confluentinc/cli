package dynamicconfig

import (
	"context"
	"fmt"
	"time"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	presource "github.com/confluentinc/cli/internal/pkg/resource"
)

type DynamicContext struct {
	*v1.Context
	Client   *ccloudv1.Client
	V2Client *ccloudv2.Client
}

func NewDynamicContext(context *v1.Context, client *ccloudv1.Client, v2Client *ccloudv2.Client) *DynamicContext {
	if context == nil {
		return nil
	}
	return &DynamicContext{
		Context:  context,
		Client:   client,
		V2Client: v2Client,
	}
}

// Parse "--environment" and "--cluster" flag values into config struct
func (d *DynamicContext) ParseFlagsIntoContext(cmd *cobra.Command, client *ccloudv1.Client) error {
	if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
		if d.Credential.CredentialType == v1.APIKey {
			return errors.New(errors.EnvironmentFlagWithApiLoginErrorMsg)
		}

		// If environment ID is not found in config, make api call and check against those accounts
		if !d.verifyEnvironmentId(environment, d.GetEnvironments()) {
			if client == nil {
				return fmt.Errorf(errors.EnvironmentNotFoundErrorMsg, environment, d.Name)
			}

			accounts, err := d.getAllEnvironments(client)
			if err != nil {
				return err
			}

			if d.verifyEnvironmentId(environment, accounts) {
				d.SetEnvironments(accounts)
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

// getAllEnvironments retrives all environments listed by ccloud v1 client.
// It also includes the audit-log environment when that's enabled
func (d *DynamicContext) getAllEnvironments(client *ccloudv1.Client) ([]*ccloudv1.Account, error) {
	environments, err := client.Account.List(context.Background(), &ccloudv1.Account{})
	if err != nil {
		return environments, err
	}

	if d.State.Auth == nil || d.State.Auth.Organization == nil || d.State.Auth.Organization.GetAuditLog() == nil || d.State.Auth.Organization.AuditLog.ServiceAccountId == 0 {
		return environments, nil
	}
	auditLogAccountId := d.State.Auth.Organization.GetAuditLog().GetAccountId()
	auditLogEnvironment, err := client.Account.Get(context.Background(), &ccloudv1.Account{Id: auditLogAccountId})
	return append(environments, auditLogEnvironment), err
}

func (d *DynamicContext) verifyEnvironmentId(envId string, environments []*ccloudv1.Account) bool {
	for _, env := range environments {
		if env.Id == envId {
			d.Config.SetOverwrittenAccount(d.GetEnvironment())
			d.State.Auth.Account = env
			return true
		}
	}
	return false
}

func (d *DynamicContext) GetKafkaClusterForCommand() (*v1.KafkaClusterConfig, error) {
	if d.KafkaClusterContext == nil {
		return nil, errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	clusterId := d.KafkaClusterContext.GetActiveKafkaClusterId()
	if clusterId == "" {
		return nil, errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaSelectedSuggestions)
	}

	cluster, err := d.FindKafkaCluster(clusterId)
	if presource.LookupType(clusterId) != presource.KafkaCluster && clusterId != "anonymous-id" {
		return nil, errors.Errorf(errors.KafkaClusterMissingPrefixErrorMsg, clusterId)
	}
	return cluster, errors.CatchKafkaNotFoundError(err, clusterId, nil)
}

func (d *DynamicContext) FindKafkaCluster(clusterId string) (*v1.KafkaClusterConfig, error) {
	if config := d.KafkaClusterContext.GetKafkaClusterConfig(clusterId); config != nil && config.Bootstrap != "" {
		if clusterId == "anonymous-id" {
			return config, nil
		}
		const week = 7 * 24 * time.Hour
		if time.Now().Before(config.LastUpdate.Add(week)) {
			return config, nil
		}
	}

	// Don't attempt to fetch cluster details if the client isn't initialized/authenticated yet
	if d.V2Client == nil {
		return nil, nil
	}

	// Resolve cluster details if not found locally.
	config, err := d.FetchCluster(clusterId)
	if err != nil {
		return nil, err
	}

	d.KafkaClusterContext.AddKafkaClusterConfig(config)
	err = d.Save()

	return config, err
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
		return d.FetchAPIKeyError(apiKey, clusterId)
	}
	kcc.APIKey = apiKey
	return d.Save()
}

// SchemaRegistryCluster returns the SchemaRegistryCluster of the Context,
// or an empty SchemaRegistryCluster if there is none set,
// or an ErrNotLoggedIn if the user is not logged in.
func (d *DynamicContext) SchemaRegistryCluster(cmd *cobra.Command) (*v1.SchemaRegistryCluster, error) {
	resource, _ := cmd.Flags().GetString("resource")
	resourceType := presource.LookupType(resource)

	envId, err := d.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	var cluster *v1.SchemaRegistryCluster
	var clusterChanged bool
	if resourceType == presource.SchemaRegistryCluster {
		for _, srCluster := range d.SchemaRegistryClusters {
			if srCluster.Id == resource {
				cluster = srCluster
			}
		}
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := d.FetchSchemaRegistryById(context.Background(), resource, envId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resource)
			}
			cluster = makeSRCluster(srCluster)
			clusterChanged = true
		}
	} else {
		cluster = d.SchemaRegistryClusters[envId]
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := d.FetchSchemaRegistryByEnvironmentId(context.Background(), envId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resource)
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

func (d *DynamicContext) HasLogin() bool {
	credType := d.Credential.CredentialType
	switch credType {
	case v1.Username:
		return d.GetAuthToken() != ""
	case v1.APIKey:
		return false
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
	if !d.HasLogin() {
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
	if cmd.Flag("api-key") == nil || cmd.Flag("api-secret") == nil {
		return "", "", nil
	}
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return "", "", err
	}

	apiSecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return "", "", err
	}

	if apiKey == "" && apiSecret != "" {
		return "", "", errors.NewErrorWithSuggestions(errors.PassedSecretButNotKeyErrorMsg, errors.PassedSecretButNotKeySuggestions)
	}

	return apiKey, apiSecret, nil
}

func missingDetails(cluster *v1.SchemaRegistryCluster) bool {
	return cluster.SchemaRegistryEndpoint == "" || cluster.Id == ""
}

func makeSRCluster(cluster *ccloudv1.SchemaRegistryCluster) *v1.SchemaRegistryCluster {
	return &v1.SchemaRegistryCluster{
		Id:                     cluster.GetId(),
		SchemaRegistryEndpoint: cluster.GetEndpoint(),
		SrCredentials:          nil, // For now.
	}
}
