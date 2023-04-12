package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

const (
	configFileFlagName = "config-file"
	dryrunFlagName     = "dry-run"
)

type linkCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newLinkCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "link",
		Short:       "Manage inter-cluster links.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &linkCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newConfigurationCommand(cfg))
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newConfigurationCommand(cfg))
		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}

func (c *linkCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteLinks()
}

func (c *linkCommand) autocompleteLinks() []string {
	links, err := c.getLinks()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(links.Data))
	for i, link := range links.Data {
		description := fmt.Sprintf("%s: %s", link.GetSourceClusterId(), link.GetDestinationClusterId())
		suggestions[i] = fmt.Sprintf("%s\t%s", link.GetLinkName(), description)
	}
	return suggestions
}

func (c *linkCommand) getLinks() (kafkarestv3.ListLinksResponseDataList, error) {
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return kafkarestv3.ListLinksResponseDataList{}, err
		}
		return kafkarestv3.ListLinksResponseDataList{}, errors.New(errors.RestProxyNotAvailableMsg)
	}

	clusterId, err := getKafkaClusterLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return kafkarestv3.ListLinksResponseDataList{}, err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.CloudClient.ListKafkaLinks(clusterId)

	return listLinksRespDataList, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
}
