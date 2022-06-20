package schemaregistry

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
		RunE:        c.enable,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Enable Schema Registry, using "aws" in "us-east-1" region with "advanced" package for environment "env-12345"`,
				Code: fmt.Sprintf("%s schema-registry cluster enable --cloud aws --region us-east-1 --package advanced --environment env-12345", version.CLIName),
			},
		),
	}

	pcmd.AddStreamGovernancePackageFlag(cmd, getAllPackageDisplayNames())
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", `Cloud region name`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("package")
	_ = cmd.MarkFlagRequired("region")

	return cmd
}

func (c *clusterCommand) enable(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	// Collect the parameters
	serviceProvider, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	serviceProviderRegion, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	clouds, err := c.Client.EnvironmentMetadata.Get(ctx)
	if err != nil {
		return err
	}

	if err := checkServiceProviderAndRegion(serviceProvider, serviceProviderRegion, clouds); err != nil {
		return err
	}

	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	packageInternalName, isValid := getPackageInternalName(packageDisplayName)
	if !isValid {
		return errors.New(fmt.Sprintf(errors.SRInvalidPackageType, packageDisplayName))
	}

	// Build the SR instance
	clusterConfig := &schedv1.SchemaRegistryClusterConfig{
		AccountId:             c.EnvironmentId(),
		ServiceProvider:       serviceProvider,
		ServiceProviderRegion: serviceProviderRegion,
		Package:               packageInternalName,
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
	}

	return nil
}

func checkServiceProviderAndRegion(cloudId string, regionId string, clouds []*schedv1.CloudMetadata) error {
	for _, cloud := range clouds {
		if cloudId == cloud.Id {
			for _, region := range cloud.Regions {
				if regionId == region.Id {
					return nil
				}
			}
			return errors.New(fmt.Sprintf(errors.CloudRegionNotAvailableErrorMsg, regionId, cloudId))
		}
	}
	return errors.New(fmt.Sprintf(errors.CloudProviderNotAvailableErrorMsg, cloudId))
}
