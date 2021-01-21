package connector_catalog

import (
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	connectorID   = "lcc-123"
	pluginType    = "DummyPlugin"
	connectorName = "myTestConnector"
)

type CatalogTestSuite struct {
	suite.Suite
	conf         *v3.Config
	kafkaCluster *schedv1.KafkaCluster
	connector    *schedv1.Connector
	connectMock  *ccsdkmock.Connect
	kafkaMock    *ccsdkmock.Kafka
}

func (suite *CatalogTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         ctx.KafkaClusterContext.GetActiveKafkaClusterId(),
		Name:       "KafkaMock",
		AccountId:  "testAccount",
		Enterprise: true,
	}
	suite.connector = &schedv1.Connector{
		Name:           connectorName,
		Id:             connectorID,
		KafkaClusterId: suite.kafkaCluster.Id,
		AccountId:      "testAccount",
		Status:         schedv1.Connector_RUNNING,
		UserConfigs:    map[string]string{},
	}
}

func (suite *CatalogTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		ValidateFunc: func(arg0 context.Context, arg1 *schedv1.ConnectorConfig) (connector *opv1.ConfigInfos, e error) {
			return &opv1.ConfigInfos{Configs: []*opv1.Configs{{Value: &opv1.ConfigValue{Value: "abc", Errors: []string{"new error"}}}}}, errors.New("config.name")
		},
		GetPluginsFunc: func(arg0 context.Context, arg1 *schedv1.Connector, arg2 string) (infos []*opv1.ConnectorPluginInfo, e error) {
			return []*opv1.ConnectorPluginInfo{
				{
					Class: "test-plugin",
					Type:  "source",
				},
			}, nil
		},
	}
}

func (suite *CatalogTestSuite) newCmd() *command {
	prerunner := cliMock.NewPreRunnerMock(&ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}, nil, suite.conf)
	cmd := New("ccloud", prerunner)
	return cmd
}

func (suite *CatalogTestSuite) TestCatalogList() {
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetPluginsCalled())
	retVal := suite.connectMock.GetPluginsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *CatalogTestSuite) TestCatalogDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"describe", pluginType}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ValidateCalled())
	retVal := suite.connectMock.ValidateCalls()[0]
	req.Equal(retVal.Arg1.Plugin, pluginType)
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogTestSuite))
}

func (suite *CatalogTestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *command
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(),
			},
			want: []prompt.Suggest{
				{
					Text:        "test-plugin",
					Description: "source",
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				Command: func() *command {
					oldConf := suite.conf
					suite.conf = v3.UnauthenticatedCloudConfigMock()
					c := suite.newCmd()
					suite.conf = oldConf
					return c
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func (suite *CatalogTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd()
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"connector-catalog describe"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}
