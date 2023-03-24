package schemaregistry

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type clusterOut struct {
	Name                  string `human:"Name" serialized:"name"`
	ClusterId             string `human:"Cluster" serialized:"cluster_id"`
	EndpointUrl           string `human:"Endpoint URL" serialized:"endpoint_url"`
	UsedSchemas           string `human:"Used Schemas" serialized:"used_schemas"`
	AvailableSchemas      string `human:"Available Schemas" serialized:"available_schemas"`
	FreeSchemasLimit      int    `human:"Free Schemas Limit" serialized:"free_schemas_limit"`
	GlobalCompatibility   string `human:"Global Compatibility" serialized:"global_compatibility"`
	Mode                  string `human:"Mode" serialized:"mode"`
	ServiceProvider       string `human:"Service Provider" serialized:"service_provider"`
	ServiceProviderRegion string `human:"Service Provider Region" serialized:"service_provider_region"`
	Package               string `human:"Package" serialized:"package"`
}

const (
	defaultSchemaLimitAdvanced            = 20000
	streamGovernancePriceTableProductName = "stream-governance"
	schemaRegistryPriceTableName          = "SchemaRegistry"
)

func (c *command) newClusterDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe the Schema Registry cluster for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterDescribe,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) clusterDescribe(cmd *cobra.Command, _ []string) error {
	var compatibility string
	var mode string
	var numSchemas string
	var availableSchemas string
	var srClient *srsdk.APIClient
	ctx := context.Background()

	// Collect the parameters
	cluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, c.EnvironmentId())
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
		org := &ccloudv1.Organization{Id: c.Context.GetOrganization().GetId()}
		prices, err := c.Client.Billing.GetPriceTable(context.Background(), org, streamGovernancePriceTableProductName)
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

	table := output.NewTable(cmd)
	table.Add(&clusterOut{
		Name:                  cluster.Name,
		ClusterId:             cluster.Id,
		EndpointUrl:           cluster.Endpoint,
		ServiceProvider:       cluster.ServiceProvider,
		ServiceProviderRegion: cluster.ServiceProviderRegion,
		Package:               getPackageDisplayName(cluster.Package),
		UsedSchemas:           numSchemas,
		AvailableSchemas:      availableSchemas,
		FreeSchemasLimit:      freeSchemasLimit,
		GlobalCompatibility:   compatibility,
		Mode:                  mode,
	})
	return table.Print()
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
