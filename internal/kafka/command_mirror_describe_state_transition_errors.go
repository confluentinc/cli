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

func (c *mirrorCommand) newDescribeStateTransitionErrorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe-state-transition-errors <destination-topic-name>",
		Short:             "Describe a mirror topics state transition errors.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describeStateTransitionErrors,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe mirror topic state transition errors "my-topic" under the link "my-link":`,
				Code: "confluent kafka mirror describe-state-transition-errors my-topic --link my-link",
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

func (c *mirrorCommand) describeStateTransitionErrors(cmd *cobra.Command, args []string) error {
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

	mirrorStateTransitionErrors := toMirrorStateTransitionError(mirror.GetMirrorStateTransitionErrors())
	list := output.NewList(cmd)
	for i := range mirrorStateTransitionErrors {
		list.Add(&mirrorStateTransitionErrors[i])
	}
	return list.Print()
}

type mirrorTransitionErrorOut struct {
	ErrorCode    string `human:"Mirror State Transition Error Code" serialized:"error_code"`
	ErrorMessage string `human:"Mirror State Transition Error Message" serialized:"error_message"`
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
