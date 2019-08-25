package schema_registry

import (
	"context"
	"fmt"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	metricsv1 "github.com/confluentinc/ccloudapis/metrics/v1"
	srv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"
	cmd2 "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/confluentinc/ccloud-sdk-go/mock"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	cliMock "github.com/confluentinc/cli/mock"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	net_http "net/http"
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
	srClientMock *srsdk.APIClient
	metrics      *ccsdkmock.Metrics
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
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetTopLevelConfigFunc: func(ctx context.Context) (srsdk.Config, *net_http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			GetTopLevelModeFunc: func(ctx context.Context) (srsdk.ModeGetResponse, *net_http.Response, error) {
				return srsdk.ModeGetResponse{}, nil, nil
			},
		},
	}
	suite.metrics = &ccsdkmock.Metrics{
		SchemaRegistryMetricsFunc: func(arg0 context.Context, arg1 string) (*metricsv1.SchemaRegistryMetric, error) {
			return &metricsv1.SchemaRegistryMetric{
				NumSchemas: 8,
			}, nil
		},
	}
}

func (suite *SRTestSuite) SetupTest() {
	suite.srMock = &mock.SchemaRegistry{
		CreateSchemaRegistryClusterFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryClusterConfig) (*srv1.SchemaRegistryCluster, error) {
			return suite.srCluster, nil
		},
		GetSchemaRegistryClustersFunc: func(ctx context.Context, clusterConfig *srv1.SchemaRegistryCluster) ([]*srv1.SchemaRegistryCluster, error) {
			return []*srv1.SchemaRegistryCluster{suite.srCluster}, nil
		},
	}
}

func (suite *SRTestSuite) newCMD() *cobra.Command {
	cmd := New(&cliMock.Commander{}, suite.conf, suite.srMock, &cmd2.ConfigHelper{}, suite.srClientMock, suite.metrics)
	return cmd
}

func (suite *SRTestSuite) TestCreateSR() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"enable", "--cloud", "aws"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.CreateSchemaRegistryClusterCalled())
}

func (suite *SRTestSuite) TestDescribeSR() {
	cmd := suite.newCMD()
	cmd.SetArgs(append([]string{"describe"}))

	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.srMock.GetSchemaRegistryClustersCalled())
}

func TestSrTestSuite(t *testing.T) {
	suite.Run(t, new(SRTestSuite))
}
