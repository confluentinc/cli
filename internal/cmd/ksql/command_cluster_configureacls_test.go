package ksql

import (
	"testing"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
)

func TestGetCreateAclRequestDataList(t *testing.T) {
	actual := getCreateAclRequestDataList([]*ccstructs.ACLBinding{{}})
	expected := kafkarestv3.CreateAclRequestDataList{Data: []kafkarestv3.CreateAclRequestData{{}}}
	require.Equal(t, expected, actual)
}
