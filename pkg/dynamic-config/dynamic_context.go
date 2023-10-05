package dynamicconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	presource "github.com/confluentinc/cli/v3/pkg/resource"
)

type DynamicContext struct {
	*config.Context
	V2Client *ccloudv2.Client
}

func NewDynamicContext(context *config.Context, v2Client *ccloudv2.Client) *DynamicContext {
	if context == nil {
		return nil
	}
	return &DynamicContext{
		Context:  context,
		V2Client: v2Client,
	}
}

func (d *DynamicContext) ParseFlagsIntoContext(cmd *cobra.Command) error {
	if environment, _ := cmd.Flags().GetString("environment"); environment != "" {
		if d.GetCredentialType() == config.APIKey {
			output.ErrPrintln(d.Config.EnableColor, "WARNING: The `--environment` flag is ignored when using API key credentials.")
		} else {
			ctx := d.Config.Context()
			d.Config.SetOverwrittenCurrentEnvironment(ctx.CurrentEnvironment)
			ctx.SetCurrentEnvironment(environment)
		}
	}

	if cluster, _ := cmd.Flags().GetString("cluster"); cluster != "" {
		if d.GetCredentialType() == config.APIKey {
			output.ErrPrintln(d.Config.EnableColor, "WARNING: The `--cluster` flag is ignored when using API key credentials.")
		} else {
			ctx := d.Config.Context()
			d.Config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
			ctx.KafkaClusterContext.SetActiveKafkaCluster(cluster)
		}
	}

	if computePool, _ := cmd.Flags().GetString("compute-pool"); computePool != "" {
		ctx := d.Config.Context()
		d.Config.SetOverwrittenFlinkComputePool(ctx.GetCurrentFlinkComputePool())
		if err := ctx.SetCurrentFlinkComputePool(computePool); err != nil {
			return err
		}
	}

	if region, _ := cmd.Flags().GetString("region"); region != "" {
		ctx := d.Config.Context()
		if err := ctx.SetCurrentFlinkRegion(region); err != nil {
			return err
		}
	}

	if cloud, _ := cmd.Flags().GetString("cloud"); cloud != "" {
		ctx := d.Config.Context()
		if err := ctx.SetCurrentFlinkCloudProvider(cloud); err != nil {
			return err
		}
	}

	return nil
}

func (d *DynamicContext) GetKafkaClusterForCommand() (*config.KafkaClusterConfig, error) {
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

func (d *DynamicContext) FindKafkaCluster(clusterId string) (*config.KafkaClusterConfig, error) {
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
	environmentId, err := d.EnvironmentId()
	if err != nil {
		return nil, err
	}

	cluster, httpResp, err := d.V2Client.DescribeKafkaCluster(clusterId, environmentId)
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

func (d *DynamicContext) UseAPIKey(apiKey, clusterId string) error {
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

func (d *DynamicContext) HasLogin() bool {
	credType := d.GetCredentialType()
	switch credType {
	case config.Username:
		return d.GetAuthToken() != ""
	case config.APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, d.Name))
	}
}

func (d *DynamicContext) EnvironmentId() (string, error) {
	if id := d.GetCurrentEnvironment(); id != "" {
		return id, nil
	}

	return "", errors.NewErrorWithSuggestions("no environment found", "This issue may occur if this user has no valid role bindings. Contact an Organization Admin to create a role binding for this user.")
}

// AuthenticatedState returns the context's state if authenticated, and an error otherwise.
// A view of the state is returned, rather than a pointer to the actual state. Changing the state
// should be done by accessing the state field directly.
func (d *DynamicContext) AuthenticatedState() (*config.ContextState, error) {
	if !d.HasLogin() {
		return nil, new(errors.NotLoggedInError)
	}
	return d.State, nil
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
		return "", "", errors.NewErrorWithSuggestions(
			"no API key specified",
			"Use the `--api-key` flag to specify an API key.",
		)
	}

	return apiKey, apiSecret, nil
}
