package auditlog

import (
	"context"
	net_http "net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	cliMock "github.com/confluentinc/cli/mock"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/confluentinc/mds-sdk-go/mock"
)

type AuditConfigTestSuite struct {
	suite.Suite
	conf      *config.Config
	kafkaApi  mds.AuditLogConfigurationApi
	preRunner pcmd.PreRunner
}

func (suite *AuditConfigTestSuite) SetupSuite() {
	suite.conf = config.New()
	suite.conf.CLIName = "confluent"
	suite.conf.Logger = log.New()
	suite.conf.AuthURL = "http://test"
	suite.conf.AuthToken = "T0k3n"
}

func (suite *AuditConfigTestSuite) TearDownSuite() {
}

func (suite *AuditConfigTestSuite) SetupTest() {
	suite.preRunner = &cliMock.Commander{}
}

func (suite *AuditConfigTestSuite) newMockConfigCmd(expect chan interface{}, message string) *cobra.Command {
	suite.kafkaApi = &mock.AuditLogConfigurationApi{
		GetConfigFunc: func(ctx context.Context) (mds.AuditLogConfigSpec, *net_http.Response, error) {
			return mds.AuditLogConfigSpec{
				Destinations: mds.AuditLogConfigDestinations{
					BootstrapServers: []string{"localhost:8090"},
					Topics: map[string]mds.AuditLogConfigDestinationConfig{
						"_confluent-audit-log_default": {
							RetentionMs: 30 * 24 * 60 * 60 * 1000,
						},
					},
				},
				ExcludedPrincipals: []string{},
				DefaultTopics: mds.AuditLogConfigDefaultTopics{
					Allowed: "_confluent-audit-log_default",
					Denied:  "_confluent-audit-log_default",
				},
				Routes:   map[string]mds.AuditLogConfigRouteCategories{},
				Metadata: mds.AuditLogConfigMetadata{},
			}, nil, nil
		},
		PutConfigFunc:            nil, //TODO
		ListRoutesFunc:           nil, //TODO
		ResolveResourceRouteFunc: nil, //TODO
	}
	mdsClient := mds.NewAPIClient(mds.NewConfiguration())
	mdsClient.AuditLogConfigurationApi = suite.kafkaApi
	return New(suite.preRunner, suite.conf, mdsClient)
}

func TestAuditConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AuditConfigTestSuite))
}
