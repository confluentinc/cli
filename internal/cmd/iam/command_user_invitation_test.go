package iam

import (
	"context"
	"net/http"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

type InvitationTestSuite struct {
	suite.Suite
	conf           *v1.Config
	userMock       *ccsdkmock.User
	invitationMock *iammock.InvitationsIamV2Api
}

func (suite *InvitationTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.userMock = &ccsdkmock.User{
		DescribeFunc: func(arg0 context.Context, arg1 *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				FirstName: "TEST",
				LastName:  "lastname",
			}, nil
		},
	}
	suite.invitationMock = &iammock.InvitationsIamV2Api{
		CreateIamV2InvitationFunc: func(_ context.Context) iamv2.ApiCreateIamV2InvitationRequest {
			return iamv2.ApiCreateIamV2InvitationRequest{}
		},
		CreateIamV2InvitationExecuteFunc: func(req iamv2.ApiCreateIamV2InvitationRequest) (iamv2.IamV2Invitation, *http.Response, error) {
			return iamv2.IamV2Invitation{Id: iamv2.PtrString("invitation1"), Email: iamv2.PtrString("cli@confluent.io")}, nil, nil
		},
		ListIamV2InvitationsFunc: func(ctx context.Context) iamv2.ApiListIamV2InvitationsRequest {
			return iamv2.ApiListIamV2InvitationsRequest{}
		},
		ListIamV2InvitationsExecuteFunc: func(r iamv2.ApiListIamV2InvitationsRequest) (iamv2.IamV2InvitationList, *http.Response, error) {
			return iamv2.IamV2InvitationList{
				Data: []iamv2.IamV2Invitation{
					{
						Id:     iamv2.PtrString("invitation1"),
						Email:  iamv2.PtrString("invitation1@confluent.io"),
						User:   &iamv2.GlobalObjectReference{Id: "u-1234"},
						Status: iamv2.PtrString("SENT"),
					},
					{
						Id:     iamv2.PtrString("invitation2"),
						Email:  iamv2.PtrString("invitation2@confluent.io"),
						User:   &iamv2.GlobalObjectReference{Id: "u-4321"},
						Status: iamv2.PtrString("PENDING"),
					},
				},
			}, nil, nil
		},
	}
}

func (suite *InvitationTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloud.Client{
		User: suite.userMock,
	}
	iamClient := &iamv2.APIClient{
		InvitationsIamV2Api: suite.invitationMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, &ccloudv2.Client{IamClient: iamClient}, nil, nil, conf)
	return newUserCommand(prerunner)
}

func (suite *InvitationTestSuite) TestInvitationList() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"invitation", "list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.invitationMock.ListIamV2InvitationsCalled())
	req.True(suite.invitationMock.ListIamV2InvitationsExecuteCalled())
	req.Equal(2, len(suite.userMock.DescribeCalls()))
}

func (suite *InvitationTestSuite) TestCreateInvitation() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"invitation", "create", "cli@confluent.io"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.invitationMock.CreateIamV2InvitationCalled())
	req.True(suite.invitationMock.CreateIamV2InvitationExecuteCalled())
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(InvitationTestSuite))
}
