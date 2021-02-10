package ksql

import (
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId   = "some id"
	clusterName = "clustertruck"
)

type KsqlClusterTestSuite struct {
	suite.Suite
}

func (suite *KsqlClusterTestSuite) newCmd(conf *v3.Config) *clusterCommand {
	client := &ccloud.Client{
		KSQL: &ccsdkmock.KSQL{
			ListFunc: func(arg0 context.Context, arg1 *schedv1.KSQLCluster) (clusters []*schedv1.KSQLCluster, err error) {
				return []*schedv1.KSQLCluster{
					{
						Id:   clusterId,
						Name: clusterName,
					},
				}, nil
			},
		},
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	cmd := NewClusterCommand(prerunner, cliMock.NewDummyAnalyticsMock())
	return cmd
}

func (suite *KsqlClusterTestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *clusterCommand
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(v3.AuthenticatedCloudConfigMock()),
			},
			want: []prompt.Suggest{
				{
					Text:        clusterId,
					Description: clusterName,
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				suite.newCmd(v3.UnauthenticatedCloudConfigMock()),
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

func (suite *KsqlClusterTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"app describe", "app delete", "app configure-acls"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func TestKsqlClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KsqlClusterTestSuite))
}
