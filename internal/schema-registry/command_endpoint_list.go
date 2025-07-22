package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
)

type listEndpoint struct {
	Public          string            `human:"Public Endpoint URL" serialized:"public_endpoint_url"`
	Private         string            `human:"Private Endpoint URL" serialized:"private_endpoint_url"`
	PrivateRegional map[string]string `human:"Private Regional Endpoint URLs" serialized:"private_regional_endpoint_urls"`
	Catalog         string            `human:"Catalog Endpoint URL" serialized:"catalog_endpoint_url"`
}

func (c *command) newEndpointsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all schema registry endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.endpointList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environmentId)
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return schemaregistry.ErrNotEnabled
	}
	cluster := clusters[0]

	privateRegionalEndpoints := cluster.Spec.PrivateNetworkingConfig.GetRegionalEndpoints()
	if privateRegionalEndpoints == nil {
		privateRegionalEndpoints = make(map[string]string)
	}

	// Note the region has to be empty slice instead of `nil` in case of no filter
	awsNetworks, err := c.V2Client.ListNetworks(environmentId, nil, []string{cloud.Aws}, []string{}, nil, []string{"READY"}, nil)
	if err != nil {
		return fmt.Errorf("unable to list Schema Registry endpoints: failed to list AWS networks: %w", err)
	}
	// Filter out non-PrivateLink networks for Azure
	azureNetworks, err := c.V2Client.ListNetworks(environmentId, nil, []string{cloud.Azure}, []string{}, nil, []string{"READY"}, []string{"PRIVATELINK"})
	if err != nil {
		return fmt.Errorf("unable to list Schema Registry endpoints: failed to list Azure PrivateLink networks: %w", err)
	}

	networks := append(awsNetworks, azureNetworks...)
	for _, network := range networks {
		suffix := network.Status.GetEndpointSuffix()
		if suffix == "-" {
			continue
		}
		endpoint := fmt.Sprintf("https://%s%s", cluster.GetId(), suffix)
		privateRegionalEndpoints[network.GetId()] = endpoint
	}

	table := output.NewTable(cmd)
	table.Add(&listEndpoint{
		Public:          cluster.Spec.GetHttpEndpoint(),
		Private:         cluster.Spec.GetPrivateHttpEndpoint(),
		PrivateRegional: privateRegionalEndpoints,
		Catalog:         cluster.Spec.GetCatalogHttpEndpoint(),
	})

	return table.Print()
}
