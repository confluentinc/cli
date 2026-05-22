package flink

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

const (
	flinkEndpointService   = "FLINK"
	flinkRestEndpointType  = "REST"
	flinkEndpointUrlScheme = "https://"
)

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
	cloud := c.Context.GetCurrentFlinkCloudProvider()
	region := c.Context.GetCurrentFlinkRegion()
	if cloud == "" || region == "" {
		return errors.NewErrorWithSuggestions(
			"Current Flink cloud provider or region is empty",
			"Run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink cloud provider and region first.",
		)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	endpoints, err := c.V2Client.ListEndpoints(environmentId, cloud, region, flinkEndpointService, nil, "")
	if err != nil {
		return fmt.Errorf("unable to list Flink endpoints: %w", err)
	}

	currentEndpoint := c.Context.GetCurrentFlinkEndpoint()
	results := make([]*flinkEndpointOut, 0, len(endpoints))
	for _, e := range endpoints {
		// The Endpoints API also returns LANGUAGE_SERVICE endpoints (`flinkpls.*`) used by
		// the Cloud Console SQL editor's language server. Those are not usable by CLI
		// commands like `flink shell`, `flink statement`, or `flink endpoint use`.
		if e.GetEndpointType() != flinkRestEndpointType {
			continue
		}
		endpointType := publicFlinkEndpointType
		if e.GetIsPrivate() {
			endpointType = privateFlinkEndpointType
		}
		url := flinkEndpointUrl(e.GetEndpoint())
		results = append(results, &flinkEndpointOut{
			IsCurrent: url == currentEndpoint,
			Endpoint:  url,
			Cloud:     strings.ToUpper(e.GetCloud()),
			Region:    e.GetRegion(),
			Type:      endpointType,
		})
	}

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

	list := output.NewList(cmd)
	for _, r := range results {
		list.Add(r)
	}
	list.Sort(false)
	return list.Print()
}

// flinkEndpointUrl renders the bare FQDN the Endpoints API returns into the URL
// form the legacy command produced (and that `flink shell` / `flink endpoint use`
// expect). If the API response already includes a URL scheme — as the test server
// gateway does (`http://127.0.0.1:1026`) — it is preserved.
func flinkEndpointUrl(endpoint string) string {
	if strings.Contains(endpoint, "://") {
		return endpoint
	}
	return flinkEndpointUrlScheme + endpoint
}
