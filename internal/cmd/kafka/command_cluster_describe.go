package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var basicDescribeFields = []string{"IsCurrent", "Id", "Name", "Type", "IngressLimit", "EgressLimit", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}

type describeStruct struct {
	IsCurrent          bool   `human:"Current" serialized:"is_current"`
	Id                 string `human:"ID" serialized:"id"`
	Name               string `human:"Name" serialized:"name"`
	Type               string `human:"Type" serialized:"type"`
	ClusterSize        int32  `human:"Cluster Size" serialized:"cluster_size"`
	PendingClusterSize int32  `human:"Pending Cluster Size" serialized:"pending_cluster_size"`
	IngressLimit       int32  `human:"Ingress Limit (MB/s)" serialized:"ingress"`
	EgressLimit        int32  `human:"Egress Limit (MB/s)" serialized:"egress"`
	Storage            string `human:"Storage" serialized:"storage"`
	ServiceProvider    string `human:"Provider" serialized:"provider"`
	Region             string `human:"Region" serialized:"region"`
	Availability       string `human:"Availability" serialized:"availability"`
	Status             string `human:"Status" serialized:"status"`
	Endpoint           string `human:"Endpoint" serialized:"endpoint"`
	ByokKeyId          string `human:"BYOK Key ID" serialized:"byok_key_id"`
	EncryptionKeyId    string `human:"Encryption Key ID" serialized:"encryption_key_id"`
	RestEndpoint       string `human:"REST Endpoint" serialized:"rest_endpoint"`
	TopicCount         int    `human:"Topic Count,omitempty" serialized:"topic_count"`
}

func (c *clusterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id|name]",
		Short:             "Describe a Kafka cluster.",
		Long:              "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	clusterId, err := c.getLkcForDescribe(args)
	if err != nil {
		return err
	}

	ctx := c.Context.Config.Context()
	c.Context.Config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
	ctx.KafkaClusterContext.SetActiveKafkaCluster(clusterId)

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(clusterId, environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, clusterId, httpResp)
	}

	return c.outputKafkaClusterDescription(cmd, &cluster, true)
}

func (c *clusterCommand) getLkcForDescribe(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	lkc := c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId()
	if lkc == "" {
		return "", errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions)
	}

	return lkc, nil
}

func (c *clusterCommand) outputKafkaClusterDescription(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster, getTopicCount bool) error {
	out := convertClusterToDescribeStruct(cluster, c.Context.Context)

	if getTopicCount {
		topicCount, err := c.getTopicCountForKafkaCluster(cluster)
		// topicCount is 0 when err != nil, and will be omitted by `omitempty`
		if err != nil {
			log.CliLogger.Infof(errors.OmitTopicCountMsg, err)
		}
		out.TopicCount = topicCount
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(getKafkaClusterDescribeFields(cluster, basicDescribeFields, getTopicCount))
	return table.Print()
}

func convertClusterToDescribeStruct(cluster *cmkv2.CmkV2Cluster, ctx *v1.Context) *describeStruct {
	clusterStorage := getKafkaClusterStorage(cluster)
	ingress, egress := getCmkClusterIngressAndEgressMbps(cluster)

	return &describeStruct{
		IsCurrent:          cluster.GetId() == ctx.KafkaClusterContext.GetActiveKafkaClusterId(),
		Id:                 cluster.GetId(),
		Name:               cluster.Spec.GetDisplayName(),
		Type:               getCmkClusterType(cluster),
		ClusterSize:        getCmkClusterSize(cluster),
		PendingClusterSize: getCmkClusterPendingSize(cluster),
		IngressLimit:       ingress,
		EgressLimit:        egress,
		Storage:            clusterStorage,
		ServiceProvider:    strings.ToLower(cluster.Spec.GetCloud()),
		Region:             cluster.Spec.GetRegion(),
		Availability:       availabilitiesToHuman[cluster.Spec.GetAvailability()],
		Status:             getCmkClusterStatus(cluster),
		Endpoint:           cluster.Spec.GetKafkaBootstrapEndpoint(),
		ByokKeyId:          getCmkByokId(cluster),
		EncryptionKeyId:    getCmkEncryptionKey(cluster),
		RestEndpoint:       cluster.Spec.GetHttpEndpoint(),
	}
}

func getKafkaClusterStorage(cluster *cmkv2.CmkV2Cluster) string {
	if !isBasic(cluster) {
		return "Infinite"
	} else {
		return "5 TB"
	}
}

func getKafkaClusterDescribeFields(cluster *cmkv2.CmkV2Cluster, basicFields []string, getTopicCount bool) []string {
	describeFields := basicFields
	if isDedicated(cluster) {
		describeFields = append(describeFields, "ClusterSize")
		if isExpanding(cluster) || isShrinking(cluster) {
			describeFields = append(describeFields, "PendingClusterSize")
		}
		if cluster.Spec.Config.CmkV2Dedicated.EncryptionKey != nil && *cluster.Spec.Config.CmkV2Dedicated.EncryptionKey != "" {
			describeFields = append(describeFields, "EncryptionKeyId")
		}
		if cluster.Spec.Byok != nil {
			describeFields = append(describeFields, "ByokId")
		}
	}

	if getTopicCount {
		describeFields = append(describeFields, "TopicCount")
	}

	return describeFields
}

func (c *clusterCommand) getTopicCountForKafkaCluster(cluster *cmkv2.CmkV2Cluster) (int, error) {
	if getCmkClusterStatus(cluster) == ccloudv2.StatusProvisioning {
		return 0, nil
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return 0, err
	}

	topics, httpResp, err := kafkaREST.CloudClient.ListKafkaTopics(cluster.GetId())
	if err != nil {
		return 0, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return len(topics.Data), nil
}
