package schema_registry

import (
	"context"
	"net/http"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	subjectName = "Subject"
)

type SubjectTestSuite struct {
	suite.Suite
	conf             *v3.Config
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *mock.SchemaRegistry
	srClientMock     *srsdk.APIClient
}

func (suite *SubjectTestSuite) SetupSuite() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	ctx := suite.conf.Context()
	srCluster := ctx.SchemaRegistryClusters[ctx.State.Auth.Account.Id]
	srCluster.SrCredentials = &v0.APIKeyPair{Key: "key", Secret: "secret"}
	cluster := ctx.KafkaClusterContext.GetActiveKafkaClusterConfig()
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

func (suite *SubjectTestSuite) SetupTest() {
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
			ListFunc: func(ctx context.Context) ([]string, *http.Response, error) {
				return []string{"subject 1", "subject 2"}, nil, nil
			},
			ListVersionsFunc: func(ctx context.Context, subject string) (int32s []int32, response *http.Response, e error) {
				return []int32{1234, 4567}, nil, nil
			},
			UpdateSubjectLevelConfigFunc: func(ctx context.Context, subject string, body srsdk.ConfigUpdateRequest) (request srsdk.ConfigUpdateRequest, response *http.Response, e error) {
				return srsdk.ConfigUpdateRequest{Compatibility: body.Compatibility}, nil, nil
			},
			UpdateModeFunc: func(ctx context.Context, subject string, body srsdk.ModeUpdateRequest) (request srsdk.ModeUpdateRequest, response *http.Response, e error) {
				return srsdk.ModeUpdateRequest{Mode: body.Mode}, nil, nil
			},
		},
	}
}

func (suite *SubjectTestSuite) newCMD() *cobra.Command {
	client := &ccloud.Client{
		SchemaRegistry: suite.srMothershipMock,
	}
	cmd := New(cliMock.NewPreRunnerMock(client, nil), suite.conf, suite.srClientMock, suite.conf.Logger)
	return cmd
}

func (suite *SubjectTestSuite) TestSubjectList() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ListCalled())
}

func (suite *SubjectTestSuite) TestSubjectUpdateMode() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "update", subjectName, "--mode", "READWRITE"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.False(apiMock.UpdateTopLevelModeCalled())
	req.True(apiMock.UpdateModeCalled())
	retVal := apiMock.UpdateModeCalls()[0]
	req.Equal(retVal.Subject, subjectName)
}

func (suite *SubjectTestSuite) TestSubjectUpdateCompatibility() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "update", subjectName, "--compatibility", "BACKWARD"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.UpdateSubjectLevelConfigCalled())
	retVal := apiMock.UpdateSubjectLevelConfigCalls()[0]
	req.Equal(retVal.Subject, subjectName)
}

func (suite *SubjectTestSuite) TestSubjectUpdateNoArgs() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "update", subjectName}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err, "Error: flag string not set")
}

func (suite *SubjectTestSuite) TestSubjectDescribe() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "describe", subjectName}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ListVersionsCalled())
	retVal := apiMock.ListVersionsCalls()[0]
	req.Equal(retVal.Subject, subjectName)
}

func TestSubjectSuite(t *testing.T) {
	suite.Run(t, new(SubjectTestSuite))
}
