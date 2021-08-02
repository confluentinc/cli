package kafka

import (
	"context"
	"fmt"
	"net/http"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	replicationFactorFlagName = "replication-factor"
	mirrorStatusFlagName      = "mirror-status"
)

var (
	listMirrorOutputFieldsCamel          = []string{"LinkName", "MirrorTopicName", "NumPartition", "MaxPerPartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"}
	listMirrorOutputFieldsUnderscore     = []string{"link_name", "mirror_topic_name", "num_partition", "max_per_partition_mirror_lag", "source_topic_name", "mirror_status", "status_time_ms"}
	describeMirrorOutputFieldsCamel      = []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"}
	describeMirrorOutputFieldsUnderscore = []string{"link_name", "mirror_topic_name", "partition", "partition_mirror_lag", "source_topic_name", "mirror_status", "status_time_ms"}
	alterMirrorOutputFieldsCamel         = []string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode"}
	alterMirrorOutputFieldsUnderscore    = []string{"mirror_topic_name", "partition", "partition_mirror_lag", "error_message", "error_code"}
)

type listMirrorWrite struct {
	LinkName                 string
	MirrorTopicName          string
	SourceTopicName          string
	MirrorStatus             string
	StatusTimeMs             int64
	NumPartition             int32
	MaxPerPartitionMirrorLag int64
}

type describeMirrorWrite struct {
	LinkName           string
	MirrorTopicName    string
	SourceTopicName    string
	MirrorStatus       string
	StatusTimeMs       int64
	Partition          int32
	PartitionMirrorLag int64
}

type alterMirrorWrite struct {
	MirrorTopicName    string
	Partition          int32
	ErrorMessage       string
	ErrorCode          string
	PartitionMirrorLag int64
}

type mirrorCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func NewMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:    "mirror",
			Short:  "Manages cluster linking mirror topics.",
			Hidden: true,
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
		Short: "List all mirror topics in the cluster or under the given cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all mirrors in the cluster or under the given cluster link.",
				Code: "ccloud kafka mirror list --link <link> --mirror-status <mirror-status>",
			},
		),
		RunE:   c.list,
		Args:   cobra.NoArgs,
		Hidden: true,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().String(linkFlagName, "", "Cluster link name. If not specified, list all mirror topics in the cluster.")
	listCmd.Flags().String(mirrorStatusFlagName, "", "Mirror topic status. Can be one of "+
		"[active, failed, paused, stopped, pending_stopped]. If not specified, list all mirror topics.")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <destination-topic-name>",
		Short: "Describe a mirror topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a mirror topic under the link.",
				Code: "ccloud kafka mirror describe <destination-topic-name> --link <link>",
			},
		),
		RunE:   c.describe,
		Args:   cobra.ExactArgs(1),
		Hidden: true,
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().String(linkFlagName, "", "Cluster link name.")
	check(describeCmd.MarkFlagRequired(linkFlagName))
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	createCmd := &cobra.Command{
		Use:   "create <source-topic-name>",
		Short: "Create a mirror topic under the link. Currently, destination topic name is required to be the same as the Source topic name.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using supplied Source URL and properties.",
				Code: "ccloud kafka mirror create <source-topic-name> --link <link> " +
					"--replication-factor <replication-factor> --config-file mirror_config.txt",
			},
		),
		RunE:   c.create,
		Args:   cobra.ExactArgs(1),
		Hidden: true,
	}
	createCmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	check(createCmd.MarkFlagRequired(linkFlagName))
	createCmd.Flags().Int32(replicationFactorFlagName, 3, "Replication-factor, default: 3.")
	createCmd.Flags().String(configFileFlagName, "", "Name of the file containing topic config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	promoteCmd := &cobra.Command{
		Use:   "promote <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
		Short: "Promote the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Promote the mirror topics.",
				Code: "ccloud kafka mirror promote <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
			},
		),
		RunE:   c.promote,
		Args:   cobra.MinimumNArgs(1),
		Hidden: true,
	}
	promoteCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	promoteCmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	promoteCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	check(promoteCmd.MarkFlagRequired(linkFlagName))
	c.AddCommand(promoteCmd)

	failoverCmd := &cobra.Command{
		Use:   "failover <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
		Short: "Failover the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Failover the mirror topics.",
				Code: "ccloud kafka mirror failover <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
			},
		),
		RunE:   c.failover,
		Args:   cobra.MinimumNArgs(1),
		Hidden: true,
	}
	failoverCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	failoverCmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	failoverCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	check(failoverCmd.MarkFlagRequired(linkFlagName))
	c.AddCommand(failoverCmd)

	pauseCmd := &cobra.Command{
		Use:   "pause <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
		Short: "Pause the mirror topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pause the mirror topics.",
				Code: "ccloud kafka mirror pause <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link <link>",
			},
		),
		RunE:   c.pause,
		Args:   cobra.MinimumNArgs(1),
		Hidden: true,
	}
	pauseCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	pauseCmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	pauseCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	check(pauseCmd.MarkFlagRequired(linkFlagName))
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
		RunE:   c.resume,
		Args:   cobra.MinimumNArgs(1),
		Hidden: true,
	}
	resumeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	resumeCmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	resumeCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	check(resumeCmd.MarkFlagRequired(linkFlagName))
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

	var listMirrorTopicsResponseDataList kafkarestv3.ListMirrorTopicsResponseDataList
	var httpResp *http.Response

	if linkName == "" {
		listMirrorTopicsResponseDataList, httpResp, err = kafkaREST.Client.ClusterLinkingApi.
			ClustersClusterIdLinksMirrorsGet(
				kafkaREST.Context, lkc,
				&kafkarestv3.ClustersClusterIdLinksMirrorsGetOpts{MirrorStatus: mirrorStatusOpt})
	} else {
		listMirrorTopicsResponseDataList, httpResp, err = kafkaREST.Client.ClusterLinkingApi.
			ClustersClusterIdLinksLinkNameMirrorsGet(
				kafkaREST.Context, lkc, linkName,
				&kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsGetOpts{MirrorStatus: mirrorStatusOpt})
	}

	if err != nil {
		return handleOpenApiError(httpResp, err, kafkaREST)
	}

	outputWriter, err := output.NewListOutputWriter(
		cmd, listMirrorOutputFieldsCamel, listMirrorOutputFieldsCamel, listMirrorOutputFieldsUnderscore)
	if err != nil {
		return err
	}

	for _, mirror := range listMirrorTopicsResponseDataList.Data {
		var maxLag int64 = 0
		for _, mirrorLag := range mirror.MirrorLags {
			if mirrorLag.Lag > maxLag {
				maxLag = mirrorLag.Lag
			}
		}

		outputWriter.AddElement(&listMirrorWrite{
			LinkName:                 mirror.LinkName,
			MirrorTopicName:          mirror.MirrorTopicName,
			SourceTopicName:          mirror.SourceTopicName,
			MirrorStatus:             string(mirror.MirrorStatus),
			StatusTimeMs:             mirror.StateTimeMs,
			NumPartition:             mirror.NumPartitions,
			MaxPerPartitionMirrorLag: maxLag,
		})
	}

	return outputWriter.Out()
}

