package environment

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

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
}

func (suite *EnvironmentTestSuite) newCmd() *cobra.Command {
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

func (suite *EnvironmentTestSuite) TestDeleteEnvironment() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"delete", environmentID})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.accountClientMock.DeleteCalled())
}
