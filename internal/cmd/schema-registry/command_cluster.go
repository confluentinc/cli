package schema_registry

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	"github.com/confluentinc/go-printer"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
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
	describeLabels  = []string{"Name", "ID", "URL", "Used", "Available", "Compatibility", "Mode", "ServiceProvider"}
	describeRenames = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider"}
	enableLabels    = []string{"Id", "Endpoint"}
	enableRenames   = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL"}
)

type clusterCommand struct {
	*pcmd.AuthenticatedCLICommand
	logger   *log.Logger
	srClient *srsdk.APIClient
}

func NewClusterCommand(config *v2.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Schema Registry cluster.",
		},
		config, prerunner)
	clusterCmd := &clusterCommand{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
		logger:                  logger,
	}
	clusterCmd.init()
	return clusterCmd.Command
}

func (c *clusterCommand) init() {
	createCmd := &cobra.Command{
		Use:     "enable",
		Short:   `Enable Schema Registry for this environment.`,
		Example: FormatDescription(`{{.CLIName}} schema-registry cluster enable --cloud gcp --geo us`, c.Config.CLIName),
		RunE:    c.enable,
		Args:    cobra.NoArgs,
	}
	createCmd.Flags().String("cloud", "", "Cloud provider (e.g. 'aws', 'azure', or 'gcp').")
	_ = createCmd.MarkFlagRequired("cloud")
	createCmd.Flags().String("geo", "", "Either 'us', 'eu', or 'apac'.")
	_ = createCmd.MarkFlagRequired("geo")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
	describeCmd := &cobra.Command{
		Use:     "describe",
		Short:   `Describe the Schema Registry cluster for this environment.`,
		Example: FormatDescription(`{{.CLIName}} schema-registry cluster describe`, c.Config.CLIName),
		RunE:    c.describe,
		Args:    cobra.NoArgs,
	}
	c.AddCommand(describeCmd)
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: `Update global mode or compatibility of Schema Registry.`,
		Example: FormatDescription(`Update top level compatibility or mode of schema registry.

::
		{{.CLIName}} schema-registry cluster update <subjectname> --compatibility=BACKWARD
		{{.CLIName}} schema-registry cluster update <subjectname> --mode=READWRITE`, c.Config.CLIName),
		RunE: c.update,
		Args: cobra.NoArgs,
	}
	updateCmd.Flags().String("compatibility", "", "Can be BACKWARD, BACKWARD_TRANSITIVE, FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE, or NONE.")
	updateCmd.Flags().String("mode", "", "Can be READWRITE, READ, OR WRITE.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)
}

func (c *clusterCommand) enable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	// Collect the parameters
	serviceProvider, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	locationFlag, err := cmd.Flags().GetString("geo")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// Trust the API will handle CCP/CCE
	location := srv1.GlobalSchemaRegistryLocation(srv1.GlobalSchemaRegistryLocation_value[strings.ToUpper(locationFlag)])

	// Build the SR instance
	clusterConfig := &srv1.SchemaRegistryClusterConfig{
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
			return errors.HandleCommon(err, cmd)
		}
		_ = printer.RenderTableOut(cluster, enableLabels, enableRenames, os.Stdout)
	} else {
		_ = printer.RenderTableOut(newCluster, enableLabels, enableRenames, os.Stdout)
	}
	return nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
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
		return errors.HandleCommon(err, cmd)
	}
	//Retrieve SR compatibility and Mode if API key is set up in user's config.json file
	srClusterHasAPIKey, err := c.Context.CheckSchemaRegistryHasAPIKey(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
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

	// Get Schema usage metrics
	metrics, err := c.Client.Metrics.SchemaRegistryMetrics(ctx, cluster.Id)
	if err != nil {
		c.logger.Warn("Could not retrieve Schema Registry Metrics")
		numSchemas = ""
		availableSchemas = ""
	} else {
		numSchemas = strconv.Itoa(int(metrics.NumSchemas))
		availableSchemas = strconv.Itoa(int(cluster.MaxSchemas) - int(metrics.NumSchemas))
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
	_ = printer.RenderTableOut(data, describeLabels, describeRenames, os.Stdout)
	return nil
}
func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if compat != "" {
		return c.updateCompatibility(cmd, args)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if mode != "" {
		return c.updateMode(cmd, args)
	}
	return errors.New("flag --compatibility or --mode is required.")
}
func (c *clusterCommand) updateCompatibility(cmd *cobra.Command, args []string) error {
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
	pcmd.Printf(cmd, "Successfully updated Top Level compatibilty: %s \n", updateReq.Compatibility)
	return nil
}

func (c *clusterCommand) updateMode(cmd *cobra.Command, args []string) error {
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
	pcmd.Printf(cmd, "Successfully updated Top Level mode: %s \n", modeUpdate.Mode)
	return nil
}
