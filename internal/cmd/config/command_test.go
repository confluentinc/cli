package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/auth"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/mock"
)

type ConfigCommandTestSuite struct {
	suite.Suite
	conf *v3.Config
}

func (suite *ConfigCommandTestSuite) SetupTest() {
	suite.conf = v3.AuthenticatedConfluentConfigMock()
}

func (suite *ConfigCommandTestSuite) newCmd() *cobra.Command {
	prerunner := &mock.Commander{
		Config: suite.conf,
	}
	analytics := mock.NewDummyAnalyticsMock()
	return New("confluent", prerunner, analytics)
}

func (suite *ConfigCommandTestSuite) TestStatisticsDisable() {
	req := require.New(suite.T())

	suite.conf.Context().DisableTracking = false

	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"context", "statistics", "disable"}))
	err := cmd.Execute()
	req.Nil(err)

	req.True(suite.conf.Context().DisableTracking)
}

func (suite *ConfigCommandTestSuite) TestStatisticsEnable() {
	req := require.New(suite.T())

	suite.conf.Context().DisableTracking = true

	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"context", "statistics", "enable"}))
	err := cmd.Execute()
	req.Nil(err)

	req.False(suite.conf.Context().DisableTracking)
}

func (suite *ConfigCommandTestSuite) TestStatisticsDisableFailedForNotLoggedInUser() {
	req := require.New(suite.T())

	ctx := suite.conf.Context()
	ctx.DisableTracking = false
	_ = auth.PersistLogoutToConfig(suite.conf)

	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"context", "statistics", "disable"}))
	err := cmd.Execute()
	req.Error(err)
	req.Contains(err.Error(), errors.NotLoggedInErrorMsg)

	// should have no effects
	req.False(ctx.DisableTracking)
}

func (suite *ConfigCommandTestSuite) TestStatisticsEnableFailedForNotLoggedInUser() {
	req := require.New(suite.T())

	ctx := suite.conf.Context()
	ctx.DisableTracking = true
	_ = auth.PersistLogoutToConfig(suite.conf)

	cmd := suite.newCmd()
	cmd.SetArgs(append([]string{"context", "statistics", "enable"}))
	err := cmd.Execute()
	req.Error(err)
	req.Contains(err.Error(), errors.NotLoggedInErrorMsg)

	// should have no effects
	req.True(ctx.DisableTracking)
}

func TestConfigCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigCommandTestSuite))
}
