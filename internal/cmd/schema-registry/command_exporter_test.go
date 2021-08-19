package schemaregistry

import (
	"context"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

var (
	exporterName = "my_exporter"
	contextName  = "my_context"
)

type ExporterTestSuite struct {
	suite.Suite
	conf             *v3.Config
	kafkaCluster     *schedv1.KafkaCluster
	srCluster        *schedv1.SchemaRegistryCluster
	srClientMock     *srsdk.APIClient
	srMothershipMock *mock.SchemaRegistry
}

func (suite *ExporterTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	suite.srMothershipMock = &mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryClusterConfig) (*schedv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *schedv1.SchemaRegistryCluster) (*schedv1.SchemaRegistryCluster, error) {
			return nil, nil
		},
	}
	ctx := suite.conf.Context()
	srCluster := ctx.SchemaRegistryClusters[ctx.State.Auth.Account.Id]
	srCluster.SrCredentials = &v0.APIKeyPair{Key: "key", Secret: "secret"}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	suite.kafkaCluster = &schedv1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
	}
	suite.srCluster = &schedv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

func (suite *ExporterTestSuite) SetupTest() {
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			CreateExporterFunc: func(ctx context.Context, body srsdk.CreateExporterRequest) (srsdk.CreateExporterResponse, *http.Response, error) {
				return srsdk.CreateExporterResponse{Name: exporterName}, nil, nil
			},
			GetExportersFunc: func(ctx context.Context) ([]string, *http.Response, error) {
				return []string{exporterName}, nil, nil
			},
			GetExporterInfoFunc: func(ctx context.Context, name string) (srsdk.ExporterInfo, *http.Response, error) {
				return srsdk.ExporterInfo{Name: exporterName, Subjects: []string{subjectName}, ContextType: "AUTO", Config: map[string]string{}}, nil, nil
			},
			GetExporterStatusFunc: func(ctx context.Context, name string) (srsdk.ExporterStatus, *http.Response, error) {
				return srsdk.ExporterStatus{Name: exporterName, State: "PAUSED", Offset: 0, Ts: 0, Trace: ""}, nil, nil
			},
			PutExporterFunc: func(ctx context.Context, name string, body srsdk.UpdateExporterRequest) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			GetExporterConfigFunc: func(ctx context.Context, name string) (map[string]string, *http.Response, error) {
				return map[string]string{}, nil, nil
			},
			PauseExporterFunc: func(ctx context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			ResumeExporterFunc: func(ctx context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			ResetExporterFunc: func(ctx context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			DeleteExporterFunc: func(ctx context.Context, name string) (*http.Response, error) {
				return nil, nil
			},
		},
	}
}

func (suite *ExporterTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMothershipMock,
	}
	cmd := New(suite.conf.CLIName, cliMock.NewPreRunnerMock(client, nil, nil, suite.conf), suite.srClientMock, suite.conf.Logger, cliMock.NewDummyAnalyticsMock())
	return cmd
}

func (suite *ExporterTestSuite) TestCreateExporter() {
	cmd := suite.newCMD()
	req := require.New(suite.T())
	dir, err := createTempDir()
	req.Nil(err)
	configs := "key1=value1\nkey2=value2"
	configPath := filepath.Join(dir, "config.txt")
	req.NoError(ioutil.WriteFile(configPath, []byte(configs), 0644))
	cmd.SetArgs([]string{"exporter", "create", "--name", exporterName, "--context-type", "AUTO",
		"--context", contextName, "--subjects", subjectName, "--config-file", configPath})
	err = cmd.Execute()
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.CreateExporterCalled())
	req.NoError(os.RemoveAll(dir))
	retVal := apiMock.CreateExporterCalls()[0]
	req.Equal(retVal.Body.Name, exporterName)
}

func (suite *ExporterTestSuite) TestListExporters() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExportersCalled())
}

func (suite *ExporterTestSuite) TestDescribeExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "describe", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterInfoCalled())
	retVal := apiMock.GetExporterInfoCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestStatusExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "status", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterStatusCalled())
	retVal := apiMock.GetExporterStatusCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestUpdateExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "update", "--name", exporterName, "--context", contextName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.PutExporterCalled())
	retVal := apiMock.PutExporterCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestGetExporterConfig() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "get-config", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterConfigCalled())
	retVal := apiMock.GetExporterConfigCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestPauseExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "pause", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.PauseExporterCalled())
	retVal := apiMock.PauseExporterCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestResumeExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "resume", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ResumeExporterCalled())
	retVal := apiMock.ResumeExporterCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestResetExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "reset", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ResetExporterCalled())
	retVal := apiMock.ResetExporterCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func (suite *ExporterTestSuite) TestDeleteExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "delete", "--name", exporterName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteExporterCalled())
	retVal := apiMock.DeleteExporterCalls()[0]
	req.Equal(retVal.Name, exporterName)
}

func TestExporterSuite(t *testing.T) {
	suite.Run(t, new(ExporterTestSuite))
}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema-exporter")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return "", err
		}
	}
	return dir, nil
}
