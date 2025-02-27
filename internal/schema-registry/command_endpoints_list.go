package schemaregistry

import (
	"github.com/spf13/cobra"

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
	list := output.NewTable(cmd)
	list.Add(&listEndpoint{
		Public:          cluster.Spec.GetHttpEndpoint(),
		Private:         cluster.Spec.GetPrivateHttpEndpoint(),
		PrivateRegional: cluster.Spec.PrivateNetworkingConfig.GetRegionalEndpoints(),
		Catalog:         cluster.Spec.GetCatalogHttpEndpoint(),
	})

	return list.Print()
}
