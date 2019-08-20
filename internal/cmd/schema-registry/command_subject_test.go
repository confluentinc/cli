package schema_registry

import (
	"context"
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
	cliMock "github.com/confluentinc/cli/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	net_http "net/http"
	"testing"
)

type SubjectTestSuite struct {
	suite.Suite
	conf             *config.Config
	kafkaCluster     *kafkav1.KafkaCluster
	srCluster        *srv1.SchemaRegistryCluster
	srMothershipMock *mock.SchemaRegistry
	srClientMock     *srsdk.APIClient
}

func (suite *SubjectTestSuite) SetupSuite() {
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
			ListFunc: func(ctx context.Context) ([]string, *net_http.Response, error) {
				return []string{"subject 1", "subject 2"}, nil, nil
			},
		},
	}
}

func (suite *SubjectTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.srMothershipMock, &pcmd.ConfigHelper{Config: &config.Config{
		Auth: &config.AuthConfig{Account: &orgv1.Account{Id: "777"}},
	},
		Version: &version.Version{},
	}, suite.srClientMock, nil)
	return cmd
}

func (suite *SubjectTestSuite) TestSubjectUpdate() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"subject", "list"}))
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	apiMock, _ := suite.srClientMock.DefaultApi.(*srMock.DefaultApi)
	req.True(apiMock.ListCalled())
}

func TestSubjectSuite(t *testing.T) {
	suite.Run(t, new(SubjectTestSuite))
}
