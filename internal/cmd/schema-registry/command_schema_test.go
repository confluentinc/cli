package schemaregistry

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/output"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	versionString = "12345"
	versionInt32  = int32(12345)
	id            = int32(100004)
)

type SchemaTestSuite struct {
	suite.Suite
	conf             *v1.Config
	dynamicContext   *dynamicconfig.DynamicConfig
	kafkaCluster     *schedv1.KafkaCluster
	srCluster        *schedv1.SchemaRegistryCluster
	srClientMock     *srsdk.APIClient
	srMothershipMock *mock.SchemaRegistry
}

func (suite *SchemaTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
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
	srCluster.SrCredentials = &v1.APIKeyPair{Key: "key", Secret: "secret"}
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

func (suite *SchemaTestSuite) SetupTest() {
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			RegisterFunc: func(_ context.Context, _ string, _ srsdk.RegisterSchemaRequest) (srsdk.RegisterSchemaResponse, *http.Response, error) {
				return srsdk.RegisterSchemaResponse{Id: id}, nil, nil
			},
			GetSchemaFunc: func(_ context.Context, _ int32, _ *srsdk.GetSchemaOpts) (srsdk.SchemaString, *http.Response, error) {
				return srsdk.SchemaString{Schema: "Potatoes"}, nil, nil
			},
			GetSchemaByVersionFunc: func(_ context.Context, _, _ string, _ *srsdk.GetSchemaByVersionOpts) (srsdk.Schema, *http.Response, error) {
				return srsdk.Schema{Schema: "Potatoes", Version: versionInt32}, nil, nil
			},
			DeleteSchemaVersionFunc: func(_ context.Context, _, _ string, _ *srsdk.DeleteSchemaVersionOpts) (int32, *http.Response, error) {
				return id, nil, nil
			},
			DeleteSubjectFunc: func(_ context.Context, _ string, _ *srsdk.DeleteSubjectOpts) ([]int32, *http.Response, error) {
				return []int32{id}, nil, nil
			},
		},
	}
	suite.dynamicContext = cliMock.AuthenticatedDynamicConfigMock()
}

func (suite *SchemaTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMothershipMock,
	}
	cmd := New(suite.conf, cliMock.NewPreRunnerMock(client, nil, nil, nil, suite.conf), suite.srClientMock)
	return cmd
}

func (suite *SchemaTestSuite) TestGetSchemaMetaInfo() {
	req := require.New(suite.T())
	metaInfo := GetMetaInfoFromSchemaId(id)
	req.Equal([]byte{0x0, 0x0, 0x1, 0x86, 0xa4}, metaInfo)
}

func (suite *SchemaTestSuite) TestRegisterSchema() {
	cmd := suite.newCMD()
	cmd.Flags().String(output.FlagName, "human", `Specify the output format as "human", "json", or "yaml".`)
	req := require.New(suite.T())
	storePath := suite.T().TempDir()
	file, err := ioutil.TempFile(storePath, "schema-file")
	req.Nil(err)
	err = file.Close()
	req.Nil(err)
	fileName := file.Name()
	defer os.Remove(fileName)
	schemaCfg := &RegisterSchemaConfigs{
		SchemaPath: &fileName,
		Subject:    subjectName,
	}
	metaInfo, err := RegisterSchemaWithAuth(cmd, schemaCfg, suite.srClientMock, cmd.Context())
	req.Nil(err)
	expectedMetaInfo := GetMetaInfoFromSchemaId(id)
	req.Equal(expectedMetaInfo, metaInfo)
}

func (suite *SchemaTestSuite) TestRequestSchemaById() {
	tmpdir := suite.T().TempDir()
	tempStorePath, _, err := RequestSchemaWithId(123, tmpdir, "subject", suite.srClientMock, suite.newCMD().Context())
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetSchemaCalled())
	content, err := os.ReadFile(tempStorePath)
	req.Nil(err)
	req.Equal(string(content), "Potatoes")
	err = os.Remove(tempStorePath)
	req.Nil(err)
}

func (suite *SchemaTestSuite) TestDescribeById() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"schema", "describe", "100004"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetSchemaCalled())
	retVal := apiMock.GetSchemaCalls()[0]
	req.Equal(retVal.Id, id)
}

func (suite *SchemaTestSuite) TestDeleteAllSchemas() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"schema", "delete", "--subject", subjectName, "--version", "all"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteSubjectCalled())
	retVal := apiMock.DeleteSubjectCalls()[0]
	req.Equal(retVal.Subject, subjectName)
}

func (suite *SchemaTestSuite) TestDeleteSchemaVersion() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"schema", "delete", "--subject", subjectName, "--version", versionString})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteSchemaVersionCalled())
	retVal := apiMock.DeleteSchemaVersionCalls()[0]
	req.Equal(retVal.Subject, subjectName)
	req.Equal(retVal.Version, "12345")
}

func (suite *SchemaTestSuite) TestPermanentDeleteSchemaVersion() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"schema", "delete", "--subject", subjectName, "--version", versionString, "--permanent"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteSchemaVersionCalled())
	retVal := apiMock.DeleteSchemaVersionCalls()[0]
	req.Equal(retVal.Subject, subjectName)
	req.Equal(retVal.Version, "12345")
	req.Equal(retVal.LocalVarOptionals.Permanent.Value(), true)
}

func (suite *SchemaTestSuite) TestDescribeBySubjectVersion() {
	cmd := suite.newCMD()
	cmd.SetArgs([]string{"schema", "describe", "--subject", subjectName, "--version", versionString})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetSchemaByVersionCalled())
	retVal := apiMock.GetSchemaByVersionCalls()[0]
	req.Equal(retVal.Subject, subjectName)
	req.Equal(retVal.Version, versionString)
}

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}
