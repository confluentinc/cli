package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

const (
	BadRequestErrorCode              = 40002
	UnknownTopicOrPartitionErrorCode = 40403
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

func (c *KafkaRestClient) BatchCreateKafkaAcls(clusterId string, list kafkarestv3.CreateAclRequestDataList) (*http.Response, error) {
	return c.ACLV3Api.BatchCreateKafkaV3Acls(c.context(), clusterId).CreateAclRequestDataList(list).Execute()
}

func (c *KafkaRestClient) CreateKafkaAcls(clusterId string, data kafkarestv3.CreateAclRequestData) (*http.Response, error) {
	return c.ACLV3Api.CreateKafkaAcls(c.context(), clusterId).CreateAclRequestData(data).Execute()
}

func (c *KafkaRestClient) GetKafkaAcls(clusterId string, acl *ccstructs.ACLBinding) (kafkarestv3.AclDataList, *http.Response, error) {
	req := c.ACLV3Api.GetKafkaAcls(c.context(), clusterId).Host(acl.Entry.Host).Principal(acl.Entry.Principal).ResourceName(acl.Pattern.Name)

	if acl.Entry.Operation != ccstructs.ACLOperations_UNKNOWN {
		req = req.Operation(acl.Entry.Operation.String())
	}

	if acl.Pattern.PatternType != ccstructs.PatternTypes_UNKNOWN {
		req = req.PatternType(acl.Pattern.PatternType.String())
	}

	if acl.Entry.PermissionType != ccstructs.ACLPermissionTypes_UNKNOWN {
		req = req.Permission(acl.Entry.PermissionType.String())
	}

	if acl.Pattern.ResourceType != ccstructs.ResourceTypes_UNKNOWN {
		req = req.ResourceType(kafkarestv3.AclResourceType(acl.Pattern.ResourceType.String()))
	}

	return req.Execute()
}

func (c *KafkaRestClient) DeleteKafkaAcls(clusterId string, acl *ccstructs.ACLFilter) (kafkarestv3.InlineResponse200, *http.Response, error) {
	req := c.ACLV3Api.DeleteKafkaAcls(c.context(), clusterId).Host(acl.EntryFilter.Host).Principal(acl.EntryFilter.Principal).ResourceName(acl.PatternFilter.Name)

	if acl.EntryFilter.Operation != ccstructs.ACLOperations_UNKNOWN {
		req = req.Operation(acl.EntryFilter.Operation.String())
	}

	if acl.PatternFilter.PatternType != ccstructs.PatternTypes_UNKNOWN {
		req = req.PatternType(acl.PatternFilter.PatternType.String())
	}

	if acl.EntryFilter.PermissionType != ccstructs.ACLPermissionTypes_UNKNOWN {
		req = req.Permission(acl.EntryFilter.PermissionType.String())
	}

	if acl.PatternFilter.ResourceType != ccstructs.ResourceTypes_UNKNOWN {
		req = req.ResourceType(kafkarestv3.AclResourceType(acl.PatternFilter.ResourceType.String()))
	}

	return req.Execute()
}

func (c *KafkaRestClient) CreateKafkaLink(clusterId, linkName string, validateLink, validateOnly bool, data kafkarestv3.CreateLinkRequestData) (*http.Response, error) {
	return c.ClusterLinkingV3Api.CreateKafkaLink(c.context(), clusterId).LinkName(linkName).ValidateLink(validateLink).ValidateOnly(validateOnly).CreateLinkRequestData(data).Execute()
}

func (c *KafkaRestClient) CreateKafkaMirrorTopic(clusterId, linkName string, data kafkarestv3.CreateMirrorTopicRequestData) (*http.Response, error) {
	return c.ClusterLinkingV3Api.CreateKafkaMirrorTopic(c.context(), clusterId, linkName).CreateMirrorTopicRequestData(data).Execute()
}

func (c *KafkaRestClient) DeleteKafkaLink(clusterId, linkName string) (*http.Response, error) {
	return c.ClusterLinkingV3Api.DeleteKafkaLink(c.context(), clusterId, linkName).Execute()
}

func (c *KafkaRestClient) ListKafkaLinkConfigs(clusterId, linkName string) (kafkarestv3.ListLinkConfigsResponseDataList, *http.Response, error) {
	return c.ClusterLinkingV3Api.ListKafkaLinkConfigs(c.context(), clusterId, linkName).Execute()
}

func (c *KafkaRestClient) ListKafkaLinks(clusterId string) (kafkarestv3.ListLinksResponseDataList, *http.Response, error) {
	return c.ClusterLinkingV3Api.ListKafkaLinks(c.context(), clusterId).Execute()
}

func (c *KafkaRestClient) ListKafkaMirrorTopicsUnderLink(clusterId, linkName string) (kafkarestv3.ListMirrorTopicsResponseDataList, *http.Response, error) {
	return c.ClusterLinkingV3Api.ListKafkaMirrorTopicsUnderLink(c.context(), clusterId, linkName).Execute()
}

func (c *KafkaRestClient) UpdateKafkaLinkConfigBatch(clusterId, linkName string, data kafkarestv3.AlterConfigBatchRequestData) (*http.Response, error) {
	return c.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(c.context(), clusterId, linkName).AlterConfigBatchRequestData(data).Execute()
}

func (c *KafkaRestClient) ListKafkaTopicConfigs(clusterId, topicName string) (kafkarestv3.TopicConfigDataList, error) {
	res, httpResp, err := c.ConfigsV3Api.ListKafkaTopicConfigs(c.context(), clusterId, topicName).Execute()
	if err != nil {
		if restErr, err := kafkarest.ParseOpenAPIErrorCloud(err); err == nil {
			if restErr.Code == UnknownTopicOrPartitionErrorCode {
				return kafkarestv3.TopicConfigDataList{}, fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
			}
		}
	}
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaTopicConfigBatch(clusterId, topicName string, data kafkarestv3.AlterConfigBatchRequestData) (*http.Response, error) {
	return c.ConfigsV3Api.UpdateKafkaTopicConfigBatch(c.context(), clusterId, topicName).AlterConfigBatchRequestData(data).Execute()
}

func (c *KafkaRestClient) GetKafkaConsumerGroup(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupData, *http.Response, error) {
	return c.ConsumerGroupV3Api.GetKafkaConsumerGroup(c.context(), clusterId, consumerGroupId).Execute()
}

func (c *KafkaRestClient) GetKafkaConsumerGroupLagSummary(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupLagSummaryData, *http.Response, error) {
	return c.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummary(c.context(), clusterId, consumerGroupId).Execute()
}

func (c *KafkaRestClient) ListKafkaConsumerGroups(clusterId string) (kafkarestv3.ConsumerGroupDataList, *http.Response, error) {
	return c.ConsumerGroupV3Api.ListKafkaConsumerGroups(c.context(), clusterId).Execute()
}

func (c *KafkaRestClient) ListKafkaConsumerLags(clusterId, consumerGroupId string) (kafkarestv3.ConsumerLagDataList, *http.Response, error) {
	return c.ConsumerGroupV3Api.ListKafkaConsumerLags(c.context(), clusterId, consumerGroupId).Execute()
}

func (c *KafkaRestClient) ListKafkaConsumers(clusterId, consumerGroupId string) (kafkarestv3.ConsumerDataList, *http.Response, error) {
	return c.ConsumerGroupV3Api.ListKafkaConsumers(c.context(), clusterId, consumerGroupId).Execute()
}

func (c *KafkaRestClient) GetKafkaConsumerLag(clusterId, consumerGroupId, topicName string, partitionId int32) (kafkarestv3.ConsumerLagData, *http.Response, error) {
	return c.PartitionV3Api.GetKafkaConsumerLag(c.context(), clusterId, consumerGroupId, topicName, partitionId).Execute()
}

func (c *KafkaRestClient) ListKafkaPartitions(clusterId, topicName string) (kafkarestv3.PartitionDataList, *http.Response, error) {
	return c.PartitionV3Api.ListKafkaPartitions(c.context(), clusterId, topicName).Execute()
}

func (c *KafkaRestClient) CreateKafkaTopic(clusterId string, data kafkarestv3.CreateTopicRequestData) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.CreateKafkaTopic(c.context(), clusterId).CreateTopicRequestData(data).Execute()
}

func (c *KafkaRestClient) DeleteKafkaTopic(clusterId, topicName string) (*http.Response, error) {
	return c.TopicV3Api.DeleteKafkaTopic(c.context(), clusterId, topicName).Execute()
}

func (c *KafkaRestClient) ListKafkaTopics(clusterId string) (kafkarestv3.TopicDataList, *http.Response, error) {
	return c.TopicV3Api.ListKafkaTopics(c.context(), clusterId).Execute()
}

func (c *KafkaRestClient) UpdateKafkaTopicPartitionCount(clusterId, topicName string, updatePartitionCountRequestData kafkarestv3.UpdatePartitionCountRequestData) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.UpdatePartitionCountKafkaTopic(c.context(), clusterId, topicName).UpdatePartitionCountRequestData(updatePartitionCountRequestData).Execute()
}

func (c *KafkaRestClient) GetKafkaTopic(clusterId, topicName string) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.GetKafkaTopic(c.context(), clusterId, topicName).Execute()
}
