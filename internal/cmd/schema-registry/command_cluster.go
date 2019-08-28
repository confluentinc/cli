package schema_registry

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	ccsdk "github.com/confluentinc/ccloud-sdk-go"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
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

type clusterCommand struct {
	*cobra.Command
	config       *config.Config
	ccClient     ccsdk.SchemaRegistry
	metricClient ccsdk.Metrics
	srClient     *srsdk.APIClient
	ch           *pcmd.ConfigHelper
}

func NewClusterCommand(config *config.Config, ccloudClient ccsdk.SchemaRegistry, ch *pcmd.ConfigHelper, srClient *srsdk.APIClient, metricClient ccsdk.Metrics) *cobra.Command {
	clusterCmd := &clusterCommand{
		Command: &cobra.Command{
			Use:   "cluster",
			Short: "Manage Schema Registry cluster.",
		},
		config:       config,
		ccClient:     ccloudClient,
		ch:           ch,
		srClient:     srClient,
		metricClient: metricClient,
	}
	clusterCmd.init()
	return clusterCmd.Command
}

func (c *clusterCommand) init() {
	createCmd := &cobra.Command{
		Use:     "enable",
		Short:   `Enable Schema Registry for this account.`,
		Example: `ccloud schema-registry cluster enable --cloud gcp --geo us`,
		RunE:    c.enable,
		Args:    cobra.NoArgs,
	}
	createCmd.Flags().String("cloud", "", "Cloud provider (e.g. 'aws', 'azure', or 'gcp').")
	_ = createCmd.MarkFlagRequired("cloud")
	createCmd.Flags().String("geo", "", "Either 'us', 'eu', or 'apac'.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
	describeCmd := &cobra.Command{
		Use:     "describe",
		Short:   `Describe an instance of Schema Registry.`,
		Example: `ccloud schema-registry cluster describe`,
		RunE:    c.describe,
		Args:    cobra.NoArgs,
	}
	c.AddCommand(describeCmd)
	
}

func (c *clusterCommand) enable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Collect the parameters
	accountId, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	serviceProvider, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	locationFlag, err := cmd.Flags().GetString("geo")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// Trust the API will handle CCP/CCE and whether geo is required
	location := srv1.GlobalSchemaRegistryLocation(srv1.GlobalSchemaRegistryLocation_value[strings.ToUpper(locationFlag)])

	// Build the SR instance
	clusterConfig := &srv1.SchemaRegistryClusterConfig{
		AccountId:       accountId,
		Location:        location,
		ServiceProvider: serviceProvider,
		// Name is a special string that everyone expects. Originally, this field was added to support
		// multiple SR instances, but for now there's a contract between our services that it will be
		// this hardcoded string constant
		Name: "account schema-registry",
	}

	newCluster, err := c.ccClient.CreateSchemaRegistryCluster(ctx, clusterConfig)
	if err != nil {
		// If it already exists, return the existing one
		cluster, getExistingErr := GetSchemaRegistryByAccountId(c.ccClient, ctx, accountId)
		if getExistingErr != nil {
			return errors.HandleCommon(err, cmd)
		}
		pcmd.Println(cmd, "Schema Registry already enabled:")
		pcmd.Println(cmd, "Cluster ID: "+cluster.Id)
		pcmd.Println(cmd, "Endpoint: "+cluster.Endpoint)
		return nil
	}
	pcmd.Println(cmd, "Schema Registry enabled:")
	pcmd.Println(cmd, "Cluster ID: "+newCluster.Id)
	pcmd.Println(cmd, "Endpoint: "+newCluster.Endpoint)
	return nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {

	var compatibility string
	var mode string
	var numSchemas string
	var availableSchemas string
	ctx := context.Background()
	fields := []string{"Name", "ID", "URL", "Used", "Available", "Compatibility", "Mode", "ServiceProvider"}
	renames := map[string]string{"ID": "Logical Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider"}

	// Collect the parameters
	accountId, err := pcmd.GetEnvironment(cmd, c.config)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cluster, err := GetSchemaRegistryByAccountId(c.ccClient, ctx, accountId)

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	srClient, ctx, err := GetApiClient(c.srClient, c.ch)
	if err != nil {
		return err
	}

	// Get Schema usage metrics
	metrics, err := c.metricClient.SchemaRegistryMetrics(ctx, cluster.Id)
	if err != nil {
		numSchemas = ""
		availableSchemas = ""
	} else {
		numSchemas = strconv.Itoa(int(metrics.NumSchemas))
		availableSchemas = strconv.Itoa(int(cluster.MaxSchemas) - int(metrics.NumSchemas))
	}
	// Get SR compatibility
	compatibilityResponse, _, err := srClient.DefaultApi.GetTopLevelConfig(ctx)
	if err != nil {
		compatibility = ""
	} else {
		compatibility = compatibilityResponse.CompatibilityLevel
	}
	// Get SR Mode
	ModeResponse, _, err := srClient.DefaultApi.GetTopLevelMode(ctx)
	if err != nil {
		mode = ""
	} else {
		mode = ModeResponse.Mode
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
	_ = printer.RenderTableOut(data, fields, renames, os.Stdout)
	return nil
}
