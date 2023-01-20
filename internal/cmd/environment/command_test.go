package environment

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	orgmock "github.com/confluentinc/ccloud-sdk-go-v2/org/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
)

const (
	environmentID          = "env-123"
	environmentName        = "test-env"
	environmentNameUpdated = "test-env-updated"
)

type EnvironmentTestSuite struct {
	suite.Suite
	conf              *v1.Config
	accountClientMock *ccloudv1mock.AccountInterface
	V2ClientMock      *V2ClientMock
}

type V2ClientMock struct {
	orgClientMock *orgmock.EnvironmentsOrgV2Api
}

var orgEnvironment = orgv2.OrgV2Environment{
	Id:          orgv2.PtrString(environmentID),
	DisplayName: orgv2.PtrString(environmentName),
}

var orgEnvironmentUpdated = orgv2.OrgV2Environment{
	Id:          orgv2.PtrString(environmentID),
	DisplayName: orgv2.PtrString(environmentNameUpdated),
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}

func (suite *EnvironmentTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.accountClientMock = &ccloudv1mock.AccountInterface{
		CreateFunc: func(arg0 context.Context, arg1 *ccloudv1.Account) (account *ccloudv1.Account, e error) {
			return &ccloudv1.Account{
				Id:   environmentID,
				Name: environmentName,
			}, nil
		},
	}
	orgClientMock := &orgmock.EnvironmentsOrgV2Api{
		GetOrgV2EnvironmentFunc: func(_ context.Context, _ string) orgv2.ApiGetOrgV2EnvironmentRequest {
			return orgv2.ApiGetOrgV2EnvironmentRequest{}
		},
		GetOrgV2EnvironmentExecuteFunc: func(_ orgv2.ApiGetOrgV2EnvironmentRequest) (orgv2.OrgV2Environment, *http.Response, error) {
			return orgEnvironment, nil, nil
		},
		ListOrgV2EnvironmentsFunc: func(_ context.Context) orgv2.ApiListOrgV2EnvironmentsRequest {
			return orgv2.ApiListOrgV2EnvironmentsRequest{}
		},
		ListOrgV2EnvironmentsExecuteFunc: func(_ orgv2.ApiListOrgV2EnvironmentsRequest) (orgv2.OrgV2EnvironmentList, *http.Response, error) {
			return orgv2.OrgV2EnvironmentList{Data: []orgv2.OrgV2Environment{orgEnvironment, orgEnvironmentUpdated}}, nil, nil
		},
		UpdateOrgV2EnvironmentFunc: func(_ context.Context, _ string) orgv2.ApiUpdateOrgV2EnvironmentRequest {
			return orgv2.ApiUpdateOrgV2EnvironmentRequest{}
		},
		UpdateOrgV2EnvironmentExecuteFunc: func(_ orgv2.ApiUpdateOrgV2EnvironmentRequest) (orgv2.OrgV2Environment, *http.Response, error) {
			return orgEnvironmentUpdated, nil, nil
		},
		DeleteOrgV2EnvironmentFunc: func(_ context.Context, _ string) orgv2.ApiDeleteOrgV2EnvironmentRequest {
			return orgv2.ApiDeleteOrgV2EnvironmentRequest{}
		},
		DeleteOrgV2EnvironmentExecuteFunc: func(_ orgv2.ApiDeleteOrgV2EnvironmentRequest) (*http.Response, error) {
			return nil, nil
		},
	}
	suite.V2ClientMock = &V2ClientMock{orgClientMock: orgClientMock}
}

func (suite *EnvironmentTestSuite) newCmd() *cobra.Command {
	client := &ccloudv1.Client{
		Account: suite.accountClientMock,
	}
	orgClient := &orgv2.APIClient{
		EnvironmentsOrgV2Api: suite.V2ClientMock.orgClientMock,
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &climock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		MDSClient:    nil,
		V2Client:     &ccloudv2.Client{OrgClient: orgClient, AuthToken: "auth-token"},
		Config:       suite.conf,
	}
	return New(prerunner)
}

func (suite *EnvironmentTestSuite) TestCreateEnvironment() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"create", environmentName})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.accountClientMock.CreateCalled())
}

func (suite *EnvironmentTestSuite) TestListEnvironments() {
	cmd := suite.newCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	got := buf.String()
	req.Contains(got, environmentID)
	req.Contains(got, environmentName)
	req.Contains(got, environmentNameUpdated)
	req.True(suite.V2ClientMock.orgClientMock.ListOrgV2EnvironmentsCalled())
}

func (suite *EnvironmentTestSuite) TestDeleteEnvironment() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"delete", environmentID, "--force"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.V2ClientMock.orgClientMock.DeleteOrgV2EnvironmentCalled())
}

func (suite *EnvironmentTestSuite) TestUpdateEnvironment() {
	cmd := suite.newCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"update", environmentID, "--name", "test-env-updated"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.Equal([]byte("Updated the name of environment \"env-123\" to \"test-env-updated\".\n"), buf.Bytes())
	req.True(suite.V2ClientMock.orgClientMock.UpdateOrgV2EnvironmentCalled())
}
