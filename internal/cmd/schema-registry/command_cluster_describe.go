package schemaregistry

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/log"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	describeLabels       = []string{"Name", "ID", "URL", "Used", "Available", "FreeSchemasLimit", "Compatibility", "Mode", "ServiceProvider", "ServiceProviderRegion", "Package"}
	describeHumanRenames = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "FreeSchemasLimit": "Free Schemas Limit",
		"Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider", "ServiceProviderRegion": "Service Provider Region"}
	describeStructuredRenames = map[string]string{"Name": "name", "ID": "cluster_id", "URL": "endpoint_url", "Used": "used_schemas", "Available": "available_schemas", "FreeSchemasLimit": "free_schemas_limit",
		"Compatibility": "global_compatibility", "Mode": "mode", "ServiceProvider": "service_provider", "ServiceProviderRegion": "service_provider_region", "Package": "package"}
)

const (
	defaultSchemaLimitAdvanced            = 20000
	streamGovernancePriceTableProductName = "stream-governance"
	schemaRegistryPriceTableName          = "SchemaRegistry"
)

type describeDisplay struct {
	Name                  string
	ID                    string
	URL                   string
	Used                  string
	Available             string
	FreeSchemasLimit      string
	Compatibility         string
	Mode                  string
	ServiceProvider       string
	ServiceProviderRegion string
	Package               string
}

func (c *clusterCommand) newDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the Schema Registry cluster for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.describe,
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
	cluster, err := c.Context.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}
	// Retrieve SR compatibility and Mode if API key is set up in user's config.json file
	srClusterHasAPIKey, err := c.Context.CheckSchemaRegistryHasAPIKey(cmd)
	if err != nil {
		return err
	}
	if srClusterHasAPIKey {
		srClient, ctx, err = getApiClient(cmd, c.srClient, c.Config, c.Version)
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
	metricsResponse, httpResp, err := c.V2Client.MetricsDatasetQuery("cloud", query)
	unmarshalErr := ccloudv2.UnmarshalFlatQueryResponseIfDataSchemaMatchError(err, metricsResponse, httpResp)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	freeSchemasLimit := int(cluster.MaxSchemas)
	if cluster.Package == essentialsPackageInternal {
		org := &orgv1.Organization{Id: c.State.Auth.Organization.Id}
		prices, err := c.PrivateClient.Billing.GetPriceTable(context.Background(), org, streamGovernancePriceTableProductName)

		if err == nil {
			priceKey := getMaxSchemaLimitPriceKey(cluster.Package, cluster.ServiceProvider, cluster.ServiceProviderRegion)
			freeSchemasLimit = int(prices.GetPriceTable()[schemaRegistryPriceTableName].Prices[priceKey])
		}
	} else {
		freeSchemasLimit = defaultSchemaLimitAdvanced
	}

	if err != nil && !ccloudv2.IsDataMatchesMoreThanOneSchemaError(err) || metricsResponse == nil {
		log.CliLogger.Warn("Could not retrieve Schema Registry Metrics: ", err)
		numSchemas = ""
		availableSchemas = ""
	} else if len(metricsResponse.FlatQueryResponse.GetData()) == 0 {
		numSchemas = "0"
		availableSchemas = strconv.Itoa(freeSchemasLimit)
	} else if len(metricsResponse.FlatQueryResponse.GetData()) == 1 {
		numSchemasInt := int(math.Round(float64(metricsResponse.FlatQueryResponse.GetData()[0].Value))) // the return value is a float32
		numSchemas = strconv.Itoa(numSchemasInt)
		// Available number of schemas should not be negative
		availableSchemas = strconv.Itoa(int(math.Max(float64(freeSchemasLimit-numSchemasInt), 0)))
	} else {
		log.CliLogger.Warn("Unexpected results from Metrics API")
		numSchemas = ""
		availableSchemas = ""
	}

	data := &describeDisplay{
		Name:                  cluster.Name,
		ID:                    cluster.Id,
		URL:                   cluster.Endpoint,
		ServiceProvider:       cluster.ServiceProvider,
		ServiceProviderRegion: cluster.ServiceProviderRegion,
		Package:               getPackageDisplayName(cluster.Package),
		Used:                  numSchemas,
		Available:             availableSchemas,
		FreeSchemasLimit:      strconv.Itoa(freeSchemasLimit),
		Compatibility:         compatibility,
		Mode:                  mode,
	}
	return output.DescribeObject(cmd, data, describeLabels, describeHumanRenames, describeStructuredRenames)
}

func schemaCountQueryFor(schemaRegistryId string) metricsv2.QueryRequest {
	aggregations := []metricsv2.Aggregation{
		{
			Metric: "io.confluent.kafka.schema_registry/schema_count",
		},
	}
	filter := metricsv2.Filter{
		FieldFilter: &metricsv2.FieldFilter{
			Field: metricsv2.PtrString("resource.schema_registry.id"),
			Op:    "EQ",
			Value: metricsv2.StringAsFieldFilterValue(metricsv2.PtrString(schemaRegistryId)),
		},
	}
	req := metricsv2.NewQueryRequest(aggregations, "ALL", []string{"PT1M/now-2m|m"})
	req.SetFilter(filter)
	return *req
}

func getMaxSchemaLimitPriceKey(sgPackage, serviceProvider, serviceProviderRegion string) string {
	return fmt.Sprintf("%s:%s:%s:1:max", serviceProvider, serviceProviderRegion, sgPackage)
}
