package kafka

import (
	"fmt"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cckafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	cpkafkarestv3 "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/ccstructs"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/kafkausagelimits"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/serdes"
)

func toCreateTopicConfigs(topicConfigsMap map[string]string) []cckafkarestv3.ConfigData {
	topicConfigs := make([]cckafkarestv3.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = cckafkarestv3.ConfigData{
			Name:  k,
			Value: *cckafkarestv3.NewNullableString(&val),
		}
		i++
	}
	return topicConfigs
}

func toCreateTopicConfigsOnPrem(topicConfigsMap map[string]string) []cpkafkarestv3.ConfigData {
	topicConfigs := make([]cpkafkarestv3.ConfigData, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		val := v
		topicConfigs[i] = cpkafkarestv3.ConfigData{
			Name:  k,
			Value: &val,
		}
		i++
	}
	return topicConfigs
}

func toAlterConfigBatchRequestData(configsMap map[string]string) []cckafkarestv3.AlterConfigBatchRequestDataData {
	configs := make([]cckafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		val := val
		configs[i] = cckafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: *cckafkarestv3.NewNullableString(&val),
		}
		i++
	}
	return configs
}

func handleOpenApiError(httpResp *http.Response, err error, client *cpkafkarestv3.APIClient) error {
	if err == nil {
		return nil
	}

	if httpResp != nil {
		return kafkarest.NewError(client.GetConfig().BasePath, err, httpResp)
	}

	return err
}

func isClusterResizeInProgress(currentCluster *cmkv2.CmkV2Cluster) error {
	if currentCluster.Status.Phase == ccloudv2.StatusProvisioning {
		return fmt.Errorf("your cluster is still provisioning, so it can't be updated yet; please retry in a few minutes")
	}
	if isExpanding(currentCluster) {
		return fmt.Errorf("your cluster is expanding; please wait for that operation to complete before updating again")
	}
	if isShrinking(currentCluster) {
		return fmt.Errorf("your cluster is shrinking; please wait for that operation to complete before updating again")
	}
	return nil
}

func getCmkClusterIngressAndEgressMbps(currentMaxEcku int32, limits *kafkausagelimits.Limits) (int32, int32) {
	ingress, egress := limits.GetIngress(), limits.GetEgress()

	// Scale limits by cluster's max eCKU when limits are set per eCKU
	if limits.GetMaxEcku() != nil {
		if currentMaxEcku > 0 {
			return ingress * currentMaxEcku, egress * currentMaxEcku
		} else if limits.GetMaxEcku().Value > 0 {
			// Use default max ecku when currentMaxEcku is not set
			return ingress * limits.GetMaxEcku().Value, egress * limits.GetMaxEcku().Value
		}
	}

	return ingress, egress
}

func getCmkClusterType(cluster *cmkv2.CmkV2Cluster) string {
	if isBasic(cluster) {
		return ccstructs.Sku_name[2]
	}
	if isStandard(cluster) {
		return ccstructs.Sku_name[3]
	}
	if isDedicated(cluster) {
		return ccstructs.Sku_name[4]
	}
	if isEnterprise(cluster) {
		return ccstructs.Sku_name[6]
	}
	if isFreight(cluster) {
		return ccstructs.Sku_name[7]
	}
	return ccstructs.Sku_name[0] // UNKNOWN
}

func getCmkClusterSize(cluster *cmkv2.CmkV2Cluster) int32 {
	if isDedicated(cluster) {
		return *cluster.Status.Cku
	}
	return -1
}

func getCmkClusterPendingSize(cluster *cmkv2.CmkV2Cluster) int32 {
	if isDedicated(cluster) {
		return cluster.Spec.Config.CmkV2Dedicated.Cku
	}
	return -1
}

func getCmkMaxEcku(cluster *cmkv2.CmkV2Cluster) int32 {
	if isBasic(cluster) {
		if cluster.GetSpec().Config.CmkV2Basic.MaxEcku != nil {
			return cluster.GetSpec().Config.CmkV2Basic.GetMaxEcku()
		}
	} else if isStandard(cluster) {
		if cluster.GetSpec().Config.CmkV2Standard.MaxEcku != nil {
			return cluster.GetSpec().Config.CmkV2Standard.GetMaxEcku()
		}
	} else if isEnterprise(cluster) {
		if cluster.GetSpec().Config.CmkV2Enterprise.MaxEcku != nil {
			return cluster.GetSpec().Config.CmkV2Enterprise.GetMaxEcku()
		}
	} else if isFreight(cluster) {
		if cluster.GetSpec().Config.CmkV2Freight.MaxEcku != nil {
			return cluster.GetSpec().Config.CmkV2Freight.GetMaxEcku()
		}
	}
	return -1
}

func getCmkByokId(cluster *cmkv2.CmkV2Cluster) string {
	if isDedicated(cluster) && cluster.Spec.Byok != nil {
		return cluster.Spec.Byok.Id
	}
	return ""
}

func getCmkEncryptionKey(cluster *cmkv2.CmkV2Cluster) string {
	if isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.EncryptionKey != nil {
		return *cluster.Spec.Config.CmkV2Dedicated.EncryptionKey
	}
	return ""
}

func isBasic(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Basic != nil
}

func isStandard(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Standard != nil
}

func isEnterprise(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Enterprise != nil
}

func isFreight(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Freight != nil
}

