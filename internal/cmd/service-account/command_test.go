package service_account

import (
	"context"
	"strconv"
	"testing"

	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	test_utils "github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	serviceAccountId   = int32(123)
	serviceDescription = "testing"
	serviceName        = "demo"
)

type ServiceAccountTestSuite struct {
	suite.Suite
	conf            *v3.Config
	userMock        *ccsdkmock.User
	analyticsOutput []segment.Message
	analyticsClient analytics.Client
}

func (suite *ServiceAccountTestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
	suite.userMock = &ccsdkmock.User{
		CreateServiceAccountFunc: func(arg0 context.Context, arg1 *orgv1.User) (user *orgv1.User, e error) {
			return &orgv1.User{
				Id:                 serviceAccountId,
				ServiceName:        serviceName,
				ServiceDescription: serviceDescription,
				ServiceAccount:     true,
			}, nil
		},
		DeleteServiceAccountFunc: func(arg0 context.Context, arg1 *orgv1.User) error {
			return nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = test_utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *ServiceAccountTestSuite) newCmd(conf *v3.Config) *command {
	client := &ccloud.Client{
		User: suite.userMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	cmd := New(prerunner, suite.analyticsClient)
	return cmd
}

func (suite *ServiceAccountTestSuite) TestCreateServiceAccountService() {
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	args := append([]string{"create", serviceName, "--description", serviceDescription})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.CreateServiceAccountCalled())
	test_utils.CheckTrackedResourceIDInt32(suite.analyticsOutput[0], serviceAccountId, req)
}

func (suite *ServiceAccountTestSuite) TestDeleteServiceAccountService() {
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	args := append([]string{"delete", strconv.Itoa(int(serviceAccountId))})
	err := test_utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.DeleteServiceAccountCalled())
	test_utils.CheckTrackedResourceIDInt32(suite.analyticsOutput[0], serviceAccountId, req)
}

func TestServiceAccountTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountTestSuite))
}
