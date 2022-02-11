package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const includeTopicsFlagName = "include-topics"

var (
	listLinkFieldsIncludeTopics           = []string{"LinkName", "TopicName", "SourceClusterId"}
	structuredListLinkFieldsIncludeTopics = camelToSnake(listLinkFieldsIncludeTopics)
	humanListLinkFieldsIncludeTopics      = camelToSpaced(listLinkFieldsIncludeTopics)
	listLinkFields                        = []string{"LinkName", "SourceClusterId"}
	structuredListLinkFields              = camelToSnake(listLinkFields)
	humanListLinkFields                   = camelToSpaced(listLinkFields)
)

type LinkTopicWriter struct {
	LinkName        string
	TopicName       string
	SourceClusterId string
}

func (c *linkCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List previously created cluster links.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().Bool(includeTopicsFlagName, false, "If set, will list mirrored topics for the links returned.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) list(cmd *cobra.Command, _ []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.ListKafkaLinks(kafkaREST.Context, lkc)
	if err != nil {
		return handleOpenApiError(httpResp, err, kafkaREST)
	}

	if includeTopics {
		outputWriter, err := output.NewListOutputWriter(cmd, listLinkFieldsIncludeTopics, humanListLinkFieldsIncludeTopics, structuredListLinkFieldsIncludeTopics)
		if err != nil {
			return err
		}

		for _, link := range listLinksRespDataList.Data {
			if len(link.TopicsNames) > 0 {
				for _, topic := range link.TopicsNames {
					outputWriter.AddElement(
						&LinkTopicWriter{
							LinkName:        link.LinkName,
							TopicName:       topic,
							SourceClusterId: link.SourceClusterId,
						})
				}
			} else {
				outputWriter.AddElement(
					&LinkTopicWriter{
						LinkName:        link.LinkName,
						TopicName:       "",
						SourceClusterId: link.SourceClusterId,
					})
			}
		}

		return outputWriter.Out()
	} else {
		outputWriter, err := output.NewListOutputWriter(cmd, listLinkFields, humanListLinkFields, structuredListLinkFields)
		if err != nil {
			return err
		}

		for _, link := range listLinksRespDataList.Data {
			outputWriter.AddElement(&LinkTopicWriter{
				LinkName:        link.LinkName,
				SourceClusterId: link.SourceClusterId,
			})
		}

		return outputWriter.Out()
	}
}