func (c *mirrorCommand) describe(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	mirrorTopicName := args[0]
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

	mirror, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsMirrorTopicNameGet(
			kafkaREST.Context, lkc, linkName, mirrorTopicName)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(
		cmd, describeMirrorOutputFieldsCamel, describeMirrorOutputFieldsCamel, describeMirrorOutputFieldsUnderscore)
	if err != nil {
		return err
	}

	for _, partitionLag := range mirror.MirrorLags {
		outputWriter.AddElement(&describeMirrorWrite{
			LinkName:           mirror.LinkName,
			MirrorTopicName:    mirror.MirrorTopicName,
			SourceTopicName:    mirror.SourceTopicName,
			MirrorStatus:       string(mirror.MirrorStatus),
			StatusTimeMs:       mirror.StateTimeMs,
			Partition:          partitionLag.Partition,
			PartitionMirrorLag: partitionLag.Lag,
		})
	}

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

	configs, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMap, err := readConfigsFromFile(configs)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		// Fall back to use kafka-api
		fmt.Println("Kafka REST is not enabled")
		return c.createWithKafkaApi(cmd, linkName, sourceTopicName, replicationFactor, configMap)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	createMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsPostOpts{
		CreateMirrorTopicRequestData: optional.NewInterface(
			kafkarestv3.CreateMirrorTopicRequestData{
				SourceTopicName:   sourceTopicName,
				ReplicationFactor: replicationFactor,
				Configs:           toCreateTopicConfigs(configMap),
			},
		),
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsPost(kafkaREST.Context, lkc, linkName, createMirrorOpt)
	if err == nil {
		utils.Printf(cmd, errors.CreatedMirrorMsg, sourceTopicName)
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
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

	promoteMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorspromotePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				MirrorTopicNames: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorspromotePost(kafkaREST.Context, lkc, linkName, promoteMirrorOpt)
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

	failoverMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsfailoverPostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				MirrorTopicNames: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsfailoverPost(kafkaREST.Context, lkc, linkName, failoverMirrorOpt)
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

	pauseMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorspausePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				MirrorTopicNames: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorspausePost(kafkaREST.Context, lkc, linkName, pauseMirrorOpt)
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

	resumeMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsresumePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(
			kafkarestv3.AlterMirrorsRequestData{
				MirrorTopicNames: args,
			},
		),
		ValidateOnly: optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.
		ClustersClusterIdLinksLinkNameMirrorsresumePost(kafkaREST.Context, lkc, linkName, resumeMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}

func (c *mirrorCommand) createWithKafkaApi(
	cmd *cobra.Command, linkName string, sourceTopicName string, factor int32, configMap map[string]string) error {
	// Kafka REST is not available, fall back to KafkaAPI

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Name:                 sourceTopicName,
			NumPartitions:        unspecifiedPartitionCount,
			ReplicationFactor:    uint32(factor),
			Configs:              configMap,
			Mirror:               &schedv1.TopicMirrorSpecification{LinkName: linkName, MirrorTopic: sourceTopicName},
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		},
		Validate: false,
	}

	if err := c.Client.Kafka.CreateTopic(context.Background(), cluster, topic); err != nil {
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}
	utils.Printf(cmd, errors.CreatedTopicMsg, topic.Spec.Name)
	return nil
}

func printAlterMirrorResult(cmd *cobra.Command, results kafkarestv3.AlterMirrorStatusResponseDataList) error {
	outputWriter, err := output.NewListOutputWriter(
		cmd, alterMirrorOutputFieldsCamel, alterMirrorOutputFieldsCamel, alterMirrorOutputFieldsUnderscore)
	if err != nil {
		return err
	}

	for _, result := range results.Data {
		var errMsg = ""
		var code = ""

		if result.ErrorMessage != nil {
			errMsg = *result.ErrorMessage
		}

		if result.ErrorCode != nil {
			code = fmt.Sprint(*result.ErrorCode)
		}

		// fatal error
		if errMsg != "" {
			outputWriter.AddElement(&alterMirrorWrite{
				MirrorTopicName:    result.MirrorTopicName,
				Partition:          -1,
				ErrorMessage:       errMsg,
				ErrorCode:          code,
				PartitionMirrorLag: -1,
			})
			continue
		}

		for _, partitionLag := range result.MirrorLags {
			outputWriter.AddElement(&alterMirrorWrite{
				MirrorTopicName:    result.MirrorTopicName,
				Partition:          partitionLag.Partition,
				ErrorMessage:       errMsg,
				ErrorCode:          code,
				PartitionMirrorLag: partitionLag.Lag,
			})
		}
	}

	return outputWriter.Out()
}
