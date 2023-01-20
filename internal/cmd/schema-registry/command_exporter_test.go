package schemaregistry

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
)

var (
	exporterName = "my_exporter"
	contextName  = "my_context"
)

type ExporterTestSuite struct {
	suite.Suite
	conf             *v1.Config
	kafkaCluster     *ccstructs.KafkaCluster
	srCluster        *ccloudv1.SchemaRegistryCluster
	srClientMock     *srsdk.APIClient
	srMothershipMock *ccloudv1mock.SchemaRegistry
}

func (suite *ExporterTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.srMothershipMock = &ccloudv1mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(_ context.Context, clusterConfig *ccloudv1.SchemaRegistryClusterConfig) (*ccloudv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(_ context.Context, clusterConfig *ccloudv1.SchemaRegistryCluster) (*ccloudv1.SchemaRegistryCluster, error) {
			return nil, nil
		},
	}
	ctx := suite.conf.Context()
	srCluster := ctx.SchemaRegistryClusters[ctx.GetEnvironment().GetId()]
	srCluster.SrCredentials = &v1.APIKeyPair{Key: "key", Secret: "secret"}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
	suite.kafkaCluster = &ccstructs.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Enterprise: true,
	}
	suite.srCluster = &ccloudv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

func (suite *ExporterTestSuite) SetupTest() {
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			CreateExporterFunc: func(_ context.Context, _ srsdk.CreateExporterRequest) (srsdk.CreateExporterResponse, *http.Response, error) {
				return srsdk.CreateExporterResponse{Name: exporterName}, nil, nil
			},
			GetExportersFunc: func(_ context.Context) ([]string, *http.Response, error) {
				return []string{exporterName}, nil, nil
			},
			GetExporterInfoFunc: func(_ context.Context, name string) (srsdk.ExporterInfo, *http.Response, error) {
				return srsdk.ExporterInfo{Name: exporterName, Subjects: []string{subjectName}, SubjectRenameFormat: "${subject}", ContextType: "AUTO", Config: map[string]string{}}, nil, nil
			},
			GetExporterStatusFunc: func(_ context.Context, name string) (srsdk.ExporterStatus, *http.Response, error) {
				return srsdk.ExporterStatus{Name: exporterName, State: "PAUSED", Offset: 0, Ts: 0, Trace: ""}, nil, nil
			},
			PutExporterFunc: func(_ context.Context, name string, _ srsdk.UpdateExporterRequest) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			GetExporterConfigFunc: func(_ context.Context, name string) (map[string]string, *http.Response, error) {
				return map[string]string{"key": "value"}, nil, nil
			},
			PauseExporterFunc: func(_ context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			ResumeExporterFunc: func(_ context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			ResetExporterFunc: func(_ context.Context, name string) (srsdk.UpdateExporterResponse, *http.Response, error) {
				return srsdk.UpdateExporterResponse{Name: exporterName}, nil, nil
			},
			DeleteExporterFunc: func(_ context.Context, name string) (*http.Response, error) {
				return nil, nil
			},
		},
	}
}

func (suite *ExporterTestSuite) newCMD() *cobra.Command {
	client := &ccloudv1.Client{
		SchemaRegistry: suite.srMothershipMock,
	}
	return New(suite.conf, climock.NewPreRunnerMock(client, nil, nil, nil, suite.conf), suite.srClientMock)
}

func (suite *ExporterTestSuite) TestCreateExporter() {
	cmd := suite.newCMD()
	req := require.New(suite.T())
	dir, err := createTempDir()
	req.NoError(err)
	configs := "key1=value1\nkey2=value2"
	configPath := filepath.Join(dir, "config.txt")
	req.NoError(os.WriteFile(configPath, []byte(configs), 0644))
	cmd.SetArgs([]string{"exporter", "create", exporterName, "--context-type", "AUTO",
		"--context", contextName, "--subjects", subjectName, "--config-file", configPath})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err = cmd.Execute()
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.CreateExporterCalled())
	req.NoError(os.RemoveAll(dir))
	params := apiMock.CreateExporterCalls()[0]
	req.Equal(params.Body.Name, exporterName)

	req.Equal("Created schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestListExporters() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "list"})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExportersCalled())

	req.Equal("   Exporter    \n---------------\n  my_exporter  \n", output.String())
}

func (suite *ExporterTestSuite) TestDescribeExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "describe", exporterName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterInfoCalled())
	params := apiMock.GetExporterInfoCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("+----------------+-------------+\n"+
		"| Name           | my_exporter |\n"+
		"| Subjects       | Subject     |\n"+
		"| Subject Format | ${subject}  |\n"+
		"| Context Type   | AUTO        |\n"+
		"| Context        |             |\n"+
		"| Config         |             |\n"+
		"+----------------+-------------+\n", output.String())
}

func (suite *ExporterTestSuite) TestStatusExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "get-status", exporterName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterStatusCalled())
	params := apiMock.GetExporterStatusCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("+-------------+-------------+\n"+
		"| Name        | my_exporter |\n"+
		"| State       | PAUSED      |\n"+
		"| Offset      |           0 |\n"+
		"| Timestamp   |           0 |\n"+
		"| Error Trace |             |\n"+
		"+-------------+-------------+\n", output.String())
}

func (suite *ExporterTestSuite) TestUpdateExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "update", exporterName, "--context", contextName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.PutExporterCalled())
	params := apiMock.PutExporterCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("Updated schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestGetExporterConfig() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "get-config", exporterName, "--output", "yaml"})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetExporterConfigCalled())
	params := apiMock.GetExporterConfigCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("key: value\n", output.String())
}

func (suite *ExporterTestSuite) TestPauseExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "pause", exporterName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.PauseExporterCalled())
	params := apiMock.PauseExporterCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("Paused schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestResumeExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "resume", exporterName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ResumeExporterCalled())
	params := apiMock.ResumeExporterCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("Resumed schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestResetExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "reset", exporterName})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ResetExporterCalled())
	params := apiMock.ResetExporterCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("Reset schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestDeleteExporter() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"exporter", "delete", exporterName, "--force"})
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteExporterCalled())
	params := apiMock.DeleteExporterCalls()[0]
	req.Equal(params.Name, exporterName)

	req.Equal("Deleted schema exporter \"my_exporter\".\n", output.String())
}

func (suite *ExporterTestSuite) TestConvertMapToString() {
	m := map[string]string{"name": "alice", "phone": "xxx-xxx-xxxx", "age": "20"}
	req := require.New(suite.T())
	req.Equal("age=\"20\"\nname=\"alice\"\nphone=\"xxx-xxx-xxxx\"", convertMapToString(m))
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
