package environment

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/c-bata/go-prompt"
	v1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

type EnvironmentTestSuite struct {
	suite.Suite
	conf *v3.Config
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}

func (suite *EnvironmentTestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedCloudConfigMock()
}

func (suite *EnvironmentTestSuite) newCmd() *command {
	client := &ccloud.Client{
		Account: &ccsdkmock.Account{
			ListFunc: func(context.Context, *v1.Account) ([]*v1.Account, error) {
				return []*v1.Account{
					{
						Id:   "123",
						Name: "456",
					},
				}, nil
			},
		},
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
	return New("ccloud", prerunner)
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
					Text:        "123",
					Description: "456",
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				Command: func() *command {
					oldConf := suite.conf
					suite.conf = v3.UnauthenticatedCloudConfigMock()
					c := suite.newCmd()
					suite.conf = oldConf
					return c
				}(),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}
