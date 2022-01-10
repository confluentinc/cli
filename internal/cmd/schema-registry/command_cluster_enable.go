package schemaregistry

import (
	"context"
	"fmt"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	enableLabels            = []string{"Id", "SchemaRegistryEndpoint"}
	enableHumanRenames      = map[string]string{"ID": "Cluster ID", "SchemaRegistryEndpoint": "Endpoint URL"}
	enableStructuredRenames = map[string]string{"ID": "cluster_id", "SchemaRegistryEndpoint": "endpoint_url"}
)

func (c *clusterCommand) newEnableCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "enable",
		Short:       "Enable Schema Registry for this environment.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.enable),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Enable Schema Registry, using Google Cloud Platform in the US.",
				Code: fmt.Sprintf("%s schema-registry cluster enable --cloud gcp --geo us", version.CLIName),
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("geo", "", "Either 'us', 'eu', or 'apac'.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("geo")

	pcmd.RegisterFlagCompletionFunc(cmd, "geo", func(_ *cobra.Command, _ []string) []string { return []string{"apac", "eu", "us"} })

	return cmd
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
		v2Cluster := &v1.SchemaRegistryCluster{
			Id:                     newCluster.Id,
			SchemaRegistryEndpoint: newCluster.Endpoint,
		}
		_ = output.DescribeObject(cmd, v2Cluster, enableLabels, enableHumanRenames, enableStructuredRenames)
		c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, v2Cluster.Id)
	}

	return nil
}
