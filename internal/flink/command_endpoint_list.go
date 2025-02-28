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
	regionsList, err := c.V2Client.ListFlinkRegions(strings.ToUpper(cloud))

	// 1 - List all the public endpoint based on cloud and region
	for _, regionResult := range regionsList {
		publicEndpoint := &flinkEndpointOut{
			Endpoint: regionResult.GetHttpEndpoint(),
			Cloud:    regionResult.GetCloud(),
			Region:   regionResult.GetRegionName(),
			Type:     publicEndpointType,
		}
		list.Add(publicEndpoint)
	}

	// 2 - List all the private endpoint based on the presence of PrivateLinkAttachments as filter
	// TODO: double check with Flink team on the filtering implementation
	platts, err := c.V2Client.ListPrivateLinkAttachments(environmentId, []string{}, []string{cloud}, []string{region}, []string{"READY"})
	if err != nil {
		return err
	}
	filterKeyMap := buildCloudRegionKeyFilterMap(platts)

	for _, regionResult := range regionsList {
		key := CloudRegionKey{
			cloud:  regionResult.GetCloud(),
			region: regionResult.GetRegionName(),
		}

		if _, ok := filterKeyMap[key]; ok {
			privateEndpoint := &flinkEndpointOut{
				Endpoint: regionResult.GetPrivateHttpEndpoint(),
				Cloud:    regionResult.GetCloud(),
				Region:   regionResult.GetRegionName(),
				Type:     privateEndpointType,
			}
			list.Add(privateEndpoint)
		}
	}

	// 3 - List all the CCN endpoint with the list of network domains
	// TODO: check about the empty slice parameters
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

func buildCloudRegionKeyFilterMap(platts []networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment) map[CloudRegionKey]bool {
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
