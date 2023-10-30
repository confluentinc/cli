package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	"github.com/confluentinc/cli/v3/pkg/ccstructs"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
)

const (
	BadRequestErrorCode              = 40002
	UnknownTopicOrPartitionErrorCode = 40403
)

type KafkaRestClient struct {
	*kafkarestv3.APIClient
	AuthToken string
	ClusterId string
}

func NewKafkaRestClient(url, clusterId, userAgent, authToken string, unsafeTrace bool) *KafkaRestClient {
	cfg := kafkarestv3.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(nil, unsafeTrace)
	cfg.Servers = kafkarestv3.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &KafkaRestClient{
		APIClient: kafkarestv3.NewAPIClient(cfg),
		AuthToken: authToken,
		ClusterId: clusterId,
	}
}

func (c *KafkaRestClient) GetUrl() string {
	return c.GetConfig().Servers[0].URL
}

func (c *KafkaRestClient) context() context.Context {
	return context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, c.AuthToken)
}

func (c *KafkaRestClient) GetKafkaClusterConfig(name string) (kafkarestv3.ClusterConfigData, error) {
	res, httpResp, err := c.ConfigsV3Api.GetKafkaClusterConfig(c.context(), c.ClusterId, name).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaClusterConfigs() ([]kafkarestv3.ClusterConfigData, error) {
	res, httpResp, err := c.ConfigsV3Api.ListKafkaClusterConfigs(c.context(), c.ClusterId).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaClusterConfigs(req kafkarestv3.AlterConfigBatchRequestData) error {
	httpResp, err := c.ConfigsV3Api.UpdateKafkaClusterConfigs(c.context(), c.ClusterId).AlterConfigBatchRequestData(req).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) BatchCreateKafkaAcls(list kafkarestv3.CreateAclRequestDataList) error {
	httpResp, err := c.ACLV3Api.BatchCreateKafkaAcls(c.context(), c.ClusterId).CreateAclRequestDataList(list).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) CreateKafkaAcls(data kafkarestv3.CreateAclRequestData) error {
	httpResp, err := c.ACLV3Api.CreateKafkaAcls(c.context(), c.ClusterId).CreateAclRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaAcls(acl *ccstructs.ACLBinding) (kafkarestv3.AclDataList, error) {
	req := c.ACLV3Api.GetKafkaAcls(c.context(), c.ClusterId).Host(acl.GetEntry().GetHost()).Principal(acl.GetEntry().GetPrincipal()).ResourceName(acl.GetPattern().GetName())

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

func (c *KafkaRestClient) DeleteKafkaAcls(acl *ccstructs.ACLFilter) (kafkarestv3.InlineResponse200, error) {
	req := c.ACLV3Api.DeleteKafkaAcls(c.context(), c.ClusterId).Host(acl.EntryFilter.GetHost()).Principal(acl.EntryFilter.GetPrincipal()).ResourceName(acl.PatternFilter.GetName())

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

func (c *KafkaRestClient) CreateKafkaLink(linkName string, validateLink, validateOnly bool, data kafkarestv3.CreateLinkRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.CreateKafkaLink(c.context(), c.ClusterId).LinkName(linkName).ValidateLink(validateLink).ValidateOnly(validateOnly).CreateLinkRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) CreateKafkaMirrorTopic(linkName string, data kafkarestv3.CreateMirrorTopicRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.CreateKafkaMirrorTopic(c.context(), c.ClusterId, linkName).CreateMirrorTopicRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaMirrorTopics(status *kafkarestv3.MirrorTopicStatus) ([]kafkarestv3.ListMirrorTopicsResponseData, error) {
	req := c.ClusterLinkingV3Api.ListKafkaMirrorTopics(c.context(), c.ClusterId)

	if status != nil {
		req = req.MirrorStatus(*status)
	}

	res, httpResp, err := req.Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}

	return res.GetData(), nil
}

func (c *KafkaRestClient) ListKafkaMirrorTopicsUnderLink(linkName string, status *kafkarestv3.MirrorTopicStatus) ([]kafkarestv3.ListMirrorTopicsResponseData, error) {
	req := c.ClusterLinkingV3Api.ListKafkaMirrorTopicsUnderLink(c.context(), c.ClusterId, linkName)

	if status != nil {
		req = req.MirrorStatus(*status)
	}

	res, httpResp, err := req.Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}

	return res.GetData(), nil
}

func (c *KafkaRestClient) ReadKafkaMirrorTopic(linkName, mirrorTopicName string) (kafkarestv3.ListMirrorTopicsResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.ReadKafkaMirrorTopic(c.context(), c.ClusterId, linkName, mirrorTopicName).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaMirrorTopicsFailover(linkName string, validateOnly bool, data kafkarestv3.AlterMirrorsRequestData) ([]kafkarestv3.AlterMirrorStatusResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsFailover(c.context(), c.ClusterId, linkName).ValidateOnly(validateOnly).AlterMirrorsRequestData(data).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaMirrorTopicsPause(linkName string, validateOnly bool, data kafkarestv3.AlterMirrorsRequestData) ([]kafkarestv3.AlterMirrorStatusResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsPause(c.context(), c.ClusterId, linkName).ValidateOnly(validateOnly).AlterMirrorsRequestData(data).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaMirrorTopicsPromote(linkName string, validateOnly bool, data kafkarestv3.AlterMirrorsRequestData) ([]kafkarestv3.AlterMirrorStatusResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsPromote(c.context(), c.ClusterId, linkName).ValidateOnly(validateOnly).AlterMirrorsRequestData(data).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaMirrorTopicsResume(linkName string, validateOnly bool, data kafkarestv3.AlterMirrorsRequestData) ([]kafkarestv3.AlterMirrorStatusResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsResume(c.context(), c.ClusterId, linkName).ValidateOnly(validateOnly).AlterMirrorsRequestData(data).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) DeleteKafkaLink(linkName string) error {
	httpResp, err := c.ClusterLinkingV3Api.DeleteKafkaLink(c.context(), c.ClusterId, linkName).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaLink(linkName string) (kafkarestv3.ListLinksResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.GetKafkaLink(c.context(), c.ClusterId, linkName).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaLinkConfigs(linkName string) ([]kafkarestv3.ListLinkConfigsResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.ListKafkaLinkConfigs(c.context(), c.ClusterId, linkName).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) ListKafkaLinks() ([]kafkarestv3.ListLinksResponseData, error) {
	res, httpResp, err := c.ClusterLinkingV3Api.ListKafkaLinks(c.context(), c.ClusterId).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaLinkConfigBatch(linkName string, data kafkarestv3.AlterConfigBatchRequestData) error {
	httpResp, err := c.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(c.context(), c.ClusterId, linkName).AlterConfigBatchRequestData(data).Execute()
	return kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaTopicConfigs(topicName string) ([]kafkarestv3.TopicConfigData, error) {
	res, httpResp, err := c.ConfigsV3Api.ListKafkaTopicConfigs(c.context(), c.ClusterId, topicName).Execute()
	if err != nil {
		if restErr, err := kafkarest.ParseOpenAPIErrorCloud(err); err == nil {
			if restErr.Code == UnknownTopicOrPartitionErrorCode {
				return []kafkarestv3.TopicConfigData{}, fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
			}
		}
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) UpdateKafkaTopicConfigBatch(topicName string, data kafkarestv3.AlterConfigBatchRequestData) (*http.Response, error) {
	return c.ConfigsV3Api.UpdateKafkaTopicConfigBatch(c.context(), c.ClusterId, topicName).AlterConfigBatchRequestData(data).Execute()
}

func (c *KafkaRestClient) GetKafkaConsumerGroup(consumerGroupId string) (kafkarestv3.ConsumerGroupData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerGroup(c.context(), c.ClusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaConsumerGroupLagSummary(consumerGroupId string) (kafkarestv3.ConsumerGroupLagSummaryData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummary(c.context(), c.ClusterId, consumerGroupId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaConsumerGroups() ([]kafkarestv3.ConsumerGroupData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumerGroups(c.context(), c.ClusterId).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) ListKafkaConsumerLags(consumerGroupId string) ([]kafkarestv3.ConsumerLagData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumerLags(c.context(), c.ClusterId, consumerGroupId).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) ListKafkaConsumers(consumerGroupId string) ([]kafkarestv3.ConsumerData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.ListKafkaConsumers(c.context(), c.ClusterId, consumerGroupId).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) GetKafkaConsumerLag(consumerGroupId, topicName string, partitionId int32) (kafkarestv3.ConsumerLagData, error) {
	res, httpResp, err := c.ConsumerGroupV3Api.GetKafkaConsumerLag(c.context(), c.ClusterId, consumerGroupId, topicName, partitionId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaPartition(topicName string, partitionId int32) (kafkarestv3.PartitionData, error) {
	res, httpResp, err := c.PartitionV3Api.GetKafkaPartition(c.context(), c.ClusterId, topicName, partitionId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) ListKafkaPartitions(topicName string) ([]kafkarestv3.PartitionData, error) {
	res, httpResp, err := c.PartitionV3Api.ListKafkaPartitions(c.context(), c.ClusterId, topicName).Execute()
	if err != nil {
		return nil, kafkarest.NewError(c.GetUrl(), err, httpResp)
	}
	return res.GetData(), nil
}

func (c *KafkaRestClient) CreateKafkaTopic(data kafkarestv3.CreateTopicRequestData) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.CreateKafkaTopic(c.context(), c.ClusterId).CreateTopicRequestData(data).Execute()
}

func (c *KafkaRestClient) DeleteKafkaTopic(topicName string) (*http.Response, error) {
	return c.TopicV3Api.DeleteKafkaTopic(c.context(), c.ClusterId, topicName).Execute()
}

func (c *KafkaRestClient) ListKafkaTopics() (kafkarestv3.TopicDataList, error) {
	res, httpResp, err := c.TopicV3Api.ListKafkaTopics(c.context(), c.ClusterId).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) UpdateKafkaTopicPartitionCount(topicName string, updatePartitionCountRequestData kafkarestv3.UpdatePartitionCountRequestData) (kafkarestv3.TopicData, error) {
	res, httpResp, err := c.TopicV3Api.UpdatePartitionCountKafkaTopic(c.context(), c.ClusterId, topicName).UpdatePartitionCountRequestData(updatePartitionCountRequestData).Execute()
	return res, kafkarest.NewError(c.GetUrl(), err, httpResp)
}

func (c *KafkaRestClient) GetKafkaTopic(topicName string) (kafkarestv3.TopicData, *http.Response, error) {
	return c.TopicV3Api.GetKafkaTopic(c.context(), c.ClusterId, topicName).Execute()
}
