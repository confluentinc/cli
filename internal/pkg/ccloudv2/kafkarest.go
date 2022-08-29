package ccloudv2

import (
	"context"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
)

type KafkaRestClient struct {
	*kafkarestv3.APIClient
	AuthToken string
}

func NewKafkaRestClient(url, userAgent string, unsafeTrace bool, authToken string) *KafkaRestClient {
	cfg := kafkarestv3.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = kafkarestv3.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &KafkaRestClient{
		APIClient: kafkarestv3.NewAPIClient(cfg),
		AuthToken: authToken,
	}
}

func (c *KafkaRestClient) GetUrl() string {
	return c.GetConfig().Servers[0].URL
}

func (c *KafkaRestClient) context() context.Context {
	return context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, c.AuthToken)
}

func (c *KafkaRestClient) CreateKafkaAcls(clusterId string, data kafkarestv3.CreateAclRequestData) (*http.Response, error) {
	req := c.ACLV3Api.CreateKafkaAcls(c.context(), clusterId).CreateAclRequestData(data)
	return c.ACLV3Api.CreateKafkaAclsExecute(req)
}

func (c *KafkaRestClient) GetKafkaAcls(clusterId string, acl *schedv1.ACLBinding) (kafkarestv3.AclDataList, *http.Response, error) {
	req := c.ACLV3Api.GetKafkaAcls(c.context(), clusterId).Host(acl.Entry.Host).Principal(acl.Entry.Principal).ResourceName(acl.Pattern.Name)

	if acl.Entry.Operation != schedv1.ACLOperations_UNKNOWN {
		req.Operation(acl.Entry.Operation.String())
	}

	if acl.Pattern.PatternType != schedv1.PatternTypes_UNKNOWN {
		req.PatternType(acl.Pattern.PatternType.String())
	}

	if acl.Entry.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		req.Permission(acl.Entry.PermissionType.String())
	}

	if acl.Pattern.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		req.ResourceType(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	return c.ACLV3Api.GetKafkaAclsExecute(req)
}

func (c *KafkaRestClient) DeleteKafkaAcls(clusterId string, acl *schedv1.ACLFilter) (kafkarestv3.InlineResponse200, *http.Response, error) {
	req := c.ACLV3Api.DeleteKafkaAcls(c.context(), clusterId).Host(acl.EntryFilter.Host).Principal(acl.EntryFilter.Principal).ResourceName(acl.PatternFilter.Name)

	if acl.EntryFilter.Operation != schedv1.ACLOperations_UNKNOWN {
		req.Operation(acl.EntryFilter.Operation.String())
	}

	if acl.PatternFilter.PatternType != schedv1.PatternTypes_UNKNOWN {
		req.PatternType(acl.PatternFilter.PatternType.String())
	}

	if acl.EntryFilter.PermissionType != schedv1.ACLPermissionTypes_UNKNOWN {
		req.Permission(acl.EntryFilter.PermissionType.String())
	}

	if acl.PatternFilter.ResourceType != schedv1.ResourceTypes_UNKNOWN {
		req.ResourceType(kafkarestv3.AclResourceType(acl.PatternFilter.ResourceType.String()))
	}

	return c.ACLV3Api.DeleteKafkaAclsExecute(req)
}

func (c *KafkaRestClient) ListKafkaTopicConfigs(clusterId, topicName string) (kafkarestv3.TopicConfigDataList, *http.Response, error) {
	req := c.ConfigsV3Api.ListKafkaTopicConfigs(c.context(), clusterId, topicName)
	return c.ConfigsV3Api.ListKafkaTopicConfigsExecute(req)
}

func (c *KafkaRestClient) UpdateKafkaTopicConfigBatch(clusterId, topicName string, data kafkarestv3.AlterConfigBatchRequestData) (*http.Response, error) {
	req := c.ConfigsV3Api.UpdateKafkaTopicConfigBatch(c.context(), clusterId, topicName).AlterConfigBatchRequestData(data)
	return c.ConfigsV3Api.UpdateKafkaTopicConfigBatchExecute(req)
}

func (c *KafkaRestClient) GetKafkaConsumerGroup(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupData, *http.Response, error) {
	req := c.ConsumerGroupV3Api.GetKafkaConsumerGroup(c.context(), clusterId, consumerGroupId)
	return c.ConsumerGroupV3Api.GetKafkaConsumerGroupExecute(req)
}

func (c *KafkaRestClient) GetKafkaConsumerGroupLagSummary(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupLagSummaryData, *http.Response, error) {
	req := c.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummary(c.context(), clusterId, consumerGroupId)
	return c.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummaryExecute(req)
}

func (c *KafkaRestClient) ListKafkaConsumerGroups(clusterId string) (kafkarestv3.ConsumerGroupDataList, *http.Response, error) {
	req := c.ConsumerGroupV3Api.ListKafkaConsumerGroups(c.context(), clusterId)
	return c.ConsumerGroupV3Api.ListKafkaConsumerGroupsExecute(req)
}

func (c *KafkaRestClient) ListKafkaConsumerLags(clusterId, consumerGroupId string) (kafkarestv3.ConsumerLagDataList, *http.Response, error) {
	req := c.ConsumerGroupV3Api.ListKafkaConsumerLags(c.context(), clusterId, consumerGroupId)
	return c.ConsumerGroupV3Api.ListKafkaConsumerLagsExecute(req)
}

func (c *KafkaRestClient) ListKafkaConsumers(clusterId, consumerGroupId string) (kafkarestv3.ConsumerDataList, *http.Response, error) {
	req := c.ConsumerGroupV3Api.ListKafkaConsumers(c.context(), clusterId, consumerGroupId)
	return c.ConsumerGroupV3Api.ListKafkaConsumersExecute(req)
}

func (c *KafkaRestClient) GetKafkaConsumerLag(clusterId, consumerGroupId, topicName string, partitionId int32) (kafkarestv3.ConsumerLagData, *http.Response, error) {
	req := c.PartitionV3Api.GetKafkaConsumerLag(c.context(), clusterId, consumerGroupId, topicName, partitionId)
	return c.PartitionV3Api.GetKafkaConsumerLagExecute(req)
}

func (c *KafkaRestClient) ListKafkaPartitions(clusterId, topicName string) (kafkarestv3.PartitionDataList, *http.Response, error) {
	req := c.PartitionV3Api.ListKafkaPartitions(c.context(), clusterId, topicName)
	return c.PartitionV3Api.ListKafkaPartitionsExecute(req)
}

func (c *KafkaRestClient) CreateKafkaTopic(clusterId string, data kafkarestv3.CreateTopicRequestData) (kafkarestv3.TopicData, *http.Response, error) {
	req := c.TopicV3Api.CreateKafkaTopic(c.context(), clusterId).CreateTopicRequestData(data)
	return c.TopicV3Api.CreateKafkaTopicExecute(req)
}

func (c *KafkaRestClient) DeleteKafkaTopic(clusterId, topicName string) (*http.Response, error) {
	req := c.TopicV3Api.DeleteKafkaTopic(c.context(), clusterId, topicName)
	return c.TopicV3Api.DeleteKafkaTopicExecute(req)
}

func (c *KafkaRestClient) ListKafkaTopics(clusterId string) (kafkarestv3.TopicDataList, *http.Response, error) {
	req := c.TopicV3Api.ListKafkaTopics(c.context(), clusterId)
	return c.TopicV3Api.ListKafkaTopicsExecute(req)
}
