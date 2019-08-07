package schema_registry

import (
	"context"
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	cmd2 "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
	cliMock "github.com/confluentinc/cli/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	net_http "net/http"
	"testing"
)

type CompatibilityTestSuite struct {
	suite.Suite
	conf             *config.Config
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *mock.SchemaRegistry
	srClientMock     *srsdk.APIClient
}

func (suite *CompatibilityTestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.Auth = &config.AuthConfig{
		User:    new(orgv1.User),
		Account: &orgv1.Account{Id: "testAccount"},
	}
	user := suite.conf.Auth
	name := fmt.Sprintf("login-%s-%s", user.User.Email, suite.conf.AuthURL)

	suite.conf.Platforms[name] = &config.Platform{
		Server: suite.conf.AuthURL,
	}

	suite.conf.Credentials[name] = &config.Credential{
		Username: user.User.Email,
	}

	suite.conf.Contexts[name] = &config.Context{
		Platform:      name,
		Credential:    name,
		Kafka:         kafkaClusterID,
		KafkaClusters: map[string]*config.KafkaClusterConfig{kafkaClusterID: {}},
	}

	suite.conf.CurrentContext = name

	srCluster, _ := suite.conf.SchemaRegistryCluster()
	srCluster.SrCredentials = &config.APIKeyPair{Key: "key", Secret: "secret"}

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
	}

	suite.srCluster = &srv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

func (suite *CompatibilityTestSuite) SetupTest() {
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
			GetSchemaFunc: func(ctx context.Context, id int32) (srsdk.SchemaString, *net_http.Response, error) {
				return srsdk.SchemaString{Schema: "Potatoes"}, nil, nil
			},
			GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *net_http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			GetSubjectLevelConfigFunc: func(ctx context.Context, subject string) (srsdk.Config, *net_http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			UpdateTopLevelConfigFunc: func(ctx context.Context, body srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, *net_http.Response, error) {
				return srsdk.ConfigUpdateRequest{}, nil, nil
			},
			UpdateSubjectLevelConfigFunc: func(ctx context.Context, subject string, body srsdk.ConfigUpdateRequest) (srsdk.ConfigUpdateRequest, *net_http.Response, error) {
				return srsdk.ConfigUpdateRequest{}, nil, nil
			},
			TestCompatabilityBySubjectNameFunc: func(ctx context.Context, subject, version string, body srsdk.RegisterSchemaRequest, localVarOptionals *srsdk.TestCompatabilityBySubjectNameOpts) (srsdk.CompatibilityCheckResponse, *net_http.Response, error) {
				return srsdk.CompatibilityCheckResponse{}, nil, nil
			},
		},
	}
}

func (suite *CompatibilityTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.srMothershipMock, &cmd2.ConfigHelper{Config: &config.Config{
		Auth: &config.AuthConfig{Account: &orgv1.Account{Id: "777"}},
	},
		Version: &version.Version{},
	}, suite.srClientMock)
	return cmd
}

func (suite *CompatibilityTestSuite) TestDescribeGlobalCompat() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"compatibility", "describe"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.GetTopLevelConfigCalled())
	req.False(apiMock.GetSubjectLevelConfigCalled())
}

func (suite *CompatibilityTestSuite) TestDescribeSubjectCompat() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"compatibility", "describe", "--subject", "payments"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.False(apiMock.GetTopLevelConfigCalled())
	req.True(apiMock.GetSubjectLevelConfigCalled())
}

func (suite *CompatibilityTestSuite) TestUpdateGlobalCompat() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"compatibility", "update", "FULL"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.False(apiMock.UpdateSubjectLevelConfigCalled())
	req.True(apiMock.UpdateTopLevelConfigCalled())
}

func (suite *CompatibilityTestSuite) TestUpdateSubjectCompat() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"compatibility", "update", "FULL", "--subject", "payments"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.UpdateSubjectLevelConfigCalled())
	req.False(apiMock.UpdateTopLevelConfigCalled())
}

func (suite *CompatibilityTestSuite) TestCheckCompat() {
	cmd := suite.newCMD()

	f, _ := ioutil.TempFile("", "CheckCompat")
	cmd.SetArgs(append([]string{"compatibility", "check",
		"--subject", "payments",
		"--version", "latest",
		"--schema", f.Name(),
	}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.TestCompatabilityBySubjectNameCalled())
}

func TestCompatibilitySuite(t *testing.T) {
	suite.Run(t, new(CompatibilityTestSuite))
}