func isDedicated(cluster *cmkv2.CmkV2Cluster) bool {
	return cluster.Spec.Config != nil && cluster.Spec.Config.CmkV2Dedicated != nil
}

func isExpanding(cluster *cmkv2.CmkV2Cluster) bool {
	return isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.Cku > *cluster.Status.Cku
}

func isShrinking(cluster *cmkv2.CmkV2Cluster) bool {
	return isDedicated(cluster) && cluster.Spec.Config.CmkV2Dedicated.Cku < *cluster.Status.Cku
}

func getCmkClusterStatus(cluster *cmkv2.CmkV2Cluster) string {
	if isExpanding(cluster) {
		return "EXPANDING"
	}
	if isShrinking(cluster) {
		return "SHRINKING"
	}
	if cluster.Status.Phase == "PROVISIONED" {
		return "UP"
	}
	return cluster.Status.Phase
}

func topicNameStrategy(topic, mode string) string {
	return fmt.Sprintf("%s-%s", topic, mode)
}

func newSchemaRegistryClient(srClientUrl, srClusterId string, srAuth serdes.SchemaRegistryAuth) (schemaregistry.Client, error) {
	var cfg *schemaregistry.Config
	switch {
	case srAuth.ApiKey != "" && srAuth.ApiSecret != "":
		cfg = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srAuth.ApiKey, srAuth.ApiSecret)
	case srAuth.Token != "":
		cfg = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, srAuth.Token, srClusterId, "")
	default:
		cfg = schemaregistry.NewConfig(srClientUrl)
		log.CliLogger.Info("initializing schema registry client with no authentication")
	}
	cfg.SslCaLocation = srAuth.CertificateAuthorityPath
	cfg.SslCertificateLocation = srAuth.ClientCertPath
	cfg.SslKeyLocation = srAuth.ClientKeyPath
	return schemaregistry.NewClient(cfg)
}

// returns the SR subject for (topic, mode) by querying the associations API with the Kafka cluster id
// as resource namespace. Falls backt o default TopicNameStrategy (<topic>-<mode>) if unmatched.
func lookupAssociatedSubject(client schemaregistry.Client, kafkaClusterId, topic, mode string) (string, bool, error) {
	associations, err := client.GetAssociationsByResourceName(topic, kafkaClusterId, "topic", []string{mode}, "", 0, -1)
	if err != nil {
		return "", false, err
	}
	if len(associations) == 0 {
		return "", false, nil
	}
	return associations[0].Subject, true, nil
}

func resolveSubject(client schemaregistry.Client, kafkaClusterId, topic, mode string) string {
	fallback := topic + "-" + mode
	if kafkaClusterId == "" || client == nil {
		return fallback
	}
	subject, found, err := lookupAssociatedSubject(client, kafkaClusterId, topic, mode)
	if err != nil {
		log.CliLogger.Tracef("subject resolution: associations lookup failed (topic=%q mode=%q clusterId=%q): %v; using %q", topic, mode, kafkaClusterId, err, fallback)
		return fallback
	}
	if !found {
		log.CliLogger.Tracef("subject resolution: no association for topic=%q mode=%q clusterId=%q; using %q", topic, mode, kafkaClusterId, fallback)
		return fallback
	}
	log.CliLogger.Tracef("subject resolution: resolved associated subject %q (topic=%q mode=%q clusterId=%q)", subject, topic, mode, kafkaClusterId)
	return subject
}

func resolveProduceSubject(srEndpoint, srClusterId, kafkaClusterId, topic, mode string, srAuth serdes.SchemaRegistryAuth) string {
	if kafkaClusterId != "" && srEndpoint != "" {
		if client, err := newSchemaRegistryClient(srEndpoint, srClusterId, srAuth); err == nil {
			return resolveSubject(client, kafkaClusterId, topic, mode)
		}
	}
	return topicNameStrategy(topic, mode)
}

// The Protobuf deserializer fetches its schema by subject, so on an associated topic it needs the
// associated (context-qualified) value subject. Otherwise, valueSubject stays empty to preserve default behaviour
// and the deserializer falls back to the topic-derived subject
func resolveAssociatedValueSubject(valueFormat, srEndpoint, srClusterId, kafkaClusterId, topic string, srAuth serdes.SchemaRegistryAuth) string {
	if !serdes.IsProtobufSchema(valueFormat) || kafkaClusterId == "" || srEndpoint == "" {
		return ""
	}
	client, err := newSchemaRegistryClient(srEndpoint, srClusterId, srAuth)
	if err != nil {
		return ""
	}
	return associatedValueSubject(client, kafkaClusterId, topic)
}

func associatedValueSubject(client schemaregistry.Client, kafkaClusterId, topic string) string {
	subject, found, err := lookupAssociatedSubject(client, kafkaClusterId, topic, "value")
	if err != nil || !found {
		return ""
	}
	log.CliLogger.Tracef("consumeCloud: resolved associated value subject %q", subject)
	return subject
}

func getLimitsForSku(cluster *cmkv2.CmkV2Cluster, usageLimits *kafkausagelimits.UsageLimits) *kafkausagelimits.Limits {
	if isDedicated(cluster) {
		return usageLimits.GetCkuLimit(cluster.Status.GetCku())
	}

	sku := getCmkClusterType(cluster)
	return usageLimits.GetTierLimit(sku).GetClusterLimits()
}
