package ksql

import (
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/stretchr/testify/require"
)

func TestGetCreateAclRequestDataList(t *testing.T) {
	actual := getCreateAclRequestDataList([]*schedv1.ACLBinding{{}})
	expected := kafkarestv3.CreateAclRequestDataList{Data: []kafkarestv3.CreateAclRequestData{{}}}
	require.Equal(t, expected, actual)
}
