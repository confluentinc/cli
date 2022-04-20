package schemaregistry

import (
	"context"
	"math"
	"strconv"

	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeLabels            = []string{"Name", "ID", "URL", "Used", "Available", "Compatibility", "Mode", "ServiceProvider"}
	describeHumanRenames      = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider"}
	describeStructuredRenames = map[string]string{"Name": "name", "ID": "cluster_id", "URL": "endpoint_url", "Used": "used_schemas", "Available": "available_schemas", "Compatibility": "global_compatibility", "Mode": "mode", "ServiceProvider": "service_provider"}
)

type describeDisplay struct {
	Name            string
	ID              string
	URL             string
	Used            string
	Available       string
	Compatibility   string
	Mode            string
	ServiceProvider string
}

func (c *clusterCommand) newDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the Schema Registry cluster for this environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.describe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, _ []string) error {
	var compatibility string
	var mode string
	var numSchemas string
	var availableSchemas string
	var srClient *srsdk.APIClient
	ctx := context.Background()

	// Collect the parameters
	ctxClient := dynamicconfig.NewContextClient(c.Context)
	cluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}
	// Retrieve SR compatibility and Mode if API key is set up in user's config.json file
	srClusterHasAPIKey, err := c.Context.CheckSchemaRegistryHasAPIKey(cmd)
	if err != nil {
		return err
	}
	if srClusterHasAPIKey {
		srClient, ctx, err = GetApiClient(cmd, c.srClient, c.Config, c.Version)
		if err != nil {
			return err
		}
		// Get SR compatibility
		compatibilityResponse, _, err := srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			compatibility = ""
			log.CliLogger.Warn("Could not retrieve Schema Registry Compatibility")
		} else {
			compatibility = compatibilityResponse.CompatibilityLevel
		}
		// Get SR Mode
		modeResponse, _, err := srClient.DefaultApi.GetTopLevelMode(ctx)
		if err != nil {
			mode = ""
			log.CliLogger.Warn("Could not retrieve Schema Registry Mode")
		} else {
			mode = modeResponse.Mode
		}
	} else {
		srClient = nil
		compatibility = "<Requires API Key>"
		mode = "<Requires API Key>"
	}

	query := schemaCountQueryFor(cluster.Id)
	metricsResponse, err := c.Client.MetricsApi.QueryV2(ctx, "cloud", query, "")
	if err != nil || metricsResponse == nil {
		log.CliLogger.Warn("Could not retrieve Schema Registry Metrics: ", err)
		numSchemas = ""
		availableSchemas = ""
	} else if len(metricsResponse.Result) == 0 {
		numSchemas = "0"
		availableSchemas = strconv.Itoa(int(cluster.MaxSchemas))
	} else if len(metricsResponse.Result) == 1 {
		numSchemasInt := int(math.Round(metricsResponse.Result[0].Value)) // the return value is a double
		numSchemas = strconv.Itoa(numSchemasInt)
		availableSchemas = strconv.Itoa(int(cluster.MaxSchemas) - numSchemasInt)
	} else {
		log.CliLogger.Warn("Unexpected results from Metrics API")
		numSchemas = ""
		availableSchemas = ""
	}

	serviceProvider := getServiceProviderFromUrl(cluster.Endpoint)
	data := &describeDisplay{
		Name:            cluster.Name,
		ID:              cluster.Id,
		URL:             cluster.Endpoint,
		ServiceProvider: serviceProvider,
		Used:            numSchemas,
		Available:       availableSchemas,
		Compatibility:   compatibility,
		Mode:            mode,
	}
	return output.DescribeObject(cmd, data, describeLabels, describeHumanRenames, describeStructuredRenames)
}

func schemaCountQueryFor(schemaRegistryId string) *ccloud.MetricsApiRequest {
	return &ccloud.MetricsApiRequest{
		Aggregations: []ccloud.ApiAggregation{
			{
				Metric: "io.confluent.kafka.schema_registry/schema_count",
			},
		},
		Filter: ccloud.ApiFilter{
			Field: "resource.schema_registry.id",
			Op:    "EQ",
			Value: schemaRegistryId,
		},
		Granularity: "ALL",
		Intervals:   []string{"PT1M/now-2m|m"},
	}
}
