package environment

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/c-bata/go-prompt"
	segment "github.com/segmentio/analytics-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	environmentID   = "env-123"
	environmentName = "test-env"
)

type EnvironmentTestSuite struct {
	suite.Suite
	conf              *v1.Config
	accountClientMock *ccsdkmock.Account
	analyticsOutput   []segment.Message
	analyticsClient   analytics.Client
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}

func (suite *EnvironmentTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.accountClientMock = &ccsdkmock.Account{
		CreateFunc: func(arg0 context.Context, arg1 *orgv1.Account) (account *orgv1.Account, e error) {
			return &orgv1.Account{
				Id:   environmentID,
				Name: environmentName,
			}, nil
		},
		GetFunc: func(arg0 context.Context, arg1 *orgv1.Account) (account *orgv1.Account, e error) {
			return &orgv1.Account{
				Id:   environmentID,
				Name: environmentName,
			}, nil
		},
		ListFunc: func(arg0 context.Context, arg1 *orgv1.Account) (accounts []*orgv1.Account, e error) {
			return []*orgv1.Account{
				{
					Id:   environmentID,
					Name: environmentName,
				},
			}, nil
		},
		UpdateFunc: func(arg0 context.Context, arg1 *orgv1.Account) error {
			return nil
		},
		DeleteFunc: func(arg0 context.Context, arg1 *orgv1.Account) error {
			return nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *EnvironmentTestSuite) newCmd() *command {
	client := &ccloud.Client{
		Account: suite.accountClientMock,
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		MDSClient:    nil,
		Config:       suite.conf,
	}
	return New(prerunner, suite.analyticsClient)
}

func (suite *EnvironmentTestSuite) TestCreateEnvironment() {
	cmd := suite.newCmd()
	args := []string{"create", environmentName}
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.accountClientMock.CreateCalled())
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], environmentID, req)
}

func (suite *EnvironmentTestSuite) TestDeleteEnvironment() {
	cmd := suite.newCmd()
	args := []string{"delete", environmentID}
	err := utils.ExecuteCommandWithAnalytics(cmd.Command, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.accountClientMock.DeleteCalled())
	// TODO add back with analytics
	//test_utils.CheckTrackedResourceIDString(suite.analyticsOutput[0], environmentID, req)
}

func (suite *EnvironmentTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd()
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"environment delete", "environment update", "environment use"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func (suite *EnvironmentTestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *command
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(),
			},
			want: []prompt.Suggest{
				{
					Text:        environmentID,
					Description: environmentName,
				},
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_ = tt.fields.Command.PersistentPreRunE(tt.fields.Command.Command, []string{})
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}
