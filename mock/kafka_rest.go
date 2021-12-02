package mock

import (
	"context"
	nethttp "net/http"

	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/config/v1"
)

// Compile-time check interface adherence
var _ krsdk.TopicApi = (*Topic)(nil)

type Topic struct {
}

func NewTopicMock() *Topic {
	return &Topic{}
}

func (m *Topic) ClustersClusterIdTopicsGet(_ context.Context, _clusterId string) (krsdk.TopicDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 200,
	}
	return krsdk.TopicDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.TopicData{
			{
				Kind:                   "",
				Metadata:               krsdk.ResourceMetadata{},
				ClusterId:              _clusterId,
				TopicName:              "NAME",
				IsInternal:             false,
				ReplicationFactor:      0,
				Partitions:             krsdk.Relationship{},
				Configs:                krsdk.Relationship{},
				PartitionReassignments: krsdk.Relationship{},
			},
		},
	}, httpResp, nil
}

func (m *Topic) ClustersClusterIdTopicsPost(_ context.Context, _ string, _ *krsdk.ClustersClusterIdTopicsPostOpts) (krsdk.TopicData, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 201,
	}
	return krsdk.TopicData{}, httpResp, nil
}

func (m *Topic) ClustersClusterIdTopicsTopicNameDelete(_ context.Context, _ string, _ string) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 204,
	}
	return httpResp, nil
}

func (m *Topic) ClustersClusterIdTopicsTopicNameGet(_ context.Context, _ string, _ string) (krsdk.TopicData, *nethttp.Response, error) {
	return krsdk.TopicData{}, nil, nil
}

// Compile-time check interface adherence
var _ krsdk.ACLApi = (*ACL)(nil)

type ACL struct {
}

func NewACLMock() *ACL {
	return &ACL{}
}

func (m *ACL) ClustersClusterIdAclsDelete(_ context.Context, _ string, _ *krsdk.ClustersClusterIdAclsDeleteOpts) (krsdk.InlineResponse200, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 200,
	}
	return krsdk.InlineResponse200{
		Data: []krsdk.AclData{
			{},
		},
	}, httpResp, nil
}

func (m *ACL) ClustersClusterIdAclsGet(_ context.Context, _clusterId string, _ *krsdk.ClustersClusterIdAclsGetOpts) (krsdk.AclDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 200,
	}
	return krsdk.AclDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.AclData{
			{
				Kind:         "KIND",
				Metadata:     krsdk.ResourceMetadata{},
				ClusterId:    _clusterId,
				ResourceType: "TYPE",
				ResourceName: "NAME",
				PatternType:  "PATTERN",
				Principal:    "User:PRINCIPAL",
				Host:         "HOST",
				Operation:    "OP",
				Permission:   "PERMISSION",
			},
		},
	}, httpResp, nil
}

func (m *ACL) ClustersClusterIdAclsPost(_ context.Context, _ string, _ *krsdk.ClustersClusterIdAclsPostOpts) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 201,
	}
	return httpResp, nil
}

// Compile-time check interface adherence
var _ krsdk.ConsumerGroupApi = (*ConsumerGroup)(nil)

type ConsumerGroup struct {
	Expect chan interface{}
}

