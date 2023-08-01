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
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
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

func (c *KafkaRestClient) GetKafkaClusterConfig(clusterId, name string) (kafkarestv3.ClusterConfigData, error) {
	res, httpResp, err := c.ConfigsV3Api.GetKafkaClusterConfig(c.context(), clusterId, name).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaClusterConfigs(clusterId string) (kafkarestv3.ClusterConfigDataList, error) {
	res, httpResp, err := c.ConfigsV3Api.ListKafkaClusterConfigs(c.context(), clusterId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaClusterConfigs(clusterId string, req kafkarestv3.AlterConfigBatchRequestData) error {
	httpResp, err := c.ConfigsV3Api.UpdateKafkaClusterConfigs(c.context(), clusterId).AlterConfigBatchRequestData(req).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) BatchCreateKafkaAcls(clusterId string, list kafkarestv3.CreateAclRequestDataList) error {
	httpResp, err := c.ACLV3Api.BatchCreateKafkaAcls(c.context(), clusterId).CreateAclRequestDataList(list).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) CreateKafkaAcls(clusterId string, data kafkarestv3.CreateAclRequestData) error {
	httpResp, err := c.ACLV3Api.CreateKafkaAcls(c.context(), clusterId).CreateAclRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaAcls(clusterId string, acl *ccstructs.ACLBinding) (kafkarestv3.AclDataList, error) {
	req := c.ACLV3Api.GetKafkaAcls(c.context(), clusterId).Host(acl.GetEntry().GetHost()).Principal(acl.GetEntry().GetPrincipal()).ResourceName(acl.GetPattern().GetName())

	if acl.GetEntry().GetOperation() != ccstructs.ACLOperations_UNKNOWN {
		req = req.Operation(acl.GetEntry().GetOperation().String())
	}

	if acl.GetPattern().GetPatternType() != ccstructs.PatternTypes_UNKNOWN {
		req = req.PatternType(acl.GetPattern().GetPatternType().String())
	}

	if acl.GetEntry().GetPermissionType() != ccstructs.ACLPermissionTypes_UNKNOWN {
		req = req.Permission(acl.GetEntry().GetPermissionType().String())
	}

	if acl.GetPattern().GetResourceType() != ccstructs.ResourceTypes_UNKNOWN {
		req = req.ResourceType(kafkarestv3.AclResourceType(acl.GetPattern().GetResourceType().String()))
	}

	res, httpResp, err := req.Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) DeleteKafkaAcls(clusterId string, acl *ccstructs.ACLFilter) (kafkarestv3.InlineResponse200, error) {
	req := c.ACLV3Api.DeleteKafkaAcls(c.context(), clusterId).Host(acl.EntryFilter.GetHost()).Principal(acl.EntryFilter.GetPrincipal()).ResourceName(acl.PatternFilter.GetName())

	if acl.EntryFilter.GetOperation() != ccstructs.ACLOperations_UNKNOWN {
		req = req.Operation(acl.EntryFilter.GetOperation().String())
	}

	if acl.PatternFilter.GetPatternType() != ccstructs.PatternTypes_UNKNOWN {
		req = req.PatternType(acl.PatternFilter.GetPatternType().String())
	}

	if acl.EntryFilter.GetPermissionType() != ccstructs.ACLPermissionTypes_UNKNOWN {
		req = req.Permission(acl.EntryFilter.GetPermissionType().String())
	}

	if acl.PatternFilter.GetResourceType() != ccstructs.ResourceTypes_UNKNOWN {
		req = req.ResourceType(kafkarestv3.AclResourceType(acl.PatternFilter.GetResourceType().String()))
	}

	res, httpResp, err := req.Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) CreateKafkaLink(clusterId, linkName string, validateLink, validateOnly bool, data kafkarestv3.CreateLinkRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.CreateKafkaLink(c.context(), clusterId).LinkName(linkName).ValidateLink(validateLink).ValidateOnly(validateOnly).CreateLinkRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) CreateKafkaMirrorTopic(clusterId, linkName string, data kafkarestv3.CreateMirrorTopicRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.CreateKafkaMirrorTopic(c.context(), clusterId, linkName).CreateMirrorTopicRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) DeleteKafkaLink(clusterId, linkName string) error {
	httpResp, err := c.ClusterLinkingV3Api.DeleteKafkaLink(c.context(), clusterId, linkName).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaLink(clusterId, linkName string) (kafkarestv3.ListLinksResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.GetKafkaLink(c.context(), clusterId, linkName).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaLinkConfigs(clusterId, linkName string) (kafkarestv3.ListLinkConfigsResponseDataList, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.ListKafkaLinkConfigs(c.context(), clusterId, linkName).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaLinks(clusterId string) (kafkarestv3.ListLinksResponseDataList, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.ListKafkaLinks(c.context(), clusterId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaLinkConfigBatch(clusterId, linkName string, data kafkarestv3.AlterConfigBatchRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(c.context(), clusterId, linkName).AlterConfigBatchRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
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

func (c *KafkaRestClient) GetKafkaConsumerGroup(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerGroup(c.context(), clusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaConsumerGroupLagSummary(clusterId, consumerGroupId string) (kafkarestv3.ConsumerGroupLagSummaryData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummary(c.context(), clusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaConsumerGroups(clusterId string) (kafkarestv3.ConsumerGroupDataList, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumerGroups(c.context(), clusterId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaConsumerLags(clusterId, consumerGroupId string) (kafkarestv3.ConsumerLagDataList, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumerLags(c.context(), clusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaConsumers(clusterId, consumerGroupId string) (kafkarestv3.ConsumerDataList, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumers(c.context(), clusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaConsumerLag(clusterId, consumerGroupId, topicName string, partitionId int32) (kafkarestv3.ConsumerLagData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerLag(c.context(), clusterId, consumerGroupId, topicName, partitionId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
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

func (c *KafkaRestClient) ListKafkaTopics(clusterId string) (kafkarestv3.TopicDataList, error) {
	res, httpResp, err := c.TopicV3Api.ListKafkaTopics(c.context(), clusterId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaTopicPartitionCount(clusterId, topicName string, updatePartitionCountRequestData kafkarestv3.UpdatePartitionCountRequestData) (kafkarestv3.TopicData, error) {
	res, httpResp, err := c.TopicV3Api.UpdatePartitionCountKafkaTopic(c.context(), clusterId, topicName).UpdatePartitionCountRequestData(updatePartitionCountRequestData).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaTopic(clusterId, topicName string) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.GetKafkaTopic(c.context(), clusterId, topicName).Execute()
}
