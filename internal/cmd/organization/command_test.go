package organization

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	orgmock "github.com/confluentinc/ccloud-sdk-go-v2/org/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	firstOrganizationID    = "abc-123"
	firstOrganizationName  = "test-org"
	secondOrganizationID   = "abc-456"
	secondOrganizationName = "test-org-2"
)

type OrganizationTestSuite struct {
	suite.Suite
	conf              *v1.Config
	accountClientMock *ccsdkmock.Account
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

func TestOrganizationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTestSuite))
}

func (suite *OrganizationTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.accountClientMock = &ccsdkmock.Account{
		CreateFunc: func(arg0 context.Context, arg1 *orgv1.Account) (account *orgv1.Account, e error) {
			return &orgv1.Account{
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
	}
	suite.V2ClientMock = &V2ClientMock{orgClientMock: orgClientMock}
}

func (suite *OrganizationTestSuite) newCmd() *cobra.Command {
	privateClient := &ccloud.Client{
		Account: suite.accountClientMock,
	}
	orgClient := &orgv2.APIClient{
		OrganizationsOrgV2Api: suite.V2ClientMock.orgClientMock,
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver:  resolverMock,
		PrivateClient: privateClient,
		MDSClient:     nil,
		V2Client:      &ccloudv2.Client{OrgClient: orgClient, AuthToken: "auth-token"},
		Config:        suite.conf,
	}
	return New(prerunner)
}

func (suite *OrganizationTestSuite) TestDescribeOrganizations() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"describe", firstOrganizationID})
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
