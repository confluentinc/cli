package schemaregistry

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

var (
	describeLabels            = []string{"Name", "ID", "URL", "Used", "Available", "Compatibility", "Mode", "ServiceProvider"}
	describeHumanRenames      = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider"}
	describeStructuredRenames = map[string]string{"Name": "name", "ID": "cluster_id", "URL": "endpoint_url", "Used": "used_schemas", "Available": "available_schemas", "Compatibility": "global_compatibility", "Mode": "mode", "ServiceProvider": "service_provider"}
	enableLabels              = []string{"Id", "SchemaRegistryEndpoint"}
	enableHumanRenames        = map[string]string{"ID": "Cluster ID", "SchemaRegistryEndpoint": "Endpoint URL"}
	enableStructuredRenames   = map[string]string{"ID": "cluster_id", "SchemaRegistryEndpoint": "endpoint_url"}
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	logger          *log.Logger
	srClient        *srsdk.APIClient
	analyticsClient analytics.Client
}

func NewClusterCommand(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger, analyticsClient analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Schema Registry cluster.",
		}, prerunner, ClusterSubcommandFlags)
	clusterCmd := &clusterCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		srClient:                      srClient,
		logger:                        logger,
		analyticsClient:               analyticsClient,
	}
	clusterCmd.init(cliName)
	return clusterCmd.Command
}

func (c *clusterCommand) init(cliName string) {
	createCmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable Schema Registry for this environment.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.enable),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Enable Schema Registry, using Google Cloud Platform in the US:",
				Code: fmt.Sprintf("%s schema-registry cluster enable --cloud gcp --geo us", cliName),
			},
		),
	}
	createCmd.Flags().String("cloud", "", "Cloud provider (e.g. 'aws', 'azure', or 'gcp').")
	_ = createCmd.MarkFlagRequired("cloud")
	createCmd.Flags().String("geo", "", "Either 'us', 'eu', or 'apac'.")
	_ = createCmd.MarkFlagRequired("geo")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the Schema Registry cluster for this environment.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update global mode or compatibility of Schema Registry.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update top level compatibility or mode of Schema Registry.",
				Code: fmt.Sprintf("%s schema-registry cluster update --compatibility=BACKWARD\n%s schema-registry cluster update --mode=READWRITE", cliName, cliName),
			},
		),
	}
	updateCmd.Flags().String("compatibility", "", "Can be BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, or NONE.")
	updateCmd.Flags().String("mode", "", "Can be READWRITE, READ, OR WRITE.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)
}

func (c *clusterCommand) enable(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	// Collect the parameters
	serviceProvider, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	locationFlag, err := cmd.Flags().GetString("geo")
	if err != nil {
		return err
	}

	// Trust the API will handle CCP/CCE
	location := schedv1.GlobalSchemaRegistryLocation(schedv1.GlobalSchemaRegistryLocation_value[strings.ToUpper(locationFlag)])

	// Build the SR instance
	clusterConfig := &schedv1.SchemaRegistryClusterConfig{
		AccountId:       c.EnvironmentId(),
		Location:        location,
		ServiceProvider: serviceProvider,
		// Name is a special string that everyone expects. Originally, this field was added to support
		// multiple SR instances, but for now there's a contract between our services that it will be
		// this hardcoded string constant
		Name: "account schema-registry",
	}
	newCluster, err := c.Client.SchemaRegistry.CreateSchemaRegistryCluster(ctx, clusterConfig)
	if err != nil {
		// If it already exists, return the existing one
		cluster, getExistingErr := c.Context.SchemaRegistryCluster(cmd)
		if getExistingErr != nil {
			// Propagate CreateSchemaRegistryCluster error.
			return err
		}
		_ = output.DescribeObject(cmd, cluster, enableLabels, enableHumanRenames, enableStructuredRenames)
	} else {
		v2Cluster := &v2.SchemaRegistryCluster{
			Id:                     newCluster.Id,
			SchemaRegistryEndpoint: newCluster.Endpoint,
		}
		_ = output.DescribeObject(cmd, v2Cluster, enableLabels, enableHumanRenames, enableStructuredRenames)
		c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, v2Cluster.Id)
	}
	return nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, _ []string) error {
	var compatibility string
	var mode string
	var numSchemas string
	var availableSchemas string
	var srClient *srsdk.APIClient
	ctx := context.Background()

	// Collect the parameters
	ctxClient := pcmd.NewContextClient(c.Context)
	cluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
	if err != nil {
		return err
	}
	//Retrieve SR compatibility and Mode if API key is set up in user's config.json file
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
			c.logger.Warn("Could not retrieve Schema Registry Compatibility")
		} else {
			compatibility = compatibilityResponse.CompatibilityLevel
		}
		// Get SR Mode
		modeResponse, _, err := srClient.DefaultApi.GetTopLevelMode(ctx)
		if err != nil {
			mode = ""
			c.logger.Warn("Could not retrieve Schema Registry Mode")
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
		c.logger.Warn("Could not retrieve Schema Registry Metrics: ", err)
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
		c.logger.Warn("Unexpected results from Metrics API")
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

func (c *clusterCommand) update(cmd *cobra.Command, _ []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compat != "" {
		return c.updateCompatibility(cmd)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateMode(cmd)
	}
	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *clusterCommand) updateCompatibility(cmd *cobra.Command) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	updateReq := srsdk.ConfigUpdateRequest{Compatibility: strings.ToUpper(compat)}
	_, _, err = srClient.DefaultApi.UpdateTopLevelConfig(ctx, updateReq)
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.UpdatedToLevelCompatibilityMsg, updateReq.Compatibility)
	return nil
}

func (c *clusterCommand) updateMode(cmd *cobra.Command) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	modeUpdate, _, err := srClient.DefaultApi.UpdateTopLevelMode(ctx, srsdk.ModeUpdateRequest{Mode: strings.ToUpper(mode)})
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.UpdatedTopLevelModeMsg, modeUpdate.Mode)
	return nil
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
