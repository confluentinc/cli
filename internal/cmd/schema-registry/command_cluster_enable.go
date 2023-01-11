package schemaregistry

import (
	"context"
	"fmt"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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

func (c *clusterCommand) newEnableCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "enable",
		Short:       "Enable Schema Registry for this environment.",
		Args:        cobra.NoArgs,
		RunE:        c.enable,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Enable Schema Registry, using Google Cloud Platform in the US with the "advanced" package.`,
				Code: fmt.Sprintf("%s schema-registry cluster enable --cloud gcp --geo us --package advanced", version.CLIName),
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("geo", "", fmt.Sprintf("Specify the geo as %s.", utils.ArrayToCommaDelimitedString(availableGeos)))
	addPackageFlag(cmd, essentialsPackage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("geo")

	pcmd.RegisterFlagCompletionFunc(cmd, "geo", func(_ *cobra.Command, _ []string) []string { return availableGeos })

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
	location := ccloudv1.GlobalSchemaRegistryLocation(ccloudv1.GlobalSchemaRegistryLocation_value[strings.ToUpper(locationFlag)])
	err = c.validateLocation(location)
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

	// Build the SR instance
	clusterConfig := &ccloudv1.SchemaRegistryClusterConfig{
		AccountId:       c.EnvironmentId(),
		Location:        location,
		ServiceProvider: serviceProvider,
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
		existingCluster, getExistingErr := c.Context.FetchSchemaRegistryByAccountId(ctx, c.EnvironmentId())
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

func (c *clusterCommand) validateLocation(location ccloudv1.GlobalSchemaRegistryLocation) error {
	if location == ccloudv1.GlobalSchemaRegistryLocation_NONE {
		return errors.NewErrorWithSuggestions(errors.InvalidSchemaRegistryLocationErrorMsg,
			errors.InvalidSchemaRegistryLocationSuggestions)
	}
	return nil
}
