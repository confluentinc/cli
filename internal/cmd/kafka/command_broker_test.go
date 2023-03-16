package kafka

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (suite *KafkaClusterTestSuite) TestBroker_checkAllOrBrokerIdSpecified() {
	req := suite.Require()
	// only --all
	cmd := newCmdWithAllFlag()
	_ = cmd.ParseFlags([]string{"--all"})
	id, all, err := checkAllOrBrokerIdSpecified(cmd, []string{})
	req.NoError(err)
	req.True(all)
	req.Equal(int32(-1), id)
	// only broker id arg
	cmd = newCmdWithAllFlag()
	id, all, err = checkAllOrBrokerIdSpecified(cmd, []string{"1"})
	req.NoError(err)
	req.False(all)
	req.Equal(int32(1), id)
	// --all and broker id arg
	cmd = newCmdWithAllFlag()
	_ = cmd.ParseFlags([]string{"--all"})
	_, _, err = checkAllOrBrokerIdSpecified(cmd, []string{"1"})
	req.Error(err)
	req.Equal(errors.OnlySpecifyAllOrBrokerIDErrorMsg, err.Error())
	// neither
	cmd = newCmdWithAllFlag()
	_, _, err = checkAllOrBrokerIdSpecified(cmd, []string{})
	req.Error(err)
	req.Equal(errors.MustSpecifyAllOrBrokerIDErrorMsg, err.Error())
}

var expectedConfigData = []configOut{
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
	data := parseClusterConfigData(clusterConfig)
	verifyConfigData(req, data, expectedConfigData)
}

func (suite *KafkaClusterTestSuite) TestBroker_parseBrokerConfigData() {
	req := suite.Require()
	value := "testValue"
	brokerConfig := kafkarestv3.BrokerConfigDataList{
		Data: []kafkarestv3.BrokerConfigData{
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
	data := parseBrokerConfigData(brokerConfig)
	verifyConfigData(req, data, expectedConfigData)
}

func newCmdWithAllFlag() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("all", false, "All brokers.")
	return cmd
}

func verifyConfigData(req *require.Assertions, data []*configOut, expected []configOut) {
	req.Equal(len(data), len(expected))
	for i, d := range data {
		req.Equal(expected[i].Name, d.Name)
		req.Equal(expected[i].Value, d.Value)
		req.Equal(expected[i].IsDefault, d.IsDefault)
		req.Equal(expected[i].IsSensitive, d.IsSensitive)
		req.Equal(expected[i].IsReadOnly, d.IsReadOnly)
	}
}
