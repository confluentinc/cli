package iam

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	UserId             = int32(123)
	serviceAccountId   = "sa-55555"
	serviceDescription = "testing"
	serviceName        = "demo"
)

type ServiceAccountTestSuite struct {
	suite.Suite
	conf     *v1.Config
	userMock *ccsdkmock.User
}

func (suite *ServiceAccountTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.userMock = &ccsdkmock.User{
		CreateServiceAccountFunc: func(arg0 context.Context, arg1 *orgv1.User) (user *orgv1.User, e error) {
			return &orgv1.User{
				Id:                 UserId,
				ResourceId:         serviceAccountId,
				ServiceName:        serviceName,
				ServiceDescription: serviceDescription,
				ServiceAccount:     true,
			}, nil
		},
		DeleteServiceAccountFunc: func(arg0 context.Context, arg1 *orgv1.User) error {
			return nil
		},
	}
}

func (suite *ServiceAccountTestSuite) newCmd(conf *v1.Config) *serviceAccountCommand {
	client := &ccloud.Client{
		User: suite.userMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	return NewServiceAccountCommand(prerunner)
}

func (suite *ServiceAccountTestSuite) TestCreateServiceAccountService() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"create", serviceName, "--description", serviceDescription})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.CreateServiceAccountCalled())
}

func (suite *ServiceAccountTestSuite) TestDeleteServiceAccountService() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete", serviceAccountId})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.DeleteServiceAccountCalled())
}

func TestServiceAccountTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountTestSuite))
}