func NewConsumerGroupMock(expect chan interface{}) *ConsumerGroup {
	return &ConsumerGroup{expect}
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersConsumerIdAssignmentsGet(_ context.Context, _ string, _ string, _ string) (krsdk.ConsumerAssignmentDataList, *nethttp.Response, error) {
	panic("implement me")
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersConsumerIdAssignmentsTopicNamePartitionsPartitionIdGet(_ context.Context, _ string, _ string, _ string, _ string, _ int32) (krsdk.ConsumerAssignmentData, *nethttp.Response, error) {
	panic("implement me")
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersConsumerIdGet(_ context.Context, _ string, _ string, _ string) (krsdk.ConsumerData, *nethttp.Response, error) {
	panic("implement me")
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersGet(_ context.Context, clusterId string, consumerGroupId string) (krsdk.ConsumerDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	return krsdk.ConsumerDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ConsumerData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				ClusterId:       clusterId,
				ConsumerGroupId: consumerGroupId,
				ConsumerId:      "consumer-1",
				InstanceId:      nil,
				ClientId:        "client-1",
				Assignments:     krsdk.Relationship{},
			},
		},
	}, httpResp, nil

}

type GroupMatcher struct {
	ConsumerGroupId string
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdGet(_ context.Context, clusterId string, consumerGroupId string) (krsdk.ConsumerGroupData, *nethttp.Response, error) {
	expect := <-c.Expect
	matcher := expect.(GroupMatcher)
	if err := assertEqualValues(consumerGroupId, matcher.ConsumerGroupId); err != nil {
		return krsdk.ConsumerGroupData{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	return krsdk.ConsumerGroupData{
		Kind:              "",
		Metadata:          krsdk.ResourceMetadata{},
		ClusterId:         clusterId,
		ConsumerGroupId:   "consumer-group-1",
		IsSimple:          true,
		PartitionAssignor: "org.apache.kafka.clients.consumer.RoundRobinAssignor",
		State:             "STABLE",
		Coordinator:       krsdk.Relationship{},
		Consumer:          krsdk.Relationship{},
		LagSummary:        krsdk.Relationship{},
	}, httpResp, nil
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(_ context.Context, clusterId string, consumerGroupId string) (krsdk.ConsumerGroupLagSummaryData, *nethttp.Response, error) {
	expect := <-c.Expect
	matcher := expect.(GroupMatcher)
	if err := assertEqualValues(consumerGroupId, matcher.ConsumerGroupId); err != nil {
		return krsdk.ConsumerGroupLagSummaryData{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	optionalInstanceId := "instance-1"

	return krsdk.ConsumerGroupLagSummaryData{
		Kind:              "",
		Metadata:          krsdk.ResourceMetadata{},
		ClusterId:         clusterId,
		ConsumerGroupId:   consumerGroupId,
		MaxLagConsumerId:  "consumer-1",
		MaxLagInstanceId:  &optionalInstanceId,
		MaxLagClientId:    "client-1",
		MaxLagTopicName:   "topic-1",
		MaxLagPartitionId: 0,
		MaxLag:            100,
		TotalLag:          110,
		MaxLagConsumer:    krsdk.Relationship{},
		MaxLagPartition:   krsdk.Relationship{},
	}, httpResp, nil

}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsConsumerGroupIdLagsGet(_ context.Context, clusterId string, consumerGroupId string) (krsdk.ConsumerLagDataList, *nethttp.Response, error) {
	expect := <-c.Expect
	matcher := expect.(GroupMatcher)
	if err := assertEqualValues(consumerGroupId, matcher.ConsumerGroupId); err != nil {
		return krsdk.ConsumerLagDataList{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	optionalInstanceIds := []string{"instance-1", "instance-2"}

	return krsdk.ConsumerLagDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ConsumerLagData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				ClusterId:       clusterId,
				ConsumerGroupId: consumerGroupId,
				TopicName:       "topic-1",
				PartitionId:     1,
				CurrentOffset:   1,
				LogEndOffset:    101,
				Lag:             100,
				ConsumerId:      "consumer-1",
				InstanceId:      &optionalInstanceIds[0],
				ClientId:        "client-1",
			},
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				ClusterId:       clusterId,
				ConsumerGroupId: consumerGroupId,
				TopicName:       "topic-1",
				PartitionId:     2,
				CurrentOffset:   1,
				LogEndOffset:    11,
				Lag:             10,
				ConsumerId:      "consumer-2",
				InstanceId:      &optionalInstanceIds[1],
				ClientId:        "client-2",
			},
		},
	}, httpResp, nil
}

func (c ConsumerGroup) ClustersClusterIdConsumerGroupsGet(_ context.Context, clusterId string) (krsdk.ConsumerGroupDataList, *nethttp.Response, error) {
	// lkc-12345 is the id of the mock cluster set in v3/mock.go
	if err := assertEqualValues(clusterId, v1.MockKafkaClusterId()); err != nil {
		return krsdk.ConsumerGroupDataList{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	return krsdk.ConsumerGroupDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ConsumerGroupData{
			{
				Kind:              "",
				Metadata:          krsdk.ResourceMetadata{},
				ClusterId:         clusterId,
				ConsumerGroupId:   "consumer-group-1",
				IsSimple:          true,
				PartitionAssignor: "org.apache.kafka.clients.consumer.RoundRobinAssignor",
				State:             "STABLE",
				Coordinator:       krsdk.Relationship{},
				Consumer:          krsdk.Relationship{},
				LagSummary:        krsdk.Relationship{},
			},
		},
	}, httpResp, nil
}

// Compile-time check interface adherence
var _ krsdk.PartitionApi = (*Partition)(nil)

type Partition struct {
	Expect chan interface{}
}

func NewPartitionMock(expect chan interface{}) *Partition {
	return &Partition{expect}
}

type PartitionLagMatcher struct {
	ConsumerGroupId string
	TopicName       string
	PartitionId     int32
}

func (m *Partition) ClustersClusterIdConsumerGroupsConsumerGroupIdLagsTopicNamePartitionsPartitionIdGet(_ context.Context, clusterId string, consumerGroupId string, topicName string, partitionId int32) (krsdk.ConsumerLagData, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(PartitionLagMatcher)
	if err := assertEqualValues(consumerGroupId, matcher.ConsumerGroupId); err != nil {
		return krsdk.ConsumerLagData{}, nil, err
	}
	if err := assertEqualValues(topicName, matcher.TopicName); err != nil {
		return krsdk.ConsumerLagData{}, nil, err
	}
	if err := assertEqualValues(partitionId, matcher.PartitionId); err != nil {
		return krsdk.ConsumerLagData{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	optionalInstanceId := "instance-1"

	return krsdk.ConsumerLagData{
		Kind:            "",
		Metadata:        krsdk.ResourceMetadata{},
		ClusterId:       clusterId,
		ConsumerGroupId: consumerGroupId,
		TopicName:       topicName,
		PartitionId:     partitionId,
		CurrentOffset:   1,
		LogEndOffset:    101,
		Lag:             100,
		ConsumerId:      "consumer-1",
		InstanceId:      &optionalInstanceId,
		ClientId:        "client-1",
	}, httpResp, nil
}

func (m *Partition) ClustersClusterIdTopicsPartitionsReassignmentGet(_ context.Context, _ string) (krsdk.ReassignmentDataList, *nethttp.Response, error) {
	return krsdk.ReassignmentDataList{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsGet(_ context.Context, clusterId string, topicName string) (krsdk.PartitionDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 200,
	}
	return krsdk.PartitionDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.PartitionData{
			{
				Kind:         "",
				Metadata:     krsdk.ResourceMetadata{},
				ClusterId:    clusterId,
				TopicName:    topicName,
				PartitionId:  0,
				Leader:       krsdk.Relationship{},
				Replicas:     krsdk.Relationship{},
				Reassignment: krsdk.Relationship{},
			},
		},
	}, httpResp, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdGet(_ context.Context, _ string, _ string, _ int32) (krsdk.PartitionData, *nethttp.Response, error) {
	return krsdk.PartitionData{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReassignmentGet(_ context.Context, _ string, _ string, _ int32) (krsdk.ReassignmentData, *nethttp.Response, error) {
	return krsdk.ReassignmentData{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsReassignmentGet(_ context.Context, _ string, _ string) (krsdk.ReassignmentDataList, *nethttp.Response, error) {
	return krsdk.ReassignmentDataList{}, nil, nil
}

// Compile-time check interface adherence
var _ krsdk.ReplicaApi = (*Replica)(nil)

type Replica struct {
}

func NewReplicaMock() *Replica {
	return &Replica{}
}

func (m *Replica) ClustersClusterIdBrokersBrokerIdPartitionReplicasGet(_ context.Context, _ string, _ int32) (krsdk.ReplicaDataList, *nethttp.Response, error) {
	return krsdk.ReplicaDataList{}, nil, nil
}

func (m *Replica) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasBrokerIdGet(_ context.Context, _ string, _ string, _ int32, _ int32) (krsdk.ReplicaData, *nethttp.Response, error) {
	return krsdk.ReplicaData{}, nil, nil
}

func (m *Replica) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(_ context.Context, clusterId string, topicName string, partitionId int32) (krsdk.ReplicaDataList, *nethttp.Response, error) {
	return krsdk.ReplicaDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ReplicaData{
			{
				Kind:        "",
				Metadata:    krsdk.ResourceMetadata{},
				ClusterId:   clusterId,
				TopicName:   topicName,
				PartitionId: partitionId,
				BrokerId:    42,
				IsLeader:    true,
				IsInSync:    true,
				Broker:      krsdk.Relationship{},
			},
		},
	}, nil, nil
}

type ReplicaStatus struct{}

func NewReplicaStatusMock() *ReplicaStatus {
	return &ReplicaStatus{}
}

func (m *ReplicaStatus) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatusGet(_ context.Context, clusterId string, topicName string, partitionId int32) (krsdk.ReplicaStatusDataList, *nethttp.Response, error) {
	return krsdk.ReplicaStatusDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ReplicaStatusData{
			{
				Kind:        "",
				Metadata:    krsdk.ResourceMetadata{},
				ClusterId:   clusterId,
				TopicName:   topicName,
				PartitionId: partitionId,
				BrokerId:    42,
				IsLeader:    true,
				IsInIsr:     true,
			},
		},
	}, nil, nil
}

func (m *ReplicaStatus) ClustersClusterIdTopicsTopicNamePartitionsReplicaStatusGet(_ context.Context, clusterId string, topicName string) (krsdk.ReplicaStatusDataList, *nethttp.Response, error) {
	return krsdk.ReplicaStatusDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ReplicaStatusData{
			{
				Kind:        "",
				Metadata:    krsdk.ResourceMetadata{},
				ClusterId:   clusterId,
				TopicName:   topicName,
				PartitionId: 0,
				BrokerId:    42,
				IsLeader:    true,
				IsInIsr:     true,
			},
		},
	}, nil, nil
}

// Compile-time check interface adherence
var _ krsdk.ConfigsApi = (*Configs)(nil)

type Configs struct {
}

func NewConfigsMock() *Configs {
	return &Configs{}
}

func (m *Configs) ClustersClusterIdBrokerConfigsGet(_ context.Context, _ string) (krsdk.ClusterConfigDataList, *nethttp.Response, error) {
	return krsdk.ClusterConfigDataList{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNameDelete(_ context.Context, _ string, _ string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNameGet(_ context.Context, _ string, _ string) (krsdk.ClusterConfigData, *nethttp.Response, error) {
	return krsdk.ClusterConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNamePut(_ context.Context, _ string, _ string, _ *krsdk.ClustersClusterIdBrokerConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsalterPost(_ context.Context, _ string, _ *krsdk.ClustersClusterIdBrokerConfigsalterPostOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsGet(_ context.Context, _ string, _ int32) (krsdk.BrokerConfigDataList, *nethttp.Response, error) {
	return krsdk.BrokerConfigDataList{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNameDelete(_ context.Context, _ string, _ int32, _ string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNameGet(_ context.Context, _ string, _ int32, _ string) (krsdk.BrokerConfigData, *nethttp.Response, error) {
	return krsdk.BrokerConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNamePut(_ context.Context, _ string, _ int32, _ string, _ *krsdk.ClustersClusterIdBrokersBrokerIdConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsalterPost(_ context.Context, _ string, _ int32, _ *krsdk.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsGet(_ context.Context, _clusterId string, topicName string) (krsdk.TopicConfigDataList, *nethttp.Response, error) {
	v := "configValue1"
	return krsdk.TopicConfigDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.TopicConfigData{
			{
				Kind:        "",
				Metadata:    krsdk.ResourceMetadata{},
				ClusterId:   _clusterId,
				Name:        "ConfigProperty1",
				Value:       &v,
				IsDefault:   false,
				IsReadOnly:  false,
				IsSensitive: false,
				Source:      "",
				Synonyms:    []krsdk.ConfigSynonymData{},
				TopicName:   topicName,
			},
		},
	}, nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNameDelete(_ context.Context, _ string, _ string, _ string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNameGet(_ context.Context, _ string, _ string, _ string) (krsdk.TopicConfigData, *nethttp.Response, error) {
	return krsdk.TopicConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNamePut(_ context.Context, _ string, _ string, _ string, _ *krsdk.ClustersClusterIdTopicsTopicNameConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsalterPost(_ context.Context, _ string, _ string, _ *krsdk.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 204,
	}
	return httpResp, nil
}

// Compile-time check interface adherence
var _ krsdk.ClusterLinkingApi = (*ClusterLinking)(nil)

type ClusterLinking struct {
	Expect chan interface{}
}

func (m *ClusterLinking) ClustersClusterIdLinksMirrorsGet(_ context.Context, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksMirrorsGetOpts) (krsdk.ListMirrorTopicsResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(ListMirrorMatcher)

	if err := assertEqualValues(string(localVarOptionals.MirrorStatus.Value().(krsdk.MirrorTopicStatus)), matcher.Status); err != nil {
		return krsdk.ListMirrorTopicsResponseDataList{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListMirrorTopicsResponseDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ListMirrorTopicsResponseData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				LinkName:        "link-1",
				MirrorTopicName: "mirror-topic-1",
				SourceTopicName: "src-topic-1",
				NumPartitions:   3,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       142857,
					},
					{
						Partition: 1,
						Lag:       285714,
					},
					{
						Partition: 2,
						Lag:       571428,
					},
				},
				MirrorStatus: "active",
				StateTimeMs:  44444444,
			},
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				LinkName:        "link-1",
				MirrorTopicName: "mirror-topic-2",
				SourceTopicName: "src-topic-2",
				MirrorStatus:    "active",
				StateTimeMs:     55555555,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       0,
					},
					{
						Partition: 1,
						Lag:       111111,
					},
				},
			},
		},
	}, httpResp, nil
}

func NewClusterLinkingMock(expect chan interface{}) *ClusterLinking {
	return &ClusterLinking{expect}
}

func (m *ClusterLinking) ClustersClusterIdLinksGet(_ context.Context, clusterId string) (krsdk.ListLinksResponseDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListLinksResponseDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ListLinksResponseData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				SourceClusterId: clusterId,
				LinkName:        "link-1",
				LinkId:          "LinkId",
				TopicNames:      []string{"topic-1", "topic-2", "topic-3"},
			},
		},
	}, httpResp, nil
}

type DeleteLinkConfigMatcher struct {
	LinkName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameConfigsConfigNameDelete(_ context.Context, _ string, linkName string, _ string) (*nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(DeleteLinkConfigMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusNoContent,
	}
	return httpResp, nil
}

type GetLinkConfigMatcher struct {
	LinkName   string
	ConfigName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameConfigsConfigNameGet(_ context.Context, clusterId string, linkName string, configName string) (krsdk.ListLinkConfigsResponseData, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(GetLinkConfigMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return krsdk.ListLinkConfigsResponseData{}, nil, err
	}
	if err := assertEqualValues(configName, matcher.ConfigName); err != nil {
		return krsdk.ListLinkConfigsResponseData{}, nil, err
	}
	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListLinkConfigsResponseData{
		Kind:      "",
		Metadata:  krsdk.ResourceMetadata{},
		ClusterId: clusterId,
		Name:      configName,
		Value:     "value",
		ReadOnly:  false,
		Sensitive: false,
		Source:    "DYNAMIC_CLUSTER_LINK_CONFIG",
		Synonyms:  nil,
		LinkName:  linkName,
	}, httpResp, nil
}

type UpdateLinkConfigMatcher struct {
	LinkName    string
	ConfigName  string
	ConfigValue string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameConfigsConfigNamePut(_ context.Context, _ string, linkName string, configName string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameConfigsConfigNamePutOpts) (*nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(UpdateLinkConfigMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return nil, err
	}
	if err := assertEqualValues(configName, matcher.ConfigName); err != nil {
		return nil, err
	}
	if err := assertEqualValues(localVarOptionals.UpdateLinkConfigRequestData.Value().(krsdk.UpdateLinkConfigRequestData).Value, matcher.ConfigValue); err != nil {
		return nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusNoContent,
	}
	return httpResp, nil
}

type ListLinkConfigMatcher struct {
	LinkName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameConfigsGet(_ context.Context, clusterId string, linkName string) (krsdk.ListLinkConfigsResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(DescribeLinkMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return krsdk.ListLinkConfigsResponseDataList{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListLinkConfigsResponseDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ListLinkConfigsResponseData{
			{
				Kind:      "",
				Metadata:  krsdk.ResourceMetadata{},
				ClusterId: clusterId,
				Name:      "consumer.offset.sync.enable",
				Value:     "value",
				ReadOnly:  false,
				Sensitive: false,
				Source:    "DYNAMIC_CLUSTER_LINK_CONFIG",
				Synonyms:  nil,
				LinkName:  linkName,
			},
		},
	}, httpResp, nil
}

type BatchUpdateLinkConfigMatcher struct {
	LinkName string
	Configs  map[string]string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameConfigsalterPut(_ context.Context, _ string, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameConfigsalterPutOpts) (*nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(BatchUpdateLinkConfigMatcher)
	for _, batchOp := range localVarOptionals.AlterConfigBatchRequestData.Value().(krsdk.AlterConfigBatchRequestData).Data {
		if err := assertEqualValues(*batchOp.Value, matcher.Configs[batchOp.Name]); err != nil {
			return nil, err
		}
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusNoContent,
	}
	return httpResp, nil
}

type DeleteLinkMatcher struct {
	LinkName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameDelete(_ context.Context, _ string, linkName string) (*nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(DeleteLinkMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusNoContent,
	}
	return httpResp, nil
}

type DescribeLinkMatcher struct {
	LinkName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameGet(_ context.Context, clusterId string, linkName string) (krsdk.ListLinksResponseData, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(DescribeLinkMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return krsdk.ListLinksResponseData{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListLinksResponseData{
		Kind:            "",
		Metadata:        krsdk.ResourceMetadata{},
		SourceClusterId: clusterId,
		LinkName:        linkName,
		LinkId:          "link-1",
		TopicNames:      []string{"topic-1", "topic-2", "topic-3"},
	}, httpResp, nil
}

type DescribeMirrorMatcher struct {
	LinkName        string
	MirrorTopicName string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorsMirrorTopicNameGet(_ context.Context, _ string, linkName string, mirrorTopicName string) (krsdk.ListMirrorTopicsResponseData, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(DescribeMirrorMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return krsdk.ListMirrorTopicsResponseData{}, nil, err
	}
	if err := assertEqualValues(mirrorTopicName, matcher.MirrorTopicName); err != nil {
		return krsdk.ListMirrorTopicsResponseData{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListMirrorTopicsResponseData{
		Kind:            "",
		Metadata:        krsdk.ResourceMetadata{},
		LinkName:        "link-1",
		MirrorTopicName: mirrorTopicName,
		SourceTopicName: mirrorTopicName,
		NumPartitions:   3,
		MirrorLags: []krsdk.MirrorLag{
			{
				Partition: 0,
				Lag:       142857,
			},
			{
				Partition: 1,
				Lag:       285714,
			},
			{
				Partition: 2,
				Lag:       571428,
			},
		},
		MirrorStatus: "active",
		StateTimeMs:  44444444,
	}, httpResp, nil
}

type AlterMirrorMatcher struct {
	LinkName         string
	MirrorTopicNames map[string]bool
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorsfailoverPost(_ context.Context, _ string, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorsfailoverPostOpts) (krsdk.AlterMirrorStatusResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(AlterMirrorMatcher)
	for _, topic := range localVarOptionals.AlterMirrorsRequestData.Value().(krsdk.AlterMirrorsRequestData).MirrorTopicNames {
		if err := assertEqualValues(true, matcher.MirrorTopicNames[topic]); err != nil {
			return krsdk.AlterMirrorStatusResponseDataList{}, nil, err
		}
	}

	return m.AlterMirrorResultResponse()
}

type ListMirrorMatcher struct {
	LinkName string
	Status   string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorsGet(_ context.Context, _ string, linkName string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorsGetOpts) (krsdk.ListMirrorTopicsResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(ListMirrorMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return krsdk.ListMirrorTopicsResponseDataList{}, nil, err
	}
	if err := assertEqualValues(string(localVarOptionals.MirrorStatus.Value().(krsdk.MirrorTopicStatus)), matcher.Status); err != nil {
		return krsdk.ListMirrorTopicsResponseDataList{}, nil, err
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}
	return krsdk.ListMirrorTopicsResponseDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.ListMirrorTopicsResponseData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				LinkName:        "link-1",
				MirrorTopicName: "mirror-topic-1",
				SourceTopicName: "src-topic-1",
				NumPartitions:   3,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       142857,
					},
					{
						Partition: 1,
						Lag:       285714,
					},
					{
						Partition: 2,
						Lag:       571428,
					},
				},
				MirrorStatus: "active",
				StateTimeMs:  44444444,
			},
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				LinkName:        "link-1",
				MirrorTopicName: "mirror-topic-2",
				SourceTopicName: "src-topic-2",
				MirrorStatus:    "active",
				StateTimeMs:     55555555,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       0,
					},
					{
						Partition: 1,
						Lag:       111111,
					},
				},
			},
		},
	}, httpResp, nil
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorspausePost(_ context.Context, _ string, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorspausePostOpts) (krsdk.AlterMirrorStatusResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(AlterMirrorMatcher)
	for _, topic := range localVarOptionals.AlterMirrorsRequestData.Value().(krsdk.AlterMirrorsRequestData).MirrorTopicNames {
		if err := assertEqualValues(true, matcher.MirrorTopicNames[topic]); err != nil {
			return krsdk.AlterMirrorStatusResponseDataList{}, nil, err
		}
	}

	return m.AlterMirrorResultResponse()
}

type CreateMirrorMatcher struct {
	LinkName        string
	SourceTopicName string
	Configs         map[string]string
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorsPost(_ context.Context, _ string, linkName string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorsPostOpts) (*nethttp.Response, error) {
	data := localVarOptionals.CreateMirrorTopicRequestData.Value().(krsdk.CreateMirrorTopicRequestData)
	expect := <-m.Expect
	matcher := expect.(CreateMirrorMatcher)
	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return nil, err
	}
	if err := assertEqualValues(data.SourceTopicName, matcher.SourceTopicName); err != nil {
		return nil, err
	}

	for _, config := range data.Configs {
		if err := assertEqualValues(*config.Value, matcher.Configs[config.Name]); err != nil {
			return nil, err
		}
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusNoContent,
	}
	return httpResp, nil
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorspromotePost(_ context.Context, _ string, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorspromotePostOpts) (krsdk.AlterMirrorStatusResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(AlterMirrorMatcher)
	for _, topic := range localVarOptionals.AlterMirrorsRequestData.Value().(krsdk.AlterMirrorsRequestData).MirrorTopicNames {
		if err := assertEqualValues(true, matcher.MirrorTopicNames[topic]); err != nil {
			return krsdk.AlterMirrorStatusResponseDataList{}, nil, err
		}
	}

	return m.AlterMirrorResultResponse()
}

func (m *ClusterLinking) ClustersClusterIdLinksLinkNameMirrorsresumePost(_ context.Context, _ string, _ string, localVarOptionals *krsdk.ClustersClusterIdLinksLinkNameMirrorsresumePostOpts) (krsdk.AlterMirrorStatusResponseDataList, *nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(AlterMirrorMatcher)
	for _, topic := range localVarOptionals.AlterMirrorsRequestData.Value().(krsdk.AlterMirrorsRequestData).MirrorTopicNames {
		if err := assertEqualValues(true, matcher.MirrorTopicNames[topic]); err != nil {
			return krsdk.AlterMirrorStatusResponseDataList{}, nil, err
		}
	}

	return m.AlterMirrorResultResponse()
}

type CreateLinkMatcher struct {
	LinkName        string
	ValidateLink    bool
	ValidateOnly    bool
	SourceClusterId string
	Configs         map[string]string
}

func (m *ClusterLinking) ClustersClusterIdLinksPost(_ context.Context, _ string, linkName string, localVarOptionals *krsdk.ClustersClusterIdLinksPostOpts) (*nethttp.Response, error) {
	expect := <-m.Expect
	matcher := expect.(CreateLinkMatcher)
	data := localVarOptionals.CreateLinkRequestData.Value().(krsdk.CreateLinkRequestData)

	if err := assertEqualValues(linkName, matcher.LinkName); err != nil {
		return nil, err
	}
	if err := assertEqualValues(data.SourceClusterId, matcher.SourceClusterId); err != nil {
		return nil, err
	}

	for _, config := range data.Configs {
		if err := assertEqualValues(*config.Value, matcher.Configs[config.Name]); err != nil {
			return nil, err
		}
	}

	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusCreated,
	}
	return httpResp, nil
}

func (m *ClusterLinking) AlterMirrorResultResponse() (krsdk.AlterMirrorStatusResponseDataList, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: nethttp.StatusOK,
	}

	errorMsg := "Not authorized"
	var errorCode int32 = 401

	return krsdk.AlterMirrorStatusResponseDataList{
		Kind:     "",
		Metadata: krsdk.ResourceCollectionMetadata{},
		Data: []krsdk.AlterMirrorStatusResponseData{
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				MirrorTopicName: "mirror-topic-1",
				ErrorMessage:    nil,
				ErrorCode:       nil,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       142857,
					},
					{
						Partition: 1,
						Lag:       285714,
					},
					{
						Partition: 2,
						Lag:       571428,
					},
				},
			},
			{
				Kind:            "",
				Metadata:        krsdk.ResourceMetadata{},
				MirrorTopicName: "mirror-topic-2",
				ErrorMessage:    &errorMsg,
				ErrorCode:       &errorCode,
				MirrorLags: []krsdk.MirrorLag{
					{
						Partition: 0,
						Lag:       142857,
					},
					{
						Partition: 1,
						Lag:       285714,
					},
					{
						Partition: 2,
						Lag:       571428,
					},
				},
			},
		},
	}, httpResp, nil
}
