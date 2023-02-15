package connect

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	connectmock "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1/mock"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	connectorID   = "lcc-123"
	connectorName = "myTestConnector"
	pluginType    = "DummyPlugin"
)

var (
	plugin = connectv1.InlineResponse2002{
		Class: "DummySink",
		Type:  "sink",
	}

	pluginDescribe = connectv1.InlineResponse2003Configs{
		Value: &connectv1.InlineResponse2003Value{
			Name:   connectv1.PtrString("name"),
			Errors: &[]string{`"name" is required`}},
		Definition: &connectv1.InlineResponse2003Definition{
			Documentation: connectv1.PtrString("Connector Name"),
			Required:      connectv1.PtrBool(true)},
	}

	connector = connectv1.ConnectV1Connector{
		Name:   connectorName,
		Config: map[string]string{},
	}

	connectorExpansion = connectv1.ConnectV1ConnectorExpansion{
		Id: &connectv1.ConnectV1ConnectorExpansionId{Id: connectv1.PtrString(connectorID)},
		Status: &connectv1.ConnectV1ConnectorExpansionStatus{
			Name: connectorName,
			Connector: connectv1.ConnectV1ConnectorExpansionStatusConnector{
				State: "RUNNING",
				Trace: connectv1.PtrString(""),
			},
			Tasks: &[]connectv1.ConnectV1ConnectorExpansionStatusTasks{
				connectv1.ConnectV1ConnectorExpansionStatusTasks{Id: 1, State: "RUNNING"}},
			Type: "Sink",
		},
	}
)

type ConnectTestSuite struct {
	suite.Suite
	conf           *v1.Config
	kafkaCluster   *ccstructs.KafkaCluster
	connectorsMock *connectmock.ConnectorsV1Api
	lifecycleMock  *connectmock.LifecycleV1Api
	pluginMock     *connectmock.PluginsV1Api
}

func (suite *ConnectTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	suite.kafkaCluster = &ccstructs.KafkaCluster{
		Id:         ctx.KafkaClusterContext.GetActiveKafkaClusterId(),
		Name:       "KafkaMock",
		AccountId:  "testAccount",
		Enterprise: true,
	}
}

func (suite *ConnectTestSuite) SetupTest() {
	suite.connectorsMock = &connectmock.ConnectorsV1Api{
		CreateConnectv1ConnectorFunc: func(_ context.Context, _, _ string) connectv1.ApiCreateConnectv1ConnectorRequest {
			return connectv1.ApiCreateConnectv1ConnectorRequest{}
		},
		CreateConnectv1ConnectorExecuteFunc: func(_ connectv1.ApiCreateConnectv1ConnectorRequest) (connectv1.ConnectV1Connector, *http.Response, error) {
			return connector, nil, nil
		},
		CreateOrUpdateConnectv1ConnectorConfigFunc: func(_ context.Context, _, _, _ string) connectv1.ApiCreateOrUpdateConnectv1ConnectorConfigRequest {
			return connectv1.ApiCreateOrUpdateConnectv1ConnectorConfigRequest{}
		},
		CreateOrUpdateConnectv1ConnectorConfigExecuteFunc: func(_ connectv1.ApiCreateOrUpdateConnectv1ConnectorConfigRequest) (connectv1.ConnectV1Connector, *http.Response, error) {
			return connector, nil, nil
		},
		DeleteConnectv1ConnectorFunc: func(_ context.Context, _, _, _ string) connectv1.ApiDeleteConnectv1ConnectorRequest {
			return connectv1.ApiDeleteConnectv1ConnectorRequest{}
		},
		DeleteConnectv1ConnectorExecuteFunc: func(_ connectv1.ApiDeleteConnectv1ConnectorRequest) (connectv1.InlineResponse200, *http.Response, error) {
			return connectv1.InlineResponse200{}, nil, nil
		},
		ListConnectv1ConnectorsWithExpansionsFunc: func(_ context.Context, _, _ string) connectv1.ApiListConnectv1ConnectorsWithExpansionsRequest {
			return connectv1.ApiListConnectv1ConnectorsWithExpansionsRequest{}
		},
		ListConnectv1ConnectorsWithExpansionsExecuteFunc: func(_ connectv1.ApiListConnectv1ConnectorsWithExpansionsRequest) (map[string]connectv1.ConnectV1ConnectorExpansion, *http.Response, error) {
			return map[string]connectv1.ConnectV1ConnectorExpansion{connectorName: connectorExpansion}, nil, nil
		},
	}
	suite.lifecycleMock = &connectmock.LifecycleV1Api{
		PauseConnectv1ConnectorFunc: func(_ context.Context, _, _, _ string) connectv1.ApiPauseConnectv1ConnectorRequest {
			return connectv1.ApiPauseConnectv1ConnectorRequest{}
		},
		PauseConnectv1ConnectorExecuteFunc: func(_ connectv1.ApiPauseConnectv1ConnectorRequest) (*http.Response, error) {
			return nil, nil
		},
		ResumeConnectv1ConnectorFunc: func(_ context.Context, _, _, _ string) connectv1.ApiResumeConnectv1ConnectorRequest {
			return connectv1.ApiResumeConnectv1ConnectorRequest{}
		},
		ResumeConnectv1ConnectorExecuteFunc: func(_ connectv1.ApiResumeConnectv1ConnectorRequest) (*http.Response, error) {
			return nil, nil
		},
	}
	suite.pluginMock = &connectmock.PluginsV1Api{
		ListConnectv1ConnectorPluginsFunc: func(_ context.Context, _, _ string) connectv1.ApiListConnectv1ConnectorPluginsRequest {
			return connectv1.ApiListConnectv1ConnectorPluginsRequest{}
		},
		ListConnectv1ConnectorPluginsExecuteFunc: func(_ connectv1.ApiListConnectv1ConnectorPluginsRequest) ([]connectv1.InlineResponse2002, *http.Response, error) {
			return []connectv1.InlineResponse2002{plugin}, nil, nil
		},
		ValidateConnectv1ConnectorPluginFunc: func(_ context.Context, _, _, _ string) connectv1.ApiValidateConnectv1ConnectorPluginRequest {
			return connectv1.ApiValidateConnectv1ConnectorPluginRequest{}
		},
		ValidateConnectv1ConnectorPluginExecuteFunc: func(_ connectv1.ApiValidateConnectv1ConnectorPluginRequest) (connectv1.InlineResponse2003, *http.Response, error) {
			return connectv1.InlineResponse2003{
				Configs: &[]connectv1.InlineResponse2003Configs{pluginDescribe},
			}, nil, nil
		},
	}
}

