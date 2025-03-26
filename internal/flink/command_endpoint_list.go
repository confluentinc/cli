package flink

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type CloudRegionKey struct {
	cloud  string
	region string
}

func (c *command) newEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		RunE:  c.endpointList,
		Short: "List Flink endpoint.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available Flink endpoints with current cloud provider and region.",
				Code: "confluent flink endpoint list",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, _ []string) error {
	// Get the current Flink cloud and region
	cloud := c.Context.GetCurrentFlinkCloudProvider()
	region := c.Context.GetCurrentFlinkRegion()
	if cloud == "" || region == "" {
		return errors.NewErrorWithSuggestions(
			"Current Flink cloud provider or region is empty",
			"Run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink cloud provider and region first.",
		)
	}
	cloud = strings.ToUpper(cloud)

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	flinkRegions, err := c.V2Client.ListFlinkRegions(cloud, region)
	if err != nil {
		return fmt.Errorf("unable to list Flink endpoint, failed to list Flink regions: %w", err)
	}
	results := make([]*flinkEndpointOut, len(flinkRegions)*2)

	// 1 - List all the public endpoints based optionally on cloud(upper case) and region(lower case)
	for _, flinkRegion := range flinkRegions {
		results = append(results, &flinkEndpointOut{
			IsCurrent: flinkRegion.GetHttpEndpoint() == c.Context.GetCurrentFlinkEndpoint(),
			Endpoint:  flinkRegion.GetHttpEndpoint(),
			Cloud:     flinkRegion.GetCloud(),
			Region:    flinkRegion.GetRegionName(),
			Type:      publicFlinkEndpointType,
		})
	}

	// 2 - List all the private endpoints based on the presence of "READY" PrivateLinkAttachments as filter
	// Note the `cloud` and `region` parameters have to be `nil` instead of empty slice in case of no filter
	platts, err := c.V2Client.ListPrivateLinkAttachments(environmentId, nil, nil, nil, []string{"READY"})
	if err != nil {
		return fmt.Errorf("unable to list Flink endpoint, failed to list private link attachments: %w", err)
	}

	filterKeyMap := buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(platts)

	for _, flinkRegion := range flinkRegions {
		key := CloudRegionKey{
			cloud:  flinkRegion.GetCloud(),
			region: flinkRegion.GetRegionName(),
		}

		if _, ok := filterKeyMap[key]; ok {
			results = append(results, &flinkEndpointOut{
				IsCurrent: flinkRegion.GetPrivateHttpEndpoint() == c.Context.GetCurrentFlinkEndpoint(),
				Endpoint:  flinkRegion.GetPrivateHttpEndpoint(),
				Cloud:     flinkRegion.GetCloud(),
				Region:    flinkRegion.GetRegionName(),
				Type:      privateFlinkEndpointType,
			})
		}
	}

	// 3 - List all the CCN endpoint with the list of "READY" network domains
	// Note the cloud and region have to be empty slice instead of `nil` in case of no filter
	networks, err := c.V2Client.ListNetworks(environmentId, nil, []string{cloud}, []string{region}, nil, []string{"READY"}, nil)
	if err != nil {
		return fmt.Errorf("unable to list Flink endpoint, failed to list networks: %w", err)
	}

	for _, network := range networks {
		suffix := network.Status.GetEndpointSuffix()
		endpoint := fmt.Sprintf("https://flink%s", suffix)
		results = append(results, &flinkEndpointOut{
			IsCurrent: endpoint == c.Context.GetCurrentFlinkEndpoint(),
			Endpoint:  endpoint,
			Cloud:     network.Spec.GetCloud(),
			Region:    network.Spec.GetRegion(),
			Type:      privateFlinkEndpointType,
		})
	}

	// Sort the results order by cloud, region, type and endpoint
	sort.Slice(results, func(i, j int) bool {
		if results[i].Cloud != results[j].Cloud {
			return results[i].Cloud < results[j].Cloud
		}
		if results[i].Region != results[j].Region {
			return results[i].Region < results[j].Region
		}
		if results[i].Type != results[j].Type {
			return results[i].Type < results[j].Type
		}
		return results[i].Endpoint < results[j].Endpoint
	})

	for _, result := range results {
		list.Add(&flinkEndpointOut{
			IsCurrent: result.IsCurrent,
			Endpoint:  result.Endpoint,
			Cloud:     result.Cloud,
			Region:    result.Region,
			Type:      result.Type,
		})
	}

	// Disable the default sort to use the custom sort above
	list.Sort(false)
	return list.Print()
}

// buildCloudRegionKeyFilterMapFromPrivateLinkAttachments creates a map of unique cloud/region pairs from PrivateLinkAttachments.
// This function helps deduplicate scenarios where users have multiple private link attachments in the same cloud region.
// Each unique combination of cloud and region is represented as a CloudRegionKey in the returned map.
//
// Parameters:
//   - platts: A slice of NetworkingV1PrivateLinkAttachment objects to process
//
// Returns:
//   - A map with CloudRegionKey as keys and boolean 'true' as values for each unique cloud/region combination
func buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(platts []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment) map[CloudRegionKey]bool {
	result := make(map[CloudRegionKey]bool, len(platts))
	for _, platt := range platts {
		compositeKey := CloudRegionKey{
			cloud:  platt.Spec.GetCloud(),
			region: platt.Spec.GetRegion(),
		}
		result[compositeKey] = true
	}
	return result
}
