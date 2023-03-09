package iam

import (
	"context"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iamMock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
)

const (
	UserId             = int32(123)
	serviceAccountId   = "sa-55555"
	serviceDescription = "testing"
	serviceName        = "demo"
)

type ServiceAccountTestSuite struct {
	suite.Suite
	conf                  *v1.Config
	iamServiceAccountMock *iamMock.ServiceAccountsIamV2Api
}

var iamServiceAccount = iamv2.IamV2ServiceAccount{
	Id:          iamv2.PtrString(serviceAccountId),
	DisplayName: iamv2.PtrString(serviceName),
	Description: iamv2.PtrString(serviceDescription),
}

func (suite *ServiceAccountTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.iamServiceAccountMock = &iamMock.ServiceAccountsIamV2Api{
		GetIamV2ServiceAccountFunc: func(_ context.Context, _ string) iamv2.ApiGetIamV2ServiceAccountRequest {
			return iamv2.ApiGetIamV2ServiceAccountRequest{}
		},
		GetIamV2ServiceAccountExecuteFunc: func(_ iamv2.ApiGetIamV2ServiceAccountRequest) (iamv2.IamV2ServiceAccount, *http.Response, error) {
			return iamServiceAccount, nil, nil
		},
		CreateIamV2ServiceAccountFunc: func(_ context.Context) iamv2.ApiCreateIamV2ServiceAccountRequest {
			return iamv2.ApiCreateIamV2ServiceAccountRequest{}
		},
		CreateIamV2ServiceAccountExecuteFunc: func(_ iamv2.ApiCreateIamV2ServiceAccountRequest) (iamv2.IamV2ServiceAccount, *http.Response, error) {
			return iamServiceAccount, nil, nil
		},
		DeleteIamV2ServiceAccountFunc: func(_ context.Context, _ string) iamv2.ApiDeleteIamV2ServiceAccountRequest {
			return iamv2.ApiDeleteIamV2ServiceAccountRequest{}
		},
		DeleteIamV2ServiceAccountExecuteFunc: func(_ iamv2.ApiDeleteIamV2ServiceAccountRequest) (*http.Response, error) {
			return nil, nil
		},
	}
}

func (suite *ServiceAccountTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	iamClient := &iamv2.APIClient{
		ServiceAccountsIamV2Api: suite.iamServiceAccountMock,
	}
	prerunner := climock.NewPreRunnerMock(nil, &ccloudv2.Client{IamClient: iamClient, AuthToken: "auth-token"}, nil, nil, conf)
	return newServiceAccountCommand(prerunner)
}

func (suite *ServiceAccountTestSuite) TestCreateServiceAccountService() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"create", serviceName, "--description", serviceDescription})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.iamServiceAccountMock.CreateIamV2ServiceAccountCalled())
}

func (suite *ServiceAccountTestSuite) TestDeleteServiceAccountService() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"delete", serviceAccountId, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.iamServiceAccountMock.DeleteIamV2ServiceAccountCalled())
}

func TestServiceAccountTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceAccountTestSuite))
}
