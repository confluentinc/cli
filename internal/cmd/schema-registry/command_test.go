package schema_registry

import (
	"context"
	"fmt"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	kafkaClusterID = "kafka"
	srClusterID    = "sr"
)

type SRTestSuite struct {
	suite.Suite
	conf         *config.Config
	kafkaCluster *kafkav1.KafkaCluster
	srCluster    *srv1.SchemaRegistryCluster
	srMock       *mock.SchemaRegistry
}

func (suite *SRTestSuite) SetupSuite() {
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

	suite.kafkaCluster = &kafkav1.KafkaCluster{
		Id:         kafkaClusterID,
		Enterprise: true,
	}

	suite.srCluster = &srv1.SchemaRegistryCluster{
		Id: srClusterID,
	}
}

func (suite *SRTestSuite) SetupTest() {
	suite.srMock = &mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryClusterConfig) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) (*srv1.SchemaRegistryCluster, error) {
			return nil, nil
		},
	}
}

func (suite *SRTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.srMock)
	return cmd
}

func (suite *SRTestSuite) TestCreateSR() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"enable", "--cloud", "aws", "--cluster", kafkaClusterID}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.CreateSchemaRegistryClusterCalled())
}

func TestSrTestSuite(t *testing.T) {
	suite.Run(t, new(SRTestSuite))
}
