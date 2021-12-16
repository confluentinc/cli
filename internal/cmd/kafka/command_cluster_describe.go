package kafka

import (
	"context"
	"strconv"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	basicDescribeFields                = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}
	basicDescribeFieldsWithApiEndpoint = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "ApiEndpoint", "RestEndpoint"}
	describeHumanRenames               = map[string]string{
		"NetworkIngress":  "Ingress",
		"NetworkEgress":   "Egress",
		"ServiceProvider": "Provider",
		"EncryptionKeyId": "Encryption Key ID"}
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
	ApiEndpoint        string
	EncryptionKeyId    string
	RestEndpoint       string
}

type describeStructWithKAPI struct {
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
	KAPI               string
}

func (c *clusterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe [id]",
		Short:             "Describe a Kafka cluster.",
		Long:              "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.describe),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().Bool("all", false, "List all properties of a Kafka cluster.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
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

	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: lkc}
	cluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}

	return outputKafkaClusterDescriptionWithKAPI(cmd, cluster, all)
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

func outputKafkaClusterDescriptionWithKAPI(cmd *cobra.Command, cluster *schedv1.KafkaCluster, all bool) error {
	fields := basicDescribeFields
	structureRenames := describeStructuredRenames
	delete(structureRenames, "ApiEndpoint")
	if all { // expose KAPI when --all flag is set
		fields = append(fields, "KAPI")
		structureRenames["KAPI"] = "kapi"
	}
	return output.DescribeObject(cmd, convertClusterToDescribeStructWithKAPI(cluster), getKafkaClusterDescribeFields(cluster, fields), describeHumanRenames, structureRenames)
}

func convertClusterToDescribeStructWithKAPI(cluster *schedv1.KafkaCluster) *describeStructWithKAPI {
	clusterStorage := strconv.Itoa(int(cluster.Storage))
	if clusterStorage == "-1" || cluster.InfiniteStorage {
		clusterStorage = "Infinite"
	}

	return &describeStructWithKAPI{
		Id:                 cluster.Id,
		Name:               cluster.Name,
		Type:               cluster.Deployment.Sku.String(),
		ClusterSize:        cluster.Cku,
		PendingClusterSize: cluster.PendingCku,
		NetworkIngress:     cluster.NetworkIngress,
		NetworkEgress:      cluster.NetworkEgress,
		Storage:            clusterStorage,
		ServiceProvider:    cluster.ServiceProvider,
		Region:             cluster.Region,
		Availability:       durabilityToAvaiablityNameMap[cluster.Durability.String()],
		Status:             cluster.Status.String(),
		Endpoint:           cluster.Endpoint,
		EncryptionKeyId:    cluster.EncryptionKeyId,
		RestEndpoint:       cluster.RestEndpoint,
		KAPI:               cluster.ApiEndpoint,
	}
}

func outputKafkaClusterDescription(cmd *cobra.Command, cluster *schedv1.KafkaCluster) error {
	return output.DescribeObject(cmd, convertClusterToDescribeStruct(cluster), getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithApiEndpoint), describeHumanRenames, describeStructuredRenames)
}

func convertClusterToDescribeStruct(cluster *schedv1.KafkaCluster) *describeStruct {
	clusterStorage := strconv.Itoa(int(cluster.Storage))
	if clusterStorage == "-1" || cluster.InfiniteStorage {
		clusterStorage = "Infinite"
	}

	return &describeStruct{
		Id:                 cluster.Id,
		Name:               cluster.Name,
		Type:               cluster.Deployment.Sku.String(),
		ClusterSize:        cluster.Cku,
		PendingClusterSize: cluster.PendingCku,
		NetworkIngress:     cluster.NetworkIngress,
		NetworkEgress:      cluster.NetworkEgress,
		Storage:            clusterStorage,
		ServiceProvider:    cluster.ServiceProvider,
		Region:             cluster.Region,
		Availability:       durabilityToAvaiablityNameMap[cluster.Durability.String()],
		Status:             cluster.Status.String(),
		Endpoint:           cluster.Endpoint,
		ApiEndpoint:        cluster.ApiEndpoint,
		EncryptionKeyId:    cluster.EncryptionKeyId,
		RestEndpoint:       cluster.RestEndpoint,
	}
}

func getKafkaClusterDescribeFields(cluster *schedv1.KafkaCluster, basicFields []string) []string {
	describeFields := basicFields
	if isDedicated(cluster) {
		describeFields = append(describeFields, "ClusterSize")
		if isExpanding(cluster) || isShrinking(cluster) {
			describeFields = append(describeFields, "PendingClusterSize")
		}
		if cluster.EncryptionKeyId != "" {
			describeFields = append(describeFields, "EncryptionKeyId")
		}
	}
	return describeFields
}

func isDedicated(cluster *schedv1.KafkaCluster) bool {
	return cluster.Deployment.Sku == productv1.Sku_DEDICATED
}

func isExpanding(cluster *schedv1.KafkaCluster) bool {
	return cluster.Status == schedv1.ClusterStatus_EXPANDING || cluster.PendingCku > cluster.Cku
}

func isShrinking(cluster *schedv1.KafkaCluster) bool {
	return cluster.Status == schedv1.ClusterStatus_SHRINKING || (cluster.PendingCku < cluster.Cku && cluster.PendingCku != 0)
}
