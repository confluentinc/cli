package connect

import (
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	connectorID   = "lcc-123"
	connectorName = "myTestConnector"
	pluginType    = "DummyPlugin"
)

type ConnectTestSuite struct {
	suite.Suite
	conf               *v3.Config
	kafkaCluster       *schedv1.KafkaCluster
	connector          *schedv1.Connector
	connectorInfo      *opv1.ConnectorInfo
	connectMock        *ccsdkmock.Connect
	kafkaMock          *ccsdkmock.Kafka
	connectorExpansion *opv1.ConnectorExpansion
	analyticsClient    analytics.Client
	analyticsOutput    []segment.Message
}

func (suite *ConnectTestSuite) SetupSuite() {
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

	suite.connectorInfo = &opv1.ConnectorInfo{
		Name: connectorName,
		Type: "source",
	}

	suite.connectorExpansion = &opv1.ConnectorExpansion{
		Id: &opv1.ConnectorId{Id: connectorID},
		Info: &opv1.ConnectorInfo{
			Name:   connectorName,
			Type:   "Sink",
			Config: map[string]string{},
		},
		Status: &opv1.ConnectorStateInfo{Name: connectorName, Connector: &opv1.ConnectorState{State: "Running"},
			Tasks: []*opv1.TaskState{{Id: 1, State: "Running"}},
		}}

}

func (suite *ConnectTestSuite) SetupTest() {
	suite.kafkaMock = &ccsdkmock.Kafka{
		DescribeFunc: func(_ context.Context, _ *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error) {
			return suite.kafkaCluster, nil
		},
		ListFunc: func(_ context.Context, _ *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error) {
			return []*schedv1.KafkaCluster{suite.kafkaCluster}, nil
		},
	}
	suite.connectMock = &ccsdkmock.Connect{
		CreateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		UpdateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		PauseFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		ResumeFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		DeleteFunc: func(_ context.Context, _ *schedv1.Connector) error {
			return nil
		},
		ListWithExpansionsFunc: func(_ context.Context, _ *schedv1.Connector, _ string) (map[string]*opv1.ConnectorExpansion, error) {
			return map[string]*opv1.ConnectorExpansion{connectorID: suite.connectorExpansion}, nil
		},
		GetExpansionByIdFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorExpansion, error) {
			return suite.connectorExpansion, nil
		},
		GetExpansionByNameFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorExpansion, error) {
			return suite.connectorExpansion, nil
		},
		GetFunc: func(_ context.Context, _ *schedv1.Connector) (*opv1.ConnectorInfo, error) {
			return suite.connectorInfo, nil
		},
		ValidateFunc: func(_ context.Context, _ *schedv1.ConnectorConfig) (*opv1.ConfigInfos, error) {
			return &opv1.ConfigInfos{Configs: []*opv1.Configs{{Value: &opv1.ConfigValue{Value: "abc", Errors: []string{"new error"}}}}}, errors.New("config.name")
		},
		GetPluginsFunc: func(_ context.Context, _ *schedv1.Connector, _ string) ([]*opv1.ConnectorPluginInfo, error) {
			return []*opv1.ConnectorPluginInfo{
				{
					Class: "test-plugin",
					Type:  "source",
				},
			}, nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *ConnectTestSuite) newCmd() *command {
	prerunner := cliMock.NewPreRunnerMock(&ccloud.Client{Connect: suite.connectMock, Kafka: suite.kafkaMock}, nil, nil, suite.conf)
	cmd := New("ccloud", prerunner, suite.analyticsClient)
	return cmd
}

func (suite *ConnectTestSuite) TestPauseConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"pause", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.PauseCalled())
	retVal := suite.connectMock.PauseCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"resume", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ResumeCalled())
	retVal := suite.connectMock.ResumeCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCmd()
	args := []string{"delete", connectorID}
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	retVal := suite.connectMock.DeleteCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], connectorID, req)
}

func (suite *ConnectTestSuite) TestListConnectors() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.ListWithExpansionsCalled())
	retVal := suite.connectMock.ListWithExpansionsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"describe", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.GetExpansionByIdCalled())
	retVal := suite.connectMock.GetExpansionByIdCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestCreateConnector() {
	cmd := suite.newCmd()
	args := []string{"create", "--config", "../../../test/fixtures/input/connector-config.yaml"}
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
	// TODO add back with analytics
	// test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], connectorID, req)
}

func (suite *ConnectTestSuite) TestCreateConnectorNewFormat() {
	cmd := suite.newCmd()
	args := []string{"create", "--config", "../../../test/fixtures/input/connector-config-new-format.json"}
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.CreateCalled())
	retVal := suite.connectMock.CreateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], connectorID, req)
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedNewFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connector-config-malformed-new.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	fmt.Printf("error-- %s", err.Error())
	assert.Contains(suite.T(), err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedOldFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", "--config", "../../../test/fixtures/input/connector-config-malformed-old.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	fmt.Printf("error-- %s", err.Error())
	assert.Contains(suite.T(), err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestUpdateConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"update", connectorID, "--config", "../../../test/fixtures/input/connector-config.yaml"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectMock.UpdateCalled())
	retVal := suite.connectMock.UpdateCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestServerComplete() {
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
					Text:        connectorID,
					Description: connectorName,
				},
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_ = tt.fields.Command.PersistentPreRunE(tt.fields.Command.Command, []string{})
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func (suite *ConnectTestSuite) TestServerClusterFlagComplete() {
	flagName := "cluster"
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
			name: "suggest for flag",
			fields: fields{
				Command: suite.newCmd(),
			},
			want: []prompt.Suggest{
				{
					Text:        suite.kafkaCluster.Id,
					Description: suite.kafkaCluster.Name,
				},
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_ = tt.fields.Command.PersistentPreRunE(tt.fields.Command.Command, []string{})
			got := tt.fields.Command.ServerFlagComplete()[flagName]()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func (suite *ConnectTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd()
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"connector delete", "connector describe", "connector pause", "connector resume", "connector update"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func (suite *ConnectTestSuite) TestPluginList() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.connectMock.GetPluginsCalled())
	retVal := suite.connectMock.GetPluginsCalls()[0]
	req.Equal(retVal.Arg1.KafkaClusterId, suite.kafkaCluster.Id)
}

func (suite *ConnectTestSuite) TestPluginDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "describe", pluginType})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.connectMock.ValidateCalled())
	retVal := suite.connectMock.ValidateCalls()[0]
	req.Equal(retVal.Arg1.Plugin, pluginType)
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
