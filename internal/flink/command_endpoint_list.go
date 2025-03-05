package flink

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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
				Text: "List the available Flink endpoints.",
				Code: "confluent flink endpoint list",
			},
			examples.Example{
				Text: "List the available Flink endpoints for AWS on region us-west-2.",
				Code: "confluent flink endpoint list --cloud aws --region us-west-2",
			},
		),
	}

	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	var results []*flinkEndpointOut
	flinkRegions, err := c.V2Client.ListFlinkRegions(cloud, region)

	// 1 - List all the public endpoints based optionally on cloud(upper case) and region(lower case)
	for _, flinkRegion := range flinkRegions {
		results = append(results, &flinkEndpointOut{
			Endpoint: flinkRegion.GetHttpEndpoint(),
			Cloud:    flinkRegion.GetCloud(),
			Region:   flinkRegion.GetRegionName(),
			Type:     publicFlinkEndpointType,
		})
	}

	// 2 - List all the private endpoints based on the presence of "READY" PrivateLinkAttachments as filter
	platts, err := c.V2Client.ListPrivateLinkAttachments(environmentId, []string{}, []string{cloud}, []string{region}, []string{"READY"})
	if err != nil {
		return err
	}
	filterKeyMap := buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(platts)

	// TODO: De-duplications are needed
	for _, flinkRegion := range flinkRegions {
		key := CloudRegionKey{
			cloud:  flinkRegion.GetCloud(),
			region: flinkRegion.GetRegionName(),
		}

		if _, ok := filterKeyMap[key]; ok {
			results = append(results, &flinkEndpointOut{
				Endpoint: flinkRegion.GetPrivateHttpEndpoint(),
				Cloud:    flinkRegion.GetCloud(),
				Region:   flinkRegion.GetRegionName(),
				Type:     privateFlinkEndpointType,
			})
		}
	}

	// 3 - List all the CCN endpoint with the list of "READY" network domains
	networks, err := c.V2Client.ListNetworks(environmentId, nil, []string{cloud}, []string{region}, nil, []string{"READY"}, nil)
	for _, network := range networks {
		suffix := network.Status.GetEndpointSuffix()
		results = append(results, &flinkEndpointOut{
			Endpoint: fmt.Sprintf("https://flink%s", suffix),
			Cloud:    network.Spec.GetCloud(),
			Region:   network.Spec.GetRegion(),
			Type:     ccnFlinkEndpointType,
		})
	}

	// Sort the results order by cloud, then region, then type, then endpoint
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
			Endpoint: result.Endpoint,
			Cloud:    result.Cloud,
			Region:   result.Region,
			Type:     result.Type,
		})
	}

	// Disable the default sort to use the custom sort above
	list.Sort(false)
	return list.Print()
}

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
