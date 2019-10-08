package schema_registry

import (
	"context"
	"net/http"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/version"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	versionString = "12345"
	versionInt32  = int32(12345)
	id            = int32(123)
)

type SchemaTestSuite struct {
	suite.Suite
	conf             *config.Config
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *mock.SchemaRegistry
	srClientMock     *srsdk.APIClient
}

func (suite *SchemaTestSuite) SetupSuite() {
	suite.conf = config.AuthenticatedConfigMock()
	srCluster, _ := suite.conf.SchemaRegistryCluster()
	srCluster.SrCredentials = &config.APIKeyPair{Key: "key", Secret: "secret"}

	cluster := suite.conf.Context().ActiveKafkaCluster()
	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         cluster.ID,
		Name:       cluster.Name,
		Endpoint:   cluster.APIEndpoint,
		Enterprise: true,
	}
	suite.srCluster = &srv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

func (suite *SchemaTestSuite) SetupTest() {
	suite.srMothershipMock = &mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryClusterConfig) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) (*srv1.SchemaRegistryCluster, error) {
			return nil, nil
		},
	}

	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetSchemaFunc: func(ctx context.Context, id int32) (srsdk.SchemaString, *http.Response, error) {
				return srsdk.SchemaString{Schema: "Potatoes"}, nil, nil
			},
			GetSchemaByVersionFunc: func(ctx context.Context, subject, version string) (schema srsdk.Schema, response *http.Response, e error) {
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
	cmd := New(&cliMock.Commander{}, suite.conf, suite.srMothershipMock, &pcmd.ConfigHelper{
		Config:  config.AuthenticatedConfigMock(),
		Version: &version.Version{},
	}, suite.srClientMock, nil, nil)
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

func (suite *SchemaTestSuite) TestDescribeBySubject() {
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

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}