func (suite *ConnectTestSuite) newCmd() *cobra.Command {
	connectClient := &connectv1.APIClient{
		ConnectorsV1Api: suite.connectorsMock,
		LifecycleV1Api:  suite.lifecycleMock,
		PluginsV1Api:    suite.pluginMock,
	}
	prerunner := climock.NewPreRunnerMock(nil,
		&ccloudv2.Client{ConnectClient: connectClient}, nil, nil, suite.conf)
	return New(suite.conf, prerunner)
}

func (suite *ConnectTestSuite) TestPauseConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "pause", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.lifecycleMock.PauseConnectv1ConnectorCalled())
	req.True(suite.lifecycleMock.PauseConnectv1ConnectorExecuteCalled())
}

func (suite *ConnectTestSuite) TestResumeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "resume", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.lifecycleMock.ResumeConnectv1ConnectorCalled())
	req.True(suite.lifecycleMock.ResumeConnectv1ConnectorExecuteCalled())
}

func (suite *ConnectTestSuite) TestDeleteConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "delete", connectorID, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.DeleteConnectv1ConnectorCalled())
	req.True(suite.connectorsMock.DeleteConnectv1ConnectorExecuteCalled())
}

func (suite *ConnectTestSuite) TestListConnectors() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "list"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.ListConnectv1ConnectorsWithExpansionsCalled())
	req.True(suite.connectorsMock.ListConnectv1ConnectorsWithExpansionsExecuteCalled())
	req.Contains(buf.String(), connectorID)
}

func (suite *ConnectTestSuite) TestDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "describe", connectorID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.ListConnectv1ConnectorsWithExpansionsCalled())
	req.True(suite.connectorsMock.ListConnectv1ConnectorsWithExpansionsExecuteCalled())
}

func (suite *ConnectTestSuite) TestCreateConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "create", "--config-file", "../../../test/fixtures/input/connect/config.yaml"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.CreateConnectv1ConnectorCalled())
	req.True(suite.connectorsMock.CreateConnectv1ConnectorExecuteCalled())
	req.Contains(buf.String(), connectorID)
}

func (suite *ConnectTestSuite) TestCreateConnectorNewFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "create", "--config-file", "../../../test/fixtures/input/connect/config-new-format.json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.CreateConnectv1ConnectorCalled())
	req.True(suite.connectorsMock.CreateConnectv1ConnectorExecuteCalled())
	req.Contains(buf.String(), connectorID)
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedNewFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "create", "--config-file", "../../../test/fixtures/input/connect/config-malformed-new.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	req.Contains(err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestCreateConnectorMalformedOldFormat() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "create", "--config-file", "../../../test/fixtures/input/connect/config-malformed-old.json"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
	req.Contains(err.Error(), "unable to parse config")
}

func (suite *ConnectTestSuite) TestUpdateConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"cluster", "update", connectorID, "--config-file", "../../../test/fixtures/input/connect/config.yaml"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.connectorsMock.CreateOrUpdateConnectv1ConnectorConfigCalled())
	req.True(suite.connectorsMock.CreateOrUpdateConnectv1ConnectorConfigExecuteCalled())
	req.Contains(buf.String(), fmt.Sprintf(`Updated connector "%s"`, connectorID))
}

func (suite *ConnectTestSuite) TestPluginList() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "list"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.pluginMock.ListConnectv1ConnectorPluginsCalled())
	req.True(suite.pluginMock.ListConnectv1ConnectorPluginsExecuteCalled())
	req.Contains(buf.String(), "DummySink")
}

func (suite *ConnectTestSuite) TestPluginDescribeConnector() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"plugin", "describe", pluginType})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.pluginMock.ValidateConnectv1ConnectorPluginCalled())
	req.True(suite.pluginMock.ValidateConnectv1ConnectorPluginExecuteCalled())
}

func TestConnectTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}
