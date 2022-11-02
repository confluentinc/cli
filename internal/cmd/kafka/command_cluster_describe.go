package kafka

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	basicDescribeFields                = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}
	basicDescribeFieldsWithApiEndpoint = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "ApiEndpoint", "RestEndpoint"}
	basicDescribeFieldsWithKAPI        = append(basicDescribeFields, "KAPI")

	describeHumanRenames = map[string]string{
		"ApiEndpoint":        "API Endpoint",
		"ClusterSize":        "Cluster Size",
		"EncryptionKeyId":    "Encryption Key ID",
		"Id":                 "ID",
		"NetworkEgress":      "Egress",
		"NetworkIngress":     "Ingress",
		"PendingClusterSize": "Pending Cluster Size",
		"RestEndpoint":       "REST Endpoint",
		"ServiceProvider":    "Provider",
		"TopicCount":         "Topic Count",
	}
	describeStructuredRenames = map[string]string{
		"Id":                 "id",
		"Name":               "name",
		"Type":               "type",
		"ClusterSize":        "cluster_size",
		"PendingClusterSize": "pending_cluster_size",
		"NetworkIngress":     "ingress",
		"NetworkEgress":      "egress",
		"Storage":            "storage",
		"ServiceProvider":    "provider",
		"Region":             "region",
		"Availability":       "availability",
		"Status":             "status",
		"Endpoint":           "endpoint",
		"ApiEndpoint":        "api_endpoint",
		"EncryptionKeyId":    "encryption_key_id",
		"RestEndpoint":       "rest_endpoint",
		"KAPI":               "kapi",
		"TopicCount":         "topic_count",
	}
)

type describeStruct struct {
	Id                 string
	Name               string
	Type               string
	ClusterSize        int32
	PendingClusterSize int32
	NetworkIngress     int32
	NetworkEgress      int32
	Storage            string
	ServiceProvider    string
	Region             string
	Availability       string
	Status             string
	Endpoint           string
	ApiEndpoint        string
	EncryptionKeyId    string
	RestEndpoint       string
	KAPI               string
	TopicCount         int
}

func (c *clusterCommand) newDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id]",
		Short:             "Describe a Kafka cluster.",
		Long:              "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	cmd.Flags().Bool("all", false, "List all properties of a Kafka cluster.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	lkc, err := c.getLkcForDescribe(args)
	if err != nil {
		return err
	}

	ctx := c.AuthenticatedCLICommand.Context.Config.Context()
	c.AuthenticatedCLICommand.Context.Config.SetOverwrittenActiveKafka(ctx.KafkaClusterContext.GetActiveKafkaClusterId())
	ctx.KafkaClusterContext.SetActiveKafkaCluster(lkc)

	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, c.EnvironmentId())
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}

	return c.outputKafkaClusterDescriptionWithKAPI(cmd, &cluster, all)
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

func (c *clusterCommand) outputKafkaClusterDescriptionWithKAPI(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster, all bool) error {
	describeStruct := convertClusterToDescribeStruct(cluster)
	topicCount, err := c.getTopicCountForKafkaCluster(cluster)
	if err != nil {
		return err
	}
	describeStruct.TopicCount = topicCount

	if all { // expose KAPI when --all flag is set
		kAPI, err := c.getCmkClusterApiEndpoint(cluster)
		if err != nil {
			return err
		}
		describeStruct.KAPI = kAPI

		return output.DescribeObject(cmd, describeStruct, getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithKAPI, true), describeHumanRenames, describeStructuredRenames)
	}

	return output.DescribeObject(cmd, describeStruct, getKafkaClusterDescribeFields(cluster, basicDescribeFields, true), describeHumanRenames, describeStructuredRenames)
}

func (c *clusterCommand) outputKafkaClusterDescription(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster, getTopicCount bool) error {
	kAPI, err := c.getCmkClusterApiEndpoint(cluster)
	if err != nil {
		return err
	}
	describeStruct := convertClusterToDescribeStruct(cluster)
	describeStruct.ApiEndpoint = kAPI

	if getTopicCount {
		topicCount, err := c.getTopicCountForKafkaCluster(cluster)
		if err != nil {
			return err
		}
		describeStruct.TopicCount = topicCount
	}

	return output.DescribeObject(cmd, describeStruct, getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithApiEndpoint, getTopicCount), describeHumanRenames, describeStructuredRenames)
}

func convertClusterToDescribeStruct(cluster *cmkv2.CmkV2Cluster) *describeStruct {
	clusterStorage := getKafkaClusterStorage(cluster)
	ingress, egress := getCmkClusterIngressAndEgress(cluster)

	return &describeStruct{
		Id:                 *cluster.Id,
		Name:               *cluster.Spec.DisplayName,
		Type:               getCmkClusterType(cluster),
		ClusterSize:        getCmkClusterSize(cluster),
		PendingClusterSize: getCmkClusterPendingSize(cluster),
		NetworkIngress:     ingress,
		NetworkEgress:      egress,
		Storage:            clusterStorage,
		ServiceProvider:    strings.ToLower(*cluster.Spec.Cloud),
		Region:             *cluster.Spec.Region,
		Availability:       availabilitiesToHuman[*cluster.Spec.Availability],
		Status:             getCmkClusterStatus(cluster),
		Endpoint:           cluster.Spec.GetKafkaBootstrapEndpoint(),
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
	}

	if getTopicCount {
		describeFields = append(describeFields, "TopicCount")
	}

	return describeFields
}

func (c *clusterCommand) getCmkClusterApiEndpoint(cluster *cmkv2.CmkV2Cluster) (string, error) { // TODO: remove this function when KAPI is fully deprecated
	lkc := *cluster.Id
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: lkc}
	kafkaCluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return "", errors.CatchKafkaNotFoundError(err, lkc, nil)
	}
	return kafkaCluster.ApiEndpoint, nil
}

func (c *clusterCommand) getTopicCountForKafkaCluster(cluster *cmkv2.CmkV2Cluster) (int, error) {
	if getCmkClusterStatus(cluster) == ccloudv2.StatusProvisioning {
		return 0, nil
	}

	lkc := *cluster.Id
	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		topicGetResp, httpResp, err := kafkaREST.CloudClient.ListKafkaTopics(lkc)
		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			return 0, kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return 0, errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			return len(topicGetResp.Data), nil
		}
	}

	// Kafka REST is not available, fall back to KafkaAPI, to be deprecated
	req, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return 0, err
	}
	resp, err := c.Client.Kafka.ListTopics(context.Background(), req)
	return len(resp), err
}
