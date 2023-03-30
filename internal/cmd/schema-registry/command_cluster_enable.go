package schemaregistry

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type enableOut struct {
	Id          string `human:"ID" serialized:"id"`
	EndpointUrl string `human:"Endpoint URL" serialized:"endpoint_url"`
}

var availableGeos = []string{"us", "eu", "apac"}

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
				Code: fmt.Sprintf("%s schema-registry cluster enable --cloud gcp --geo us --package advanced", version.CLIName),
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("geo", "", fmt.Sprintf("Specify the geo as %s.", utils.ArrayToCommaDelimitedString(availableGeos, "or")))
	addPackageFlag(cmd, essentialsPackage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("geo"))

	pcmd.RegisterFlagCompletionFunc(cmd, "geo", func(_ *cobra.Command, _ []string) []string { return availableGeos })

	return cmd
}

func (c *command) clusterEnable(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	// Collect the parameters
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	geo, err := cmd.Flags().GetString("geo")
	if err != nil {
		return err
	}

	// Trust the API will handle CCP/CCE
	location := ccloudv1.GlobalSchemaRegistryLocation(ccloudv1.GlobalSchemaRegistryLocation_value[strings.ToUpper(geo)])
	if err := c.validateLocation(location); err != nil {
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

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	// Build the SR instance
	clusterConfig := &ccloudv1.SchemaRegistryClusterConfig{
		AccountId:       environmentId,
		Location:        location,
		ServiceProvider: cloud,
		Package:         packageInternalName,
		// Name is a special string that everyone expects. Originally, this field was added to support
		// multiple SR instances, but for now there's a contract between our services that it will be
		// this hardcoded string constant
		Name: "account schema-registry",
	}

	var out *enableOut
	newCluster, err := c.Client.SchemaRegistry.CreateSchemaRegistryCluster(ctx, clusterConfig)
	if err != nil {
		// If it already exists, return the existing one
		existingCluster, getExistingErr := c.Context.FetchSchemaRegistryByEnvironmentId(ctx, environmentId)
		if getExistingErr != nil {
			// Propagate CreateSchemaRegistryCluster error.
			return err
		}

		out = &enableOut{
			Id:          existingCluster.Id,
			EndpointUrl: existingCluster.Endpoint,
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

func (c *command) validateLocation(location ccloudv1.GlobalSchemaRegistryLocation) error {
	if location == ccloudv1.GlobalSchemaRegistryLocation_NONE {
		return errors.NewErrorWithSuggestions(errors.InvalidSchemaRegistryLocationErrorMsg,
			errors.InvalidSchemaRegistryLocationSuggestions)
	}
	return nil
}
