package broker

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

type KafkaClusterTestSuite struct {
	suite.Suite
}

var expectedConfigData = []ConfigOut{
	{
		"testConfig",
		"testValue",
		true,
		true,
		true,
	},
	{
		"testConfig2",
		"",
		true,
		true,
		true,
	},
}

func (suite *KafkaClusterTestSuite) TestBroker_parseClusterConfigData() {
	req := suite.Require()
	value := "testValue"
	clusterConfig := kafkarestv3.ClusterConfigDataList{
		Data: []kafkarestv3.ClusterConfigData{
			{
				Name:        "testConfig",
				Value:       &value,
				IsReadOnly:  true,
				IsSensitive: true,
				IsDefault:   true,
			},
			{
				Name:        "testConfig2",
				Value:       nil,
				IsReadOnly:  true,
				IsSensitive: true,
				IsDefault:   true,
			},
		},
	}
	data := ParseClusterConfigData(clusterConfig)
	verifyConfigData(req, data, expectedConfigData)
}

func verifyConfigData(req *require.Assertions, data []*ConfigOut, expected []ConfigOut) {
	req.Equal(len(data), len(expected))
	for i, d := range data {
		req.Equal(expected[i].Name, d.Name)
		req.Equal(expected[i].Value, d.Value)
		req.Equal(expected[i].IsDefault, d.IsDefault)
		req.Equal(expected[i].IsSensitive, d.IsSensitive)
		req.Equal(expected[i].IsReadOnly, d.IsReadOnly)
	}
}
