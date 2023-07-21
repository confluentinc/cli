package schemaregistry

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
	defaultSchemaLimitEssentials          = 1000
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

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	// Deprecated
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

	// Deprecated
	pcmd.AddApiSecretFlag(cmd)
	cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))

	return cmd
}

func (c *command) clusterDescribe(cmd *cobra.Command, _ []string) error {
	var numSchemas string
	var availableSchemas string

	// Collect the parameters
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return errors.NewSRNotEnabledError()
	}
	cluster := clusters[0]

	region, err := c.V2Client.GetStreamGovernanceRegionById(cluster.Spec.Region.GetId())
	if err != nil {
		return err
	}

	client, err := c.GetSchemaRegistryClient()
	if err != nil {
		return err
	}

	config, err := client.GetTopLevelConfig()
	if err != nil {
		return err
	}

	mode, err := client.GetTopLevelMode()
	if err != nil {
		return err
	}

	metricsClient, err := c.GetMetricsClient()
	if err != nil {
		return err
	}

	metricsResponse, httpResp, err := metricsClient.MetricsDatasetQuery("cloud", schemaCountQueryFor(cluster.GetId()))
	if err := ccloudv2.UnmarshalFlatQueryResponseIfDataSchemaMatchError(err, metricsResponse, httpResp); err != nil {
		return err
	}

	freeSchemasLimit := defaultSchemaLimitEssentials
	if strings.ToLower(cluster.Spec.GetPackage()) == essentialsPackage {
		user, err := c.Client.Auth.User()
		if err != nil {
			return err
		}
		prices, err := c.Client.Billing.GetPriceTable(user.GetOrganization(), streamGovernancePriceTableProductName)
		if err == nil {
			internalPackageName, err := getPackageInternalName(cluster.Spec.GetPackage())
			if err != nil {
				return err
			}
			priceKey := getMaxSchemaLimitPriceKey(region.Spec.GetCloud(), region.Spec.GetRegionName(), internalPackageName)
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
		Name:                  cluster.Spec.GetDisplayName(),
		ClusterId:             cluster.GetId(),
		EndpointUrl:           cluster.Spec.GetHttpEndpoint(),
		ServiceProvider:       region.Spec.GetCloud(),
		ServiceProviderRegion: region.Spec.GetRegionName(),
		Package:               cluster.Spec.GetPackage(),
		UsedSchemas:           numSchemas,
		AvailableSchemas:      availableSchemas,
		FreeSchemasLimit:      freeSchemasLimit,
		GlobalCompatibility:   config.CompatibilityLevel,
		Mode:                  mode.Mode,
	})
	return table.Print()
}

func schemaCountQueryFor(schemaRegistryId string) metricsv2.QueryRequest {
	aggregations := []metricsv2.Aggregation{{Metric: "io.confluent.kafka.schema_registry/schema_count"}}
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

func getMaxSchemaLimitPriceKey(serviceProvider, serviceProviderRegion, streamGovernancePackage string) string {
	return fmt.Sprintf("%s:%s:%s:1:max", strings.ToLower(serviceProvider), serviceProviderRegion, streamGovernancePackage)
}
