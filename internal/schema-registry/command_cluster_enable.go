package schemaregistry

import (
	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type enableOut struct {
	Id          string `human:"ID" serialized:"id"`
	EndpointUrl string `human:"Endpoint URL" serialized:"endpoint_url"`
}

func (c *command) newClusterEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "enable",
		Short:       "Enable Schema Registry for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterEnable,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Enable Schema Registry, using Google Cloud Platform in the US with the "advanced" package.`,
				Code: "confluent schema-registry cluster enable --cloud gcp --region us-central1 --package advanced",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagKafka(cmd, c.AuthenticatedCLICommand)
	addPackageFlag(cmd, essentialsPackage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) clusterEnable(cmd *cobra.Command, _ []string) error {
	// Collect the parameters
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	packageDisplayName, err := cmd.Flags().GetString("package")
	if err != nil {
		return err
	}

	packageInternalName, err := getPackageInternalName(packageDisplayName)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Build the SR instance
	clusterConfig := &ccloudv1.SchemaRegistryClusterConfig{
		AccountId:             environmentId,
		ServiceProvider:       cloud,
		ServiceProviderRegion: region,
		Package:               packageInternalName,
		// Name is a special string that everyone expects. Originally, this field was added to support
		// multiple SR instances, but for now there's a contract between our services that it will be
		// this hardcoded string constant
		Name: "account schema-registry",
	}

	var out *enableOut
	newCluster, err := c.Client.SchemaRegistry.CreateSchemaRegistryCluster(clusterConfig)
	if err != nil {
		existingClusters, getExistingErr := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
		if getExistingErr != nil {
			return err
		}
		if len(existingClusters) == 0 {
			return err
		}

		out = &enableOut{
			Id:          existingClusters[0].GetId(),
			EndpointUrl: existingClusters[0].Spec.GetHttpEndpoint(),
		}
	} else {
		out = &enableOut{
			Id:          newCluster.GetId(),
			EndpointUrl: newCluster.GetEndpoint(),
		}
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
