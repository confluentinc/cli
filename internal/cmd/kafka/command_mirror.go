package kafka

import (
	"fmt"
	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listMirrorOutputFields = []string{"DestinationTopicName", "SourceTopicName", "MirrorStatus", "StatusTimeMs"}
	AlterMirrorOutputFields = []string{"DestinationTopicName", "ErrorMessage", "ErrorCode"}
)

type listMirrorWrite struct {
	DestinationTopicName  string
	SourceTopicName string
	MirrorStatus string
	StatusTimeMs int32
}

type alterMirrorWrite struct {
	DestinationTopicName  string
	ErrorMessage string
	ErrorCode string
}

type mirrorCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func NewMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:    "mirror",
			Hidden: true,
			Short:  "Manages cluster linking mirror topics",
		},
		prerunner, MirrorSubcommandFlags)
	cmd := &mirrorCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:                     prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *mirrorCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all mirror topics under the cluster link",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all mirrors under the link",
				Code: "ccloud kafka mirror list --link-name <link-name> --mirror-status <mirror-status>",
			},
		),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().String(linkFlagName, "", "Cluster link ConfigName")
	listCmd.Flags().String(mirrorStatusFlagName, "", "Mirror topic status. Can be one of " +
		"[active, failed, paused, stopped, pending_stopped]. If not specified, list all mirror topics.")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <destination-topic-ConfigName>",
		Short: "Describes a mirror topic",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describes a mirror topic under the link.",
				Code: "ccloud kafka mirror describe <destination-topic-ConfigName> --link-ConfigName <link-ConfigName>",
			},
		),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().String(linkFlagName, "", "Cluster link ConfigName")
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	createCmd := &cobra.Command{
		Use:   "create <mirror-ConfigName>",
		Short: "Create a mirror topic under the link. Currently, destination topic ConfigName is required to be the same as the Source topic ConfigName.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using supplied Source URL and properties.",
				Code: "ccloud kafka mirror create <Source-topic-ConfigName> --link-ConfigName <link-ConfigName> " +
					"--replication-factor <replication-factor> --config=\"unclean.leader.election.enable=true\"",
			},
		),
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().Int32(replicationFactorFlagName, 3, "Replication-factor, default: 3.")
	createCmd.Flags().StringSlice(configFlagName, nil, "A comma-separated list of topic config overrides ('key=ConfigValue') for the topic being created.")
	createCmd.Flags().String(linkFlagName, "", "The ConfigName of the cluster link.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	promoteCmd := &cobra.Command{
		Use:   "promote <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
		Short: "Promote the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Promote the mirror topics.",
				Code: "ccloud kafka mirror promote <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
			},
		),
		RunE: c.promote,
		Args: cobra.MinimumNArgs(1),
	}
	promoteCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	promoteCmd.Flags().String(linkFlagName, "", "The ConfigName of the cluster link.")
	promoteCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	c.AddCommand(promoteCmd)

	failoverCmd := &cobra.Command{
		Use:   "failover <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
		Short: "Failover the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Failover the mirror topics.",
				Code: "ccloud kafka mirror failover <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
			},
		),
		RunE: c.failover,
		Args: cobra.MinimumNArgs(1),
	}
	failoverCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	failoverCmd.Flags().String(linkFlagName, "", "The ConfigName of the cluster link.")
	failoverCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	c.AddCommand(failoverCmd)

	pauseCmd := &cobra.Command{
	Use:   "pause <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
	Short: "Pause the mirror topics.",
	Example: examples.BuildExampleString(
		examples.Example{
			Text: "Pause the mirror topics.",
			Code: "ccloud kafka mirror pause <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link-ConfigName <link-ConfigName>",
		},
	),
	RunE: c.pause,
	Args: cobra.MinimumNArgs(1),
	}
	pauseCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	pauseCmd.Flags().String(linkFlagName, "", "The ConfigName of the cluster link.")
	pauseCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	c.AddCommand(pauseCmd)

	resumeCmd := &cobra.Command{
		Use:   "resume <destination-topic-1> <destination-topic-2> ... <destination-topic-N>",
		Short: "Resume the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume the mirror topics.",
				Code: "ccloud kafka mirror resume <destination-topic-1> <destination-topic-2> ... <destination-topic-N>",
			},
		),
		RunE: c.resume,
		Args: cobra.MinimumNArgs(1),
	}
	resumeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	resumeCmd.Flags().String(linkFlagName, "", "The ConfigName of the cluster link.")
	resumeCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	c.AddCommand(resumeCmd)
}

