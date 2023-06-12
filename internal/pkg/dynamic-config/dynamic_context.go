package dynamicconfig

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

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

func (d *DynamicContext) ParseFlagsIntoContext(cmd *cobra.Command, client *ccloudv1.Client) error {
	if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
		if d.Credential.CredentialType == v1.APIKey {
			return errors.New("`--environment` flag should not be passed for API key context")
		}
		ctx := d.Config.Context()
		d.Config.SetOverwrittenCurrentEnvironment(ctx.CurrentEnvironment)
		ctx.SetCurrentEnvironment(environment)
	}

	if cluster, _ := cmd.Flags().GetString("cluster"); cluster != "" {
		if d.Credential.CredentialType == v1.APIKey {
			return errors.New("`--cluster` flag should not be passed for API key context, cluster is inferred")
		}
		ctx := d.Config.Context()
		d.Config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
		ctx.KafkaClusterContext.SetActiveKafkaCluster(cluster)
	}

	if computePool, _ := cmd.Flags().GetString("compute-pool"); computePool != "" {
		ctx := d.Config.Context()
		d.Config.SetOverwrittenFlinkComputePool(ctx.GetCurrentFlinkComputePool())
		if err := ctx.SetCurrentFlinkComputePool(computePool); err != nil {
			return err
		}
	}

	return nil
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

	environmentId, err := d.EnvironmentId()
	if err != nil {
		return nil, err
	}

	var cluster *v1.SchemaRegistryCluster
	var clusterChanged bool
	if resourceType == presource.SchemaRegistryCluster {
		for _, srCluster := range d.SchemaRegistryClusters {
			if srCluster != nil && srCluster.GetId() == resource {
				cluster = srCluster
			}
		}
		if cluster == nil || missingDetails(cluster) {
			srCluster, err := d.V2Client.GetSchemaRegistryClusterById(resource, environmentId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resource)
			}
			cluster = makeSRCluster(&srCluster)
			clusterChanged = true
		}
	} else {
		cluster = d.SchemaRegistryClusters[environmentId]
		if cluster == nil || missingDetails(cluster) {
			srClusters, err := d.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
			if err != nil {
				return nil, errors.CatchResourceNotFoundError(err, resource)
			}
			if len(srClusters) != 0 {
				cluster = makeSRCluster(&srClusters[0])
			} else {
				cluster = nil
			}
			clusterChanged = true
		}
	}
	d.SchemaRegistryClusters[environmentId] = cluster
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

func (d *DynamicContext) EnvironmentId() (string, error) {
	if id := d.GetCurrentEnvironment(); id != "" {
		return id, nil
	}

	return "", errors.NewErrorWithSuggestions(errors.NoEnvironmentFoundErrorMsg, errors.NoEnvironmentFoundSuggestions)
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

func makeSRCluster(cluster *srcmv2.SrcmV2Cluster) *v1.SchemaRegistryCluster {
	clusterSpec := cluster.GetSpec()
	return &v1.SchemaRegistryCluster{
		Id:                     cluster.GetId(),
		SchemaRegistryEndpoint: clusterSpec.GetHttpEndpoint(),
		SrCredentials:          nil, // For now.
	}
}
