package kafka

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

var basicDescribeFields = []string{"IsCurrent", "Id", "Name", "Type", "IngressLimit", "EgressLimit", "Storage", "Cloud", "Availability", "Region", "Network", "Status", "Endpoint", "RestEndpoint"}

type describeStruct struct {
	IsCurrent          bool   `human:"Current" serialized:"is_current"`
	Id                 string `human:"ID" serialized:"id"`
	Name               string `human:"Name" serialized:"name"`
	Type               string `human:"Type" serialized:"type"`
	ClusterSize        int32  `human:"Cluster Size" serialized:"cluster_size"`
	PendingClusterSize int32  `human:"Pending Cluster Size" serialized:"pending_cluster_size"`
	IngressLimit       int32  `human:"Ingress Limit (MB/s)" serialized:"ingress_limit"`
	EgressLimit        int32  `human:"Egress Limit (MB/s)" serialized:"egress_limit"`
	Storage            string `human:"Storage" serialized:"storage"`
	Cloud              string `human:"Cloud" serialized:"cloud"`
	Region             string `human:"Region" serialized:"region"`
	Availability       string `human:"Availability" serialized:"availability"`
	Network            string `human:"Network,omitempty" serialized:"network,omitempty"`
	Status             string `human:"Status" serialized:"status"`
	Endpoint           string `human:"Endpoint" serialized:"endpoint"`
	ByokKeyId          string `human:"BYOK Key ID" serialized:"byok_key_id"`
	EncryptionKeyId    string `human:"Encryption Key ID" serialized:"encryption_key_id"`
	RestEndpoint       string `human:"REST Endpoint" serialized:"rest_endpoint"`
	TopicCount         int    `human:"Topic Count,omitempty" serialized:"topic_count,omitempty"`
}

func (c *clusterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id]",
		Short:             "Describe a Kafka cluster.",
		Long:              "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	lkc, err := c.getLkcForDescribe(args)
	if err != nil {
		return err
	}

	ctx := c.Context.Config.Context()
	c.Context.Config.SetOverwrittenCurrentKafkaCluster(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
	ctx.KafkaClusterContext.SetActiveKafkaCluster(lkc)

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}

	if activeEndpoint := c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(); activeEndpoint != "" {
		if output.GetFormat(cmd) == output.Human {
			output.Printf(c.Config.EnableColor, "The current endpoint is set to %q, "+
				"use `kafka cluster endpoint list` to view the available endpoints\n", activeEndpoint)
		}
	}

	return c.outputKafkaClusterDescription(cmd, &cluster, true)
}

func (c *clusterCommand) getLkcForDescribe(args []string) (string, error) {
	if len(args) > 0 {
		if resource.LookupType(args[0]) != resource.KafkaCluster {
			return "", fmt.Errorf(errors.KafkaClusterMissingPrefixErrorMsg, args[0])
		}
		return args[0], nil
	}

	clusterId := c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
	if clusterId == "" {
		return "", errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions)
	}

	return clusterId, nil
}

func (c *clusterCommand) outputKafkaClusterDescription(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster, getTopicCount bool) error {
	out := convertClusterToDescribeStruct(cluster, c.Context)

	if getTopicCount {
		topicCount, err := c.getTopicCountForKafkaCluster(cmd, cluster)
		// topicCount is 0 when err != nil, and will be omitted by `omitempty`
		if err != nil {
			log.CliLogger.Infof("The topic count will be omitted as Kafka topics for this cluster could not be retrieved: %v", err)
		}
		out.TopicCount = topicCount
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(getKafkaClusterDescribeFields(cluster, basicDescribeFields, getTopicCount))
	return table.Print()
}

func convertClusterToDescribeStruct(cluster *cmkv2.CmkV2Cluster, ctx *config.Context) *describeStruct {
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
		Cloud:              strings.ToLower(cluster.Spec.GetCloud()),
		Region:             cluster.Spec.GetRegion(),
		Availability:       ccloudv2.ToLower(cluster.Spec.GetAvailability()),
		Network:            cluster.Spec.Network.GetId(),
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

func (c *clusterCommand) getTopicCountForKafkaCluster(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster) (int, error) {
	if getCmkClusterStatus(cluster) == ccloudv2.StatusProvisioning {
		return 0, nil
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return 0, err
	}
	topics, err := kafkaREST.CloudClient.ListKafkaTopics()
	if err != nil {
		return 0, err
	}

	return len(topics.Data), nil
}
