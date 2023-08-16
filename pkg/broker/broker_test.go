package broker

import (
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	metricsmock "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2/mock"
)

type KafkaClusterTestSuite struct {
	suite.Suite
	conf            *config.Config
	envMetadataMock *ccloudv1mock.EnvironmentMetadata
	metricsApi      *metricsmock.Version2Api
	cmkClusterApi   *cmkmock.ClustersCmkV2Api
}

func (suite *KafkaClusterTestSuite) TestBroker_CheckAllOrIdSpecified() {
	req := suite.Require()
	// only --all
	cmd := newCmdWithAllFlag()
	_ = cmd.ParseFlags([]string{"--all"})
	id, all, err := CheckAllOrIdSpecified(cmd, []string{})
	req.NoError(err)
	req.True(all)
	req.Equal(int32(-1), id)
	// only broker id arg
	cmd = newCmdWithAllFlag()
	id, all, err = CheckAllOrIdSpecified(cmd, []string{"1"})
	req.NoError(err)
	req.False(all)
	req.Equal(int32(1), id)
	// --all and broker id arg
	cmd = newCmdWithAllFlag()
	_ = cmd.ParseFlags([]string{"--all"})
	_, _, err = CheckAllOrIdSpecified(cmd, []string{"1"})
	req.Error(err)
	req.Equal(errors.OnlySpecifyAllOrBrokerIDErrorMsg, err.Error())
	// neither
	cmd = newCmdWithAllFlag()
	_, _, err = CheckAllOrIdSpecified(cmd, []string{})
	req.Error(err)
	req.Equal(errors.MustSpecifyAllOrBrokerIDErrorMsg, err.Error())
}

func newCmdWithAllFlag() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("all", false, "All brokers.")
	return cmd
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
