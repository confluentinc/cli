package kafka

import (
	"fmt"
	"net/http"

	"github.com/antihax/optional"
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
	listMirrorFields               = []string{"LinkName", "MirrorTopicName", "NumPartition", "MaxPerPartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs"}
	structuredListMirrorFields     = camelToSnake(listMirrorFields)
	humanListMirrorFields          = camelToSpaced(listMirrorFields)
	describeMirrorFields           = []string{"LinkName", "MirrorTopicName", "Partition", "PartitionMirrorLag", "SourceTopicName", "MirrorStatus", "StatusTimeMs", "LastSourceFetchOffset"}
	structuredDescribeMirrorFields = camelToSnake(describeMirrorFields)
	humanDescribeMirrorFields      = camelToSpaced(describeMirrorFields)
	alterMirrorFields              = []string{"MirrorTopicName", "Partition", "PartitionMirrorLag", "ErrorMessage", "ErrorCode", "LastSourceFetchOffset"}
	structuredAlterMirrorFields    = camelToSnake(alterMirrorFields)
	humanAlterMirrorFields         = camelToSpaced(alterMirrorFields)
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
	LinkName              string
	MirrorTopicName       string
	SourceTopicName       string
	MirrorStatus          string
	StatusTimeMs          int64
	Partition             int32
	PartitionMirrorLag    int64
	LastSourceFetchOffset int64
}

type alterMirrorWrite struct {
	MirrorTopicName       string
	Partition             int32
	ErrorMessage          string
	ErrorCode             string
	PartitionMirrorLag    int64
	LastSourceFetchOffset int64
}

type mirrorCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func NewMirrorCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:         "mirror",
			Short:       "Manages cluster linking mirror topics.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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
		RunE: c.list,
		Args: cobra.NoArgs,
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
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
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
				Text: "Create a cluster link.",
				Code: "ccloud kafka mirror create <source-topic-name> --link <link> " +
					"--replication-factor <replication-factor> --config-file mirror_config.txt",
			},
			examples.Example{
				Code: "ccloud kafka mirror create <source-topic-name> --link <link>",
			},
		),
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(linkFlagName, "", "The name of the cluster link to attach to the mirror topic.")
	check(createCmd.MarkFlagRequired(linkFlagName))
	createCmd.Flags().Int32(replicationFactorFlagName, 3, "Replication factor.")
	createCmd.Flags().String(configFileFlagName, "", "Name of a file with additional topic configuration. "+
		"Each property should be on its own line with the format: key=value.")
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
		RunE: c.promote,
		Args: cobra.MinimumNArgs(1),
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
		RunE: c.failover,
		Args: cobra.MinimumNArgs(1),
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
		RunE: c.pause,
		Args: cobra.MinimumNArgs(1),
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
		RunE: c.resume,
		Args: cobra.MinimumNArgs(1),
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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
		cmd, listMirrorFields, humanListMirrorFields, structuredListMirrorFields)
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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
		cmd, describeMirrorFields, humanDescribeMirrorFields, structuredDescribeMirrorFields)
	if err != nil {
		return err
	}

	for _, partitionLag := range mirror.MirrorLags {
		outputWriter.AddElement(&describeMirrorWrite{
			LinkName:              mirror.LinkName,
			MirrorTopicName:       mirror.MirrorTopicName,
			SourceTopicName:       mirror.SourceTopicName,
			MirrorStatus:          string(mirror.MirrorStatus),
			StatusTimeMs:          mirror.StateTimeMs,
			Partition:             partitionLag.Partition,
			PartitionMirrorLag:    partitionLag.Lag,
			LastSourceFetchOffset: partitionLag.LastSourceFetchOffset,
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

	configMap, err := utils.ReadConfigsFromFile(configs)
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
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

func printAlterMirrorResult(cmd *cobra.Command, results kafkarestv3.AlterMirrorStatusResponseDataList) error {
	outputWriter, err := output.NewListOutputWriter(
		cmd, alterMirrorFields, humanAlterMirrorFields, structuredAlterMirrorFields)
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
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             -1,
				ErrorMessage:          errMsg,
				ErrorCode:             code,
				PartitionMirrorLag:    -1,
				LastSourceFetchOffset: -1,
			})
			continue
		}

		for _, partitionLag := range result.MirrorLags {
			outputWriter.AddElement(&alterMirrorWrite{
				MirrorTopicName:       result.MirrorTopicName,
				Partition:             partitionLag.Partition,
				ErrorMessage:          errMsg,
				ErrorCode:             code,
				PartitionMirrorLag:    partitionLag.Lag,
				LastSourceFetchOffset: partitionLag.LastSourceFetchOffset,
			})
		}
	}

	return outputWriter.Out()
}
