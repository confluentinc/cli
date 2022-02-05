package iam

import (
	"context"
	"testing"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

type InvitationTestSuite struct {
	suite.Suite
	conf     *v1.Config
	userMock *ccsdkmock.User
}

func (suite *InvitationTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.userMock = &ccsdkmock.User{
		GetUserProfileFunc: func(_ context.Context, _ *orgv1.User) (*flowv1.UserProfile, error) {
			return &flowv1.UserProfile{
				FirstName: "TEST",
				LastName:  "lastname",
			}, nil
		},
		DescribeFunc: func(arg0 context.Context, arg1 *orgv1.User) (*orgv1.User, error) {
			return &orgv1.User{
				FirstName: "TEST",
				LastName:  "lastname",
			}, nil
		},
		ListInvitationsFunc: func(_ context.Context) ([]*orgv1.Invitation, error) {
			return []*orgv1.Invitation{
				{
					Id:             "invitation1",
					Email:          "invitation1@confluent.io",
					UserResourceId: "u-1234",
					Status:         "SENT",
				},
				{
					Id:             "invitation2",
					Email:          "invitation2@confluent.io",
					UserResourceId: "u-4321",
					Status:         "PENDING",
				},
			}, nil
		},
		CreateInvitationFunc: func(_ context.Context, arg1 *flowv1.CreateInvitationRequest) (*orgv1.Invitation, error) {
			return &orgv1.Invitation{
				Id:    "invitation1",
				Email: "cli@confluent.io",
			}, nil
		},
	}
}

func (suite *InvitationTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloud.Client{
		User: suite.userMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	return NewUserCommand(prerunner)
}

func (suite *InvitationTestSuite) TestInvitationList() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"invitation", "list"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.NoError(err)
	req.True(suite.userMock.ListInvitationsCalled())
	req.Equal(2, len(suite.userMock.DescribeCalls()))
	req.Equal("u-1234", suite.userMock.DescribeCalls()[0].Arg1.ResourceId)
	req.Equal("u-4321", suite.userMock.DescribeCalls()[1].Arg1.ResourceId)
}

func (suite *InvitationTestSuite) TestCreateInvitation() {
	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	cmd.SetArgs([]string{"invitation", "create", "cli@confluent.io"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.CreateInvitationCalled())
	req.Equal("cli@confluent.io", suite.userMock.CreateInvitationCalls()[0].Arg1.User.Email)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(InvitationTestSuite))
}
