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

	apiContext := context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, kafkaREST.CloudClient.AuthToken)

	req := kafkaREST.CloudClient.ClusterLinkingV3Api.ReadKafkaMirrorTopic(apiContext, kafkaREST.CloudClient.ClusterId, link, mirrorTopicName)
	req = req.IncludeStateTransitionErrors(true)

	res, httpResp, err := req.Execute()
	mirror, err := res, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
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
		// If we are serializing the output then there's no need to do any customization of the output. It will get
		// correctly serialized.
		list.Filter(getDescribeMirrorsFields(true))
		return list.Print()
	} else {
		// If we are not serializing the output, which means it's for human consumption, then we do some customization
		// so it's more readable.
		// We print two lists. The first one contains the mirror information per partition. After that we print a list
		// containing the mirror state transition errors. This is because the state transition error information is at the
		// topic level, so if we used the first list we'd duplicate the transition error for each partition which
		// would look awkward and confusing.
		list.Filter(getDescribeMirrorsFields(false))
		err = list.Print()
		if err != nil {
			return err
		}
		if len(mirrorStateTransitionErrors) > 0 {
			errsList := output.NewList(cmd)
			for i := range mirrorStateTransitionErrors {
				errsList.Add(&mirrorStateTransitionErrors[i])
			}
			return errsList.Print()
		} else {
			return nil
		}
	}
}

type mirrorTransitionErrorOut struct {
	ErrorCode    string `human:"Mirror State Transition Error Code" serialized:"error_code"`
	ErrorMessage string `human:"Mirror State Transition Error Message" serialized:"error_message"`
}

func getDescribeMirrorsFields(includeTransitionErrors bool) []string {
	x := []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"}
	if includeTransitionErrors {
		x = append(x, "MirrorStateTransitionErrors", "ErrorCode", "ErrorMessage")
	}
	return x
}

func toMirrorStateTransitionError(errs []v3.LinkTaskError) []mirrorTransitionErrorOut {
	if errs == nil {
		return make([]mirrorTransitionErrorOut, 0)
	}
	transitionErrorOuts := make([]mirrorTransitionErrorOut, 0)
	for _, err := range errs {
		transitionErrorOuts = append(transitionErrorOuts, mirrorTransitionErrorOut{
			ErrorCode:    err.ErrorCode,
			ErrorMessage: err.ErrorMessage,
		})
	}
	return transitionErrorOuts
}
