package kafka

import (
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/cmk"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	basicDescribeFields  = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}
	describeHumanRenames = map[string]string{
		"ClusterSize":        "Cluster Size",
		"EncryptionKeyId":    "Encryption Key ID",
		"Id":                 "ID",
		"NetworkEgress":      "Egress",
		"NetworkIngress":     "Ingress",
		"PendingClusterSize": "Pending Cluster Size",
		"RestEndpoint":       "REST Endpoint",
		"ServiceProvider":    "Provider",
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
		"EncryptionKeyId":    "encryption_key_id",
		"RestEndpoint":       "rest_endpoint",
	}
)

var durabilityToAvaiablityNameMap = map[string]string{
	"LOW":  singleZone,
	"HIGH": multiZone,
}

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
	EncryptionKeyId    string
	RestEndpoint       string
}

func (c *clusterCommand) newDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id]",
		Short:             "Describe a Kafka cluster.",
		Long:              "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.describe),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	lkc, err := c.getLkcForDescribe(args)
	if err != nil {
		return err
	}

	cluster, _, err := cmk.DescribeKafkaCluster(c.CmkClient, lkc, c.EnvironmentId(), c.AuthToken())
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc)
	}

	return outputKafkaClusterDescription(cmd, &cluster)
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

func outputKafkaClusterDescription(cmd *cobra.Command, cluster *cmkv2.CmkV2Cluster) error {
	return output.DescribeObject(cmd, convertClusterToDescribeStruct(cluster), getKafkaClusterDescribeFields(cluster, basicDescribeFields), describeHumanRenames, describeStructuredRenames)
}

func convertClusterToDescribeStruct(cluster *cmkv2.CmkV2Cluster) *describeStruct {
	var clusterStorage string
	if !isBasic(cluster) {
		clusterStorage = "Infinite"
	} else {
		clusterStorage = "5 TB"
	}

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
		Availability:       availabilities[*cluster.Spec.Availability],
		Status:             getCmkClusterStatus(cluster),
		Endpoint:           cluster.Spec.GetKafkaBootstrapEndpoint(),
		// EncryptionKeyId:    cluster.EncryptionKeyId,
		RestEndpoint: cluster.Spec.GetHttpEndpoint(),
	}
}

func getKafkaClusterDescribeFields(cluster *cmkv2.CmkV2Cluster, basicFields []string) []string {
	describeFields := basicFields
	if isDedicated(cluster) {
		describeFields = append(describeFields, "ClusterSize")
		if isExpanding(cluster) || isShrinking(cluster) {
			describeFields = append(describeFields, "PendingClusterSize")
		}
		// waiting to be added!!!
		// if cluster.EncryptionKeyId != "" {
		// 	describeFields = append(describeFields, "EncryptionKeyId")
		// }
	}
	return describeFields
}
