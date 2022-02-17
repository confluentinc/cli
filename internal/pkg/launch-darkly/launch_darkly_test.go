package launch_darkly

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LaunchDarklyTestSuite struct {
	suite.Suite
	flagManager FeatureFlagManager
}

func (suite *LaunchDarklyTestSuite) SetupTest() {
	suite.flagManager = FeatureFlagManager{
		client:                 nil,
		version:                version.NewVersion("v1.2", "", "", ""),
	}
}

// For the flag variation tests, only testing use of cached flags to avoid API calls (using cc-system-tests for this)
func (suite *LaunchDarklyTestSuite) TestBoolVariation() {
	req := require.New(suite.T())
	flagMananger := FeatureFlagManager{
		flagVals: map[string]interface{}{"test":true},
	}
	ctx := cmd.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.BoolVariation("test", ctx, false)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)

	ctx = cmd.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.BoolVariation("test", ctx, false)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestStringVariation() {
	req := require.New(suite.T())
	flagMananger := FeatureFlagManager{
		flagVals: map[string]interface{}{"test":"value"},
	}
	ctx := cmd.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.StringVariation("test", ctx, "")
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)

	ctx = cmd.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.StringVariation("test", ctx, "")
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestJsonVariation() {
	req := require.New(suite.T())
	flagMananger := FeatureFlagManager{
		flagVals: map[string]interface{}{"test": struct {
			key string
			val string
		}{key: "key", val: "val"}},
	}
	ctx := cmd.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.JsonVariation("test", ctx, nil)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)

	ctx = cmd.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.JsonVariation("test", ctx, nil)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestContextToLDUser() {
	req := require.New(suite.T())
	ctx := cmd.NewDynamicContext(nil, nil, nil)
	user, anon := suite.flagManager.contextToLDUser(ctx)
	req.True(user.GetAnonymous())
	req.True(anon)

	ctx = cmd.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	user, anon = suite.flagManager.contextToLDUser(ctx)
	req.False(anon)
	resourceId, _ := user.GetCustom("user.resource_id")
	req.Equal(v1.MockUserResourceId, resourceId.StringValue())
	version, _ := user.GetCustom("cli.version")
	req.Equal("v1.2", version.StringValue())
	orgResourceId, _ := user.GetCustom("org.resource_id")
	req.Equal(v1.MockOrgResourceId, orgResourceId.StringValue())
	environmentId, _ := user.GetCustom("environment.id")
	req.Equal(v1.MockEnvironmentId, environmentId.StringValue())
	clusterId, _ := user.GetCustom("cluster.id")
	req.Equal(v1.MockKafkaClusterId(), clusterId.StringValue())
	pkc, _ := user.GetCustom("cluster.physicalClusterId")
	req.Equal("pkc-abc123", pkc.StringValue())
}

func TestLaunchDarklySuite(t *testing.T) {
	suite.Run(t, new(LaunchDarklyTestSuite))
}