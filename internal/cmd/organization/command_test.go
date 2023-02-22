package organization

import (
	"bytes"
	"context"
	"fmt"
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
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	firstOrganizationID     = "org-resource-id"
	firstOrganizationName   = "test-org"
	secondOrganizationID    = "org-resource-id-2"
	secondOrganizationName  = "test-org-2"
	organizationNameUpdated = "test-org-updated"
)

type OrganizationTestSuite struct {
	suite.Suite
	conf              *v1.Config
	accountClientMock *ccloudv1mock.AccountInterface
	V2ClientMock      *V2ClientMock
}

type V2ClientMock struct {
	orgClientMock *orgmock.OrganizationsOrgV2Api
}

var orgOrganization = orgv2.OrgV2Organization{
	Id:          orgv2.PtrString(firstOrganizationID),
	DisplayName: orgv2.PtrString(firstOrganizationName),
}

var orgOrganizationTwo = orgv2.OrgV2Organization{
	Id:          orgv2.PtrString(secondOrganizationID),
	DisplayName: orgv2.PtrString(secondOrganizationName),
}

var orgOrganizationUpdated = orgv2.OrgV2Organization{
	Id:          orgv2.PtrString(firstOrganizationID),
	DisplayName: orgv2.PtrString(organizationNameUpdated),
}

func TestOrganizationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTestSuite))
}

func (suite *OrganizationTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.accountClientMock = &ccloudv1mock.AccountInterface{
		CreateFunc: func(arg0 context.Context, arg1 *ccloudv1.Account) (account *ccloudv1.Account, e error) {
			return &ccloudv1.Account{
				Id:   firstOrganizationID,
				Name: firstOrganizationName,
			}, nil
		},
	}
	orgClientMock := &orgmock.OrganizationsOrgV2Api{
		GetOrgV2OrganizationFunc: func(_ context.Context, _ string) orgv2.ApiGetOrgV2OrganizationRequest {
			return orgv2.ApiGetOrgV2OrganizationRequest{}
		},
		GetOrgV2OrganizationExecuteFunc: func(_ orgv2.ApiGetOrgV2OrganizationRequest) (orgv2.OrgV2Organization, *http.Response, error) {
			return orgOrganization, nil, nil
		},
		ListOrgV2OrganizationsFunc: func(_ context.Context) orgv2.ApiListOrgV2OrganizationsRequest {
			return orgv2.ApiListOrgV2OrganizationsRequest{}
		},
		ListOrgV2OrganizationsExecuteFunc: func(_ orgv2.ApiListOrgV2OrganizationsRequest) (orgv2.OrgV2OrganizationList, *http.Response, error) {
			return orgv2.OrgV2OrganizationList{Data: []orgv2.OrgV2Organization{orgOrganization, orgOrganizationTwo}}, nil, nil
		},
		UpdateOrgV2OrganizationFunc: func(_ context.Context, _ string) orgv2.ApiUpdateOrgV2OrganizationRequest {
			return orgv2.ApiUpdateOrgV2OrganizationRequest{}
		},
		UpdateOrgV2OrganizationExecuteFunc: func(_ orgv2.ApiUpdateOrgV2OrganizationRequest) (orgv2.OrgV2Organization, *http.Response, error) {
			return orgOrganizationUpdated, nil, nil
		},
	}
	suite.V2ClientMock = &V2ClientMock{orgClientMock: orgClientMock}
}

func (suite *OrganizationTestSuite) newCmd() *cobra.Command {
	client := &ccloudv1.Client{
		Account: suite.accountClientMock,
	}
	orgClient := &orgv2.APIClient{
		OrganizationsOrgV2Api: suite.V2ClientMock.orgClientMock,
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		MDSClient:    nil,
		V2Client:     &ccloudv2.Client{OrgClient: orgClient, AuthToken: "auth-token"},
		Config:       suite.conf,
	}
	return New(prerunner)
}

func (suite *OrganizationTestSuite) TestDescribeOrganizations() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"describe"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.V2ClientMock.orgClientMock.GetOrgV2OrganizationCalled())
	req.True(suite.V2ClientMock.orgClientMock.GetOrgV2OrganizationExecuteCalled())
}

func (suite *OrganizationTestSuite) TestListOrganizations() {
	cmd := suite.newCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	got := buf.String()
	req.Contains(got, firstOrganizationID)
	req.Contains(got, firstOrganizationName)
	req.Contains(got, secondOrganizationID)
	req.Contains(got, secondOrganizationName)
	req.True(suite.V2ClientMock.orgClientMock.ListOrgV2OrganizationsCalled())
	req.True(suite.V2ClientMock.orgClientMock.ListOrgV2OrganizationsExecuteCalled())
}

func (suite *OrganizationTestSuite) TestUpdateOrganization() {
	cmd := suite.newCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"update", "--name", organizationNameUpdated})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.Equal([]byte(fmt.Sprintf("Updated the name of organization \"%s\" to \"%s\".\n", firstOrganizationID, organizationNameUpdated)), buf.Bytes())
	req.True(suite.V2ClientMock.orgClientMock.UpdateOrgV2OrganizationCalled())
	req.True(suite.V2ClientMock.orgClientMock.UpdateOrgV2OrganizationExecuteCalled())
}
