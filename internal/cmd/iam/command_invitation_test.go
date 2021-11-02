package iam

import (
	"context"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/spf13/cobra"
	"testing"

	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

type InvitationTestSuite struct {
	suite.Suite
	conf            *v1.Config
	userMock        *ccsdkmock.User
	analyticsOutput []segment.Message
	analyticsClient analytics.Client
}

func (suite *InvitationTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.userMock = &ccsdkmock.User{
		GetUserProfileFunc: func(_ context.Context, _ *orgv1.User) (*flowv1.UserProfile, error) {
			return &flowv1.UserProfile{
				FirstName: "TEST",
				LastName: "lastname",
			}, nil
		},
		ListInvitationsFunc: func(_ context.Context) ([]*orgv1.Invitation, error) {
			return []*orgv1.Invitation{
				{
					Id:             "invitation1",
					Email:          "invitation1@confluent.io",
					UserResourceId: "user1",
					Status:         "SENT",
				},
				{
					Id:             "invitation2",
					Email:          "invitation2@confluent.io",
					UserResourceId: "user2",
					Status:         "PENDING",
				},
			}, nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *InvitationTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	client := &ccloud.Client{
		User: suite.userMock,
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	return NewInvitationCommand(prerunner)
}

func (suite *InvitationTestSuite) TestInvitationList() {

	cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
	args := []string{"list"}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.userMock.ListInvitationsCalled())
	req.Equal(2, len(suite.userMock.GetUserProfileCalls()))
	req.Equal("user1", suite.userMock.GetUserProfileCalls()[0].Arg1.ResourceId)
	req.Equal("user2", suite.userMock.GetUserProfileCalls()[1].Arg1.ResourceId)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(InvitationTestSuite))
}