func (c *mirrorCommand) list(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	mirrorStatus, err := cmd.Flags().GetString(mirrorStatusFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	mirrorStatusOpt := optional.EmptyInterface()
	if mirrorStatus != "" {
		mirrorStatusOpt = optional.NewInterface(kafkarestv3.MirrorTopicStatus(mirrorStatus))
	}

	listMirrorTopicsResponseDataList, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsGet(
			kafkaREST.Context, lkc, linkName,
			&kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsGetOpts{MirrorStatus: mirrorStatusOpt})
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(
		cmd, listMirrorOutputFields, listMirrorOutputFields, listMirrorOutputFields)
	if err != nil {
		return err
	}

	for _, mirror := range listMirrorTopicsResponseDataList.Data {
		outputWriter.AddElement(&listMirrorWrite{
			DestinationTopicName: mirror.DestinationTopicName,
			SourceTopicName:      mirror.SourceTopicName,
			MirrorStatus:         string(mirror.MirrorTopicStatus),
			StatusTimeMs:         mirror.StateTimeMs,
		})
	}

	return outputWriter.Out()
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	destinationTopicName := args[0]
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listMirrorTopicsResponseData, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsDestinationTopicNameGet(
			kafkaREST.Context, lkc, linkName, destinationTopicName)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(
		cmd, listMirrorOutputFields, listMirrorOutputFields, listMirrorOutputFields)
	if err != nil {
		return err
	}

	outputWriter.AddElement(&listMirrorWrite{
		DestinationTopicName: listMirrorTopicsResponseData.DestinationTopicName,
		SourceTopicName:      listMirrorTopicsResponseData.SourceTopicName,
		MirrorStatus:         string(listMirrorTopicsResponseData.MirrorTopicStatus),
		StatusTimeMs:         listMirrorTopicsResponseData.StateTimeMs,
	})

	return outputWriter.Out()
}

func (c *mirrorCommand) create(cmd *cobra.Command, args []string) error {
	sourceTopicName := args[0]

	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	replicationFactor, err := cmd.Flags().GetInt32(replicationFactorFlagName)
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice(configFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	configMap, err := toMap(configs)
	if err != nil {
		return err
	}

	createMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsPostOpts{
		CreateMirrorTopicRequestData: optional.NewInterface(
			kafkarestv3.CreateMirrorTopicRequestData{
				SourceTopicName:   sourceTopicName,
				ReplicationFactor: replicationFactor,
				Configs: 		   toCreateTopicConfigs(configMap),
			},
		),
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsPost(kafkaREST.Context, lkc, linkName, createMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	utils.Printf(cmd, errors.CreatedMirrorMsg, sourceTopicName)
	return nil
}

func (c *mirrorCommand) promote(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	promoteMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsPromotePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				DestinationTopics: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsPromotePost(kafkaREST.Context, lkc, linkName, promoteMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func (c *mirrorCommand) failover(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	failoverMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsFailoverPostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				DestinationTopics: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsFailoverPost(kafkaREST.Context, lkc, linkName, failoverMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func (c *mirrorCommand) pause(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	pauseMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsPausePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				DestinationTopics: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsPausePost(kafkaREST.Context, lkc, linkName, pauseMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func (c *mirrorCommand) resume(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	resumeMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsResumePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				DestinationTopics: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsResumePost(kafkaREST.Context, lkc, linkName, resumeMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func printAlterMirrorResult(cmd *cobra.Command, results kafkarestv3.AlterMirrorStatusResponseDataList) error {
	outputWriter, err := output.NewListOutputWriter(
		cmd, AlterMirrorOutputFields, AlterMirrorOutputFields, AlterMirrorOutputFields)
	if err != nil {
		return err
	}

	for _, result := range results.Data {
		var msg = "Null"
		var code = "Null"

		if result.ErrorMessage != nil {
			msg = *result.ErrorMessage
		}

		if result.ErrorCode != nil {
			code = fmt.Sprint(*result.ErrorCode)
		}

		outputWriter.AddElement(&alterMirrorWrite{
			DestinationTopicName: result.DestinationTopicName,
			ErrorMessage: msg,
			ErrorCode: code,
		})
	}

	return outputWriter.Out()
}
