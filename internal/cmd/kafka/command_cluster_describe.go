package kafka

import (
	"context"
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	basicDescribeFields                = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}
	basicDescribeFieldsWithApiEndpoint = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "ApiEndpoint", "RestEndpoint"}
	basicDescribeFieldsWithKAPI        = append(basicDescribeFields, "KAPI")
)

type describeStruct struct {
	Id                 string `human:"ID" structured:"id"`
	Name               string `human:"Name" structured:"name"`
	Type               string `human:"Type" structured:"type"`
	ClusterSize        int32  `human:"Cluster Size" structured:"cluster_size"`
	PendingClusterSize int32  `human:"Pending Cluster Size" structured:"pending_cluster_size"`
	NetworkIngress     int32  `human:"Ingress" structured:"ingress"`
	NetworkEgress      int32  `human:"Egress" structured:"egress"`
	Storage            string `human:"Storage" structured:"storage"`
	ServiceProvider    string `human:"Provider" structured:"provider"`
	Region             string `human:"Region" structured:"region"`
	Availability       string `human:"Availability" structured:"availability"`
	Status             string `human:"Status" structured:"status"`
	Endpoint           string `human:"Endpoint" structured:"endpoint"`
	ApiEndpoint        string `human:"API Endpoint" structured:"api_endpoint"`
	EncryptionKeyId    string `human:"Encryption Key ID" structured:"encryption_key_id"`
	RestEndpoint       string `human:"REST Endpoint" structured:"rest_endpoint"`
	KAPI               string `human:"KAPI" structured:"kapi"`
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
	out := convertClusterToDescribeStruct(cluster)
	filter := getKafkaClusterDescribeFields(cluster, basicDescribeFields)

	if all { // expose KAPI when --all flag is set
		kAPI, err := c.getCmkClusterApiEndpoint(cluster)
		if err != nil {
			return err
		}
		out.KAPI = kAPI
		filter = getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithKAPI)
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(filter)
	return table.Print()
}

func (c *clusterCommand) outputKafkaClusterDescription(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster) error {
	kAPI, err := c.getCmkClusterApiEndpoint(cluster)
	if err != nil {
		return err
	}

	out := convertClusterToDescribeStruct(cluster)
	out.ApiEndpoint = kAPI

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithApiEndpoint))
	return table.Print()
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

func getKafkaClusterDescribeFields(cluster *cmkv2.CmkV2Cluster, basicFields []string) []string {
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
