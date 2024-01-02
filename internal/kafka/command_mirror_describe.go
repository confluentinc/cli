package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	v3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *mirrorCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <destination-topic-name>",
		Short:             "Describe a mirror topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe mirror topic "my-topic" under the link "my-link":`,
				Code: "confluent kafka mirror describe my-topic --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	mirrorTopicName := args[0]

	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cloudClient := kafkaREST.CloudClient
	apiContext := context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, cloudClient.AuthToken)

	req := cloudClient.ClusterLinkingV3Api.ReadKafkaMirrorTopic(apiContext, cloudClient.ClusterId, link, mirrorTopicName)
	req = req.IncludeStateTransitionErrors(true)

	res, httpResp, err := req.Execute()
	mirror, err := res, kafkarest.NewError(cloudClient.GetUrl(), err, httpResp)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)

	mirrorStateTransitionErrors, err := toMirrorStateTransitionError(mirror.GetMirrorStateTransitionErrors())
	if err != nil {
		return err
	}
	for _, partitionLag := range mirror.GetMirrorLags().Items {
		mirror.GetMirrorStateTransitionErrors()
		list.Add(&mirrorOut{
			LinkName:                    mirror.GetLinkName(),
			MirrorTopicName:             mirror.GetMirrorTopicName(),
			SourceTopicName:             mirror.GetSourceTopicName(),
			MirrorStatus:                string(mirror.GetMirrorStatus()),
			StatusTimeMs:                mirror.GetStateTimeMs(),
			Partition:                   partitionLag.GetPartition(),
			PartitionMirrorLag:          partitionLag.GetLag(),
			LastSourceFetchOffset:       partitionLag.GetLastSourceFetchOffset(),
			MirrorStateTransitionErrors: mirrorStateTransitionErrors,
		})
	}
	list.Filter(getDescribeMirrorsFields())
	return list.Print()
}

func getDescribeMirrorsFields() []string {
	x := []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset", "MirrorStateTransitionErrors"}
	return x
}

func toMirrorStateTransitionError(errs []v3.LinkTaskError) (string, error) {
	var errsToEncode []kafkarestv3.LinkTaskError
	if errs != nil {
		errsToEncode = errs
	} else {
		// If nil create an empty slice so that the encoded JSON is [] instead of null.
		errsToEncode = make([]kafkarestv3.LinkTaskError, 0)
	}
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(errsToEncode)
	if err != nil {
		return "", err
	} else {
		return b.String(), nil
	}
}
