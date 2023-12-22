package kafka

import (
	"context"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	v3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
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

	mirrorOuts := make([]mirrorOut, 0)

	mirrorStateTransitionErrors := toMirrorStateTransitionError(mirror.GetMirrorStateTransitionErrors())
	for _, partitionLag := range mirror.GetMirrorLags().Items {
		mirrorOuts = append(mirrorOuts, mirrorOut{
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
	list := output.NewList(cmd)
	for i := range mirrorOuts {
		list.Add(&mirrorOuts[i])
	}
	isSerialized := output.GetFormat(cmd).IsSerialized()
	if isSerialized {
		list.Filter(getDescribeMirrorsFields(true))
		return list.Print()
	} else {
		list.Filter(getDescribeMirrorsFields(false))
		err = list.Print()
		if err != nil {
			return err
		}
		errsList := output.NewList(cmd)
		for i := range mirrorStateTransitionErrors {
			errsList.Add(&mirrorStateTransitionErrors[i])
		}
		output.Println(false, "Transition errors")
		return errsList.Print()
	}
}

type mirrorTransitionErrorOut struct {
	ErrorCode    string `human:"Error code" serialized:"error_code"`
	ErrorMessage string `human:"Error message" serialized:"error_message"`
}

func getDescribeMirrorsFields(includeTransitionErrors bool) []string {
	x := []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"}
	if includeTransitionErrors {
		x = append(x, "MirrorStateTransitionErrors", "ErrorCode", "ErrorMessage")
	}
	return x
}

func toMirrorStateTransitionError(errs []v3.LinkTaskError) []mirrorTransitionErrorOut {
	var errsToEncode []kafkarestv3.LinkTaskError
	if errs != nil {
		errsToEncode = errs
	} else {
		errsToEncode = make([]kafkarestv3.LinkTaskError, 0)
	}
	transitionErrorOuts := make([]mirrorTransitionErrorOut, 0)
	for _, errToEncode := range errsToEncode {
		transitionErrorOuts = append(transitionErrorOuts, mirrorTransitionErrorOut{
			ErrorCode:    errToEncode.ErrorCode,
			ErrorMessage: errToEncode.ErrorMessage,
		})
	}
	return transitionErrorOuts
}
