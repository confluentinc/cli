package schema_registry

import (
	"context"
	"net/http"
	"testing"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	versionString = "12345"
	versionInt32  = int32(12345)
	id            = int32(123)
)

type SchemaTestSuite struct {
	suite.Suite
	conf             *v3.Config
	kafkaCluster     *schedv1.KafkaCluster
	srCluster        *schedv1.SchemaRegistryCluster
	srClientMock     *srsdk.APIClient
	srMothershipMock *mock.SchemaRegistry
}

func (suite *SchemaTestSuite) SetupSuite() {
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

func (suite *SchemaTestSuite) SetupTest() {
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetSchemaFunc: func(ctx context.Context, id int32, opts *srsdk.GetSchemaOpts) (srsdk.SchemaString, *http.Response, error) {
				return srsdk.SchemaString{Schema: "Potatoes"}, nil, nil
			},
			GetSchemaByVersionFunc: func(ctx context.Context, subject, version string, opts *srsdk.GetSchemaByVersionOpts) (schema srsdk.Schema, response *http.Response, e error) {
				return srsdk.Schema{Schema: "Potatoes", Version: versionInt32}, nil, nil
			},
			DeleteSchemaVersionFunc: func(ctx context.Context, subject, version string) (i int32, response *http.Response, e error) {
				return id, nil, nil
			},
			DeleteSubjectFunc: func(ctx context.Context, subject string) (int32s []int32, response *http.Response, e error) {
				return []int32{id}, nil, nil
			},
		},
	}
}

func (suite *SchemaTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMothershipMock,
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), suite.conf, suite.srClientMock, suite.conf.Logger)
	return cmd
}

func (suite *SchemaTestSuite) TestDescribeById() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"schema", "describe", "123"}))
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
	cmd.SetArgs(append([]string{"schema", "delete", "--subject", subjectName, "--version", "all"}))
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
	cmd.SetArgs(append([]string{"schema", "delete", "--subject", subjectName, "--version", versionString}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.DeleteSchemaVersionCalled())
	retVal := apiMock.DeleteSchemaVersionCalls()[0]
	req.Equal(retVal.Subject, subjectName)
	req.Equal(retVal.Version, "12345")
}

func (suite *SchemaTestSuite) TestDescribeBySubjectVersion() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"schema", "describe", "--subject", subjectName, "--version", versionString}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetSchemaByVersionCalled())
	retVal := apiMock.GetSchemaByVersionCalls()[0]
	req.Equal(retVal.Subject, subjectName)
	req.Equal(retVal.Version, versionString)
}

func (suite *SchemaTestSuite) TestDescribeByBothSubjectVersionAndId() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"schema", "describe", "--subject", subjectName, "--version", versionString, "123"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
}

func (suite *SchemaTestSuite) TestDescribeBySubjectVersionMissingVersion() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"schema", "describe", "--subject", subjectName}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
}

func (suite *SchemaTestSuite) TestDescribeBySubjectVersionMissingSubject() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"schema", "describe", "--version", versionString}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NotNil(err)
}

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}
