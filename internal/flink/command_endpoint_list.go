package flink

import (
	"fmt"
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

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	flinkRegions, err := c.V2Client.ListFlinkRegions(strings.ToUpper(cloud), region)

	// 1 - List all the public endpoints based optionally on cloud and region
	for _, flinkRegion := range flinkRegions {
		publicEndpoint := &flinkEndpointOut{
			Endpoint: flinkRegion.GetHttpEndpoint(),
			Cloud:    flinkRegion.GetCloud(),
			Region:   flinkRegion.GetRegionName(),
			Type:     publicEndpointType,
		}
		list.Add(publicEndpoint)
	}

	// 2 - List all the private endpoints based on the presence of PrivateLinkAttachments as filter
	platts, err := c.V2Client.ListPrivateLinkAttachments(environmentId, []string{}, []string{cloud}, []string{region}, []string{"READY"})
	if err != nil {
		return err
	}
	filterKeyMap := buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(platts)

	for _, flinkRegion := range flinkRegions {
		key := CloudRegionKey{
			cloud:  flinkRegion.GetCloud(),
			region: flinkRegion.GetRegionName(),
		}

		if _, ok := filterKeyMap[key]; ok {
			privateEndpoint := &flinkEndpointOut{
				Endpoint: flinkRegion.GetPrivateHttpEndpoint(),
				Cloud:    flinkRegion.GetCloud(),
				Region:   flinkRegion.GetRegionName(),
				Type:     privateEndpointType,
			}
			list.Add(privateEndpoint)
		}
	}

	// 3 - List all the CCN endpoint with the list of network domains
	networks, err := c.V2Client.ListNetworks(environmentId, []string{}, []string{cloud}, []string{region}, []string{""}, []string{""}, []string{""})
	for _, network := range networks {
		suffix := network.Status.GetEndpointSuffix()
		ccnEndpoint := &flinkEndpointOut{
			Endpoint: fmt.Sprintf("flink%s", suffix),
			Cloud:    network.Spec.GetCloud(),
			Region:   network.Spec.GetRegion(),
			Type:     ccnEndpointType,
		}
		list.Add(ccnEndpoint)
	}

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
