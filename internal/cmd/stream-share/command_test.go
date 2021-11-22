package stream_share

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccsdkmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	cdxv1 "github.com/confluentinc/cdx-schema/cdx/v1"
	"github.com/confluentinc/cli/internal/cmd/utils"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type SharedTokenTestSuite struct {
	suite.Suite
	conf                  *v1.Config
	streamShareClientMock *ccsdkmock.StreamShare
	analyticsOutput       []segment.Message
	analyticsClient       analytics.Client
}

func TestSharedTokenTestSuite(t *testing.T) {
	suite.Run(t, new(SharedTokenTestSuite))
}

func (suite *SharedTokenTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.streamShareClientMock = &ccsdkmock.StreamShare{
		CreateSharedTokenFunc: func(input *cdxv1.CdxV1CreateSharedTokenRequest) (*cdxv1.CdxV1SharedToken, error) {
			return &cdxv1.CdxV1SharedToken{
				Token: stringToPtr("token"),
			}, nil
		},
		RedeemSharedTokenFunc: func(token string) (*cdxv1.CdxV1RedeemToken, error) {
			return &cdxv1.CdxV1RedeemToken{
				Apikey: stringToPtr("key"),
				Secret: stringToPtr("secret"),
			}, nil
		},
	}
	suite.analyticsOutput = make([]segment.Message, 0)
	suite.analyticsClient = utils.NewTestAnalyticsClient(suite.conf, &suite.analyticsOutput)
}

func (suite *SharedTokenTestSuite) newCmd() *cobra.Command {
	client := &ccloud.Client{
		StreamShare: suite.streamShareClientMock,
	}
	resolverMock := &pcmd.FlagResolverImpl{
		Out: os.Stdout,
	}
	prerunner := &cliMock.Commander{
		FlagResolver: resolverMock,
		Client:       client,
		Config:       suite.conf,
	}
	return New(prerunner, suite.analyticsClient)
}

func (suite *SharedTokenTestSuite) TestCreateSharedTokenReturnsErrorWhenEmailIsInvalid() {
	cmd := suite.newCmd()
	args := []string{"shared-token", "create", "--consumer_email", "confluent.io"}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.NotNil(err)
	req.Equal(errors.BadEmailFormatErrorMsg, err.Error())
	req.False(suite.streamShareClientMock.CreateSharedTokenCalled())
}

func (suite *SharedTokenTestSuite) TestCreateSharedTokenReturnsErrorWhenTopicIsEmpty() {
	cmd := suite.newCmd()
	args := []string{"shared-token", "create", "--consumer_email", "stokkar+provider@confluent.io", "--topic", ""}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.NotNil(err)
	req.Equal(errors.TopicEmptyErrorMsg, err.Error())
	req.False(suite.streamShareClientMock.CreateSharedTokenCalled())
}

func (suite *SharedTokenTestSuite) TestCreateSharedTokenReturnsErrorWhenClusterIsEmpty() {
	cmd := suite.newCmd()
	args := []string{"shared-token", "create", "--consumer_email", "stokkar+provider@confluent.io", "--topic", "test_topic",
		"--cluster", ""}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.NotNil(err)
	req.Equal(errors.ClusterEmptyErrorMsg, err.Error())
	req.False(suite.streamShareClientMock.CreateSharedTokenCalled())
}

func (suite *SharedTokenTestSuite) TestCreateSharedToken() {
	cmd := suite.newCmd()
	args := []string{"shared-token", "create", "--consumer_email", "stokkar+provider@confluent.io", "--topic", "test_topic",
		"--cluster", "clstr"}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.streamShareClientMock.CreateSharedTokenCalled())
}

func (suite *SharedTokenTestSuite) TestRedeemSharedTokenReturnsErrorWhenTokenIsEmpty() {
	cmd := suite.newCmd()
	args := []string{"shared-token", "redeem", "--token", ""}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	req := require.New(suite.T())
	req.NotNil(err)
	req.Equal(errors.TokenEmptyErrorMsg, err.Error())
	req.False(suite.streamShareClientMock.RedeemSharedTokenCalled())
}

func (suite *SharedTokenTestSuite) TestRedeemSharedTokenWritesToDefaultOutputFile() {
	cmd := suite.newCmd()
	outputPath := "./consumer.config"
	args := []string{"shared-token", "redeem", "--token", "test_token"}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	defer os.Remove(outputPath)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.streamShareClientMock.RedeemSharedTokenCalled())
	_, err = os.Stat(outputPath)
	req.NoError(err)
}

func (suite *SharedTokenTestSuite) TestRedeemSharedTokenWritesToOutputFile() {
	cmd := suite.newCmd()
	outputPath := "./file.text"
	args := []string{"shared-token", "redeem", "--token", "test_token", "--output", outputPath}
	err := utils.ExecuteCommandWithAnalytics(cmd, args, suite.analyticsClient)
	defer os.Remove(outputPath)
	req := require.New(suite.T())
	req.Nil(err)
	req.True(suite.streamShareClientMock.RedeemSharedTokenCalled())
	_, err = os.Stat(outputPath)
	req.NoError(err)
}
