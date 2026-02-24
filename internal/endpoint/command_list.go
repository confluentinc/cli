package endpoint

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	// Required flags
	cmd.Flags().String("service", "", "REQUIRED: Filter by service type (KAFKA, KSQL, SCHEMA_REGISTRY, etc.).")
	cobra.CheckErr(cmd.MarkFlagRequired("service"))
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	// Optional filter flags
	cmd.Flags().String("cloud", "", "Filter by cloud provider (AWS, GCP, AZURE).")
	cmd.Flags().String("region", "", "Filter by region.")
	cmd.Flags().Bool("is-private", false, "Filter by privacy (true for private, false for public).")
	cmd.Flags().String("resource", "", "Filter by resource ID.")

	// Standard flags
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	// Get environment ID
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Get required service flag
	service, err := cmd.Flags().GetString("service")
	if err != nil {
		return err
	}

	// Get optional filter flags
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return err
	}

	// Handle is-private flag (optional boolean)
	// nil means no filter, &true means filter for private, &false means filter for public
	var isPrivate *bool
	if cmd.Flags().Changed("is-private") {
		val, err := cmd.Flags().GetBool("is-private")
		if err != nil {
			return err
		}
		isPrivate = &val
	}

	// Convert to uppercase for API consistency
	service = strings.ToUpper(service)
	if cloud != "" {
		cloud = strings.ToUpper(cloud)
	}

	// Call API via client wrapper
	endpoints, err := c.V2Client.ListEndpoints(environmentId, cloud, region, service, isPrivate, resource)
	if err != nil {
		return err
	}

	// Sort by Cloud, then Region, then ID for deterministic output
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].GetCloud() != endpoints[j].GetCloud() {
			return endpoints[i].GetCloud() < endpoints[j].GetCloud()
		}
		if endpoints[i].GetRegion() != endpoints[j].GetRegion() {
			return endpoints[i].GetRegion() < endpoints[j].GetRegion()
		}
		return endpoints[i].GetId() < endpoints[j].GetId()
	})

	// Build output
	list := output.NewList(cmd)
	for _, endpoint := range endpoints {
		out := &out{
			Id:        endpoint.GetId(),
			Cloud:     endpoint.GetCloud(),
			Region:    endpoint.GetRegion(),
			Service:   endpoint.GetService(),
			IsPrivate: endpoint.GetIsPrivate(),
			Endpoint:  endpoint.GetEndpoint(),
		}

		// Add environment if present
		if endpoint.Environment != nil {
			out.Environment = endpoint.Environment.GetId()
		}

		// Add optional fields if present
		if endpoint.ConnectionType != nil {
			out.ConnectionType = endpoint.GetConnectionType()
		}
		if endpoint.EndpointType != nil {
			out.EndpointType = endpoint.GetEndpointType()
		}
		if endpoint.Resource != nil {
			out.Resource = endpoint.Resource.Id
		}
		if endpoint.Gateway != nil {
			out.Gateway = endpoint.Gateway.Id
		}
		if endpoint.AccessPoint != nil {
			out.AccessPoint = endpoint.AccessPoint.Id
		}

		list.Add(out)
	}

	// Sort and filter fields for display
	list.Sort(false)
	list.Filter([]string{"Id", "Cloud", "Region", "Service", "IsPrivate", "ConnectionType", "Endpoint", "EndpointType", "Environment", "Resource", "Gateway", "AccessPoint"})

	return list.Print()
}
