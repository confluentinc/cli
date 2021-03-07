package mock

import (
	"context"

	nethttp "net/http"

	krsdk "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

// Compile-time check interface adherence
var _ krsdk.TopicApi = (*Topic)(nil)

type Topic struct {
}

func NewTopicMock() *Topic {
	return &Topic{}
}

func (m *Topic) ClustersClusterIdTopicsGet(_ctx context.Context, _clusterId string) (krsdk.TopicDataList, *nethttp.Response, error) {
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

func (m *Topic) ClustersClusterIdTopicsPost(_ctx context.Context, _clusterId string, _localVarOptionals *krsdk.ClustersClusterIdTopicsPostOpts) (krsdk.TopicData, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 201,
	}
	return krsdk.TopicData{}, httpResp, nil
}

func (m *Topic) ClustersClusterIdTopicsTopicNameDelete(_ctx context.Context, _clusterId string, _topicName string) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 204,
	}
	return httpResp, nil
}

func (m *Topic) ClustersClusterIdTopicsTopicNameGet(_ctx context.Context, _clusterId string, _topicName string) (krsdk.TopicData, *nethttp.Response, error) {
	return krsdk.TopicData{}, nil, nil
}

// Compile-time check interface adherence
var _ krsdk.ACLApi = (*ACL)(nil)

type ACL struct {
}

func NewACLMock() *ACL {
	return &ACL{}
}

func (m *ACL) ClustersClusterIdAclsDelete(_ctx context.Context, _clusterId string, _localVarOptionals *krsdk.ClustersClusterIdAclsDeleteOpts) (krsdk.InlineResponse200, *nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 200,
	}
	return krsdk.InlineResponse200{
		Data: []krsdk.AclData{
			{},
		},
	}, httpResp, nil
}

func (m *ACL) ClustersClusterIdAclsGet(_ctx context.Context, _clusterId string, _localVarOptionals *krsdk.ClustersClusterIdAclsGetOpts) (krsdk.AclDataList, *nethttp.Response, error) {
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
				Principal:    "PRINCIPAL",
				Host:         "HOST",
				Operation:    "OP",
				Permission:   "PERMISSION",
			},
		},
	}, httpResp, nil
}

func (m *ACL) ClustersClusterIdAclsPost(_ctx context.Context, _clusterId string, _localVarOptionals *krsdk.ClustersClusterIdAclsPostOpts) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 201,
	}
	return httpResp, nil
}

// Compile-time check interface adherence
var _ krsdk.PartitionApi = (*Partition)(nil)

type Partition struct {
}

func NewPartitionMock() *Partition {
	return &Partition{}
}

func (m *Partition) ClustersClusterIdTopicsPartitionsReassignmentGet(_ctx context.Context, _clusterId string) (krsdk.ReassignmentDataList, *nethttp.Response, error) {
	return krsdk.ReassignmentDataList{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsGet(_ctx context.Context, clusterId string, topicName string) (krsdk.PartitionDataList, *nethttp.Response, error) {
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

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdGet(_ctx context.Context, _clusterId string, _topicName string, _partitionId int32) (krsdk.PartitionData, *nethttp.Response, error) {
	return krsdk.PartitionData{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReassignmentGet(_ctx context.Context, _clusterId string, _topicName string, _partitionId int32) (krsdk.ReassignmentData, *nethttp.Response, error) {
	return krsdk.ReassignmentData{}, nil, nil
}

func (m *Partition) ClustersClusterIdTopicsTopicNamePartitionsReassignmentGet(_ctx context.Context, _clusterId string, _topicName string) (krsdk.ReassignmentDataList, *nethttp.Response, error) {
	return krsdk.ReassignmentDataList{}, nil, nil
}

// Compile-time check interface adherence
var _ krsdk.ReplicaApi = (*Replica)(nil)

type Replica struct {
}

func NewReplicaMock() *Replica {
	return &Replica{}
}

func (m *Replica) ClustersClusterIdBrokersBrokerIdPartitionReplicasGet(_ctx context.Context, _clusterId string, _brokerId int32) (krsdk.ReplicaDataList, *nethttp.Response, error) {
	return krsdk.ReplicaDataList{}, nil, nil
}

func (m *Replica) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasBrokerIdGet(_ctx context.Context, _clusterId string, _topicName string, partitionId int32, brokerId int32) (krsdk.ReplicaData, *nethttp.Response, error) {
	return krsdk.ReplicaData{}, nil, nil
}

func (m *Replica) ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(_ctx context.Context, clusterId string, topicName string, partitionId int32) (krsdk.ReplicaDataList, *nethttp.Response, error) {
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

// Compile-time check interface adherence
var _ krsdk.ConfigsApi = (*Configs)(nil)

type Configs struct {
}

func NewConfigsMock() *Configs {
	return &Configs{}
}

func (m *Configs) ClustersClusterIdBrokerConfigsGet(_ctx context.Context, _clusterId string) (krsdk.ClusterConfigDataList, *nethttp.Response, error) {
	return krsdk.ClusterConfigDataList{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNameDelete(_ctx context.Context, _clusterId string, name string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNameGet(_ctx context.Context, _clusterId string, name string) (krsdk.ClusterConfigData, *nethttp.Response, error) {
	return krsdk.ClusterConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsNamePut(_ctx context.Context, _clusterId string, name string, localVarOptionals *krsdk.ClustersClusterIdBrokerConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokerConfigsalterPost(_ctx context.Context, _clusterId string, localVarOptionals *krsdk.ClustersClusterIdBrokerConfigsalterPostOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsGet(_ctx context.Context, _clusterId string, brokerId int32) (krsdk.BrokerConfigDataList, *nethttp.Response, error) {
	return krsdk.BrokerConfigDataList{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNameDelete(_ctx context.Context, _clusterId string, brokerId int32, name string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNameGet(_ctx context.Context, _clusterId string, brokerId int32, name string) (krsdk.BrokerConfigData, *nethttp.Response, error) {
	return krsdk.BrokerConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsNamePut(_ctx context.Context, _clusterId string, brokerId int32, name string, localVarOptionals *krsdk.ClustersClusterIdBrokersBrokerIdConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdBrokersBrokerIdConfigsalterPost(_ctx context.Context, _clusterId string, brokerId int32, localVarOptionals *krsdk.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsGet(_ctx context.Context, _clusterId string, topicName string) (krsdk.TopicConfigDataList, *nethttp.Response, error) {
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

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNameDelete(_ctx context.Context, _clusterId string, topicName string, name string) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNameGet(_ctx context.Context, _clusterId string, topicName string, name string) (krsdk.TopicConfigData, *nethttp.Response, error) {
	return krsdk.TopicConfigData{}, nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsNamePut(_ctx context.Context, _clusterId string, topicName string, name string, localVarOptionals *krsdk.ClustersClusterIdTopicsTopicNameConfigsNamePutOpts) (*nethttp.Response, error) {
	return nil, nil
}

func (m *Configs) ClustersClusterIdTopicsTopicNameConfigsalterPost(_ctx context.Context, _clusterId string, topicName string, localVarOptionals *krsdk.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts) (*nethttp.Response, error) {
	httpResp := &nethttp.Response{
		StatusCode: 204,
	}
	return httpResp, nil
}
