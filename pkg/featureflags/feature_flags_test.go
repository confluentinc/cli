package featureflags

import (
	"testing"
	"time"

	"github.com/dghubble/sling"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/version"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

type LaunchDarklyTestSuite struct {
	suite.Suite
	ctx *dynamicconfig.DynamicContext
}

func (suite *LaunchDarklyTestSuite) SetupTest() {
	suite.ctx = dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)

	type kv struct {
		key string
		val string
	}

	ld := launchDarklyManager{}
	suite.ctx.FeatureFlags = &config.FeatureFlags{
		CliValues:      map[string]any{"testJson": kv{key: "key", val: "val"}, "testBool": true, "testInt": 3, "testString": "value", "testAnotherInt": 99},
		LastUpdateTime: time.Now().Unix(),
		User:           ld.contextToLDUser(suite.ctx),
	}
}

func (suite *LaunchDarklyTestSuite) TestFlags() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		cliClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:   version.NewVersion("1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	boolFlag := ld.BoolVariation("testBool", ctx, config.CliLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)

	stringFlag := ld.StringVariation("testString", ctx, config.CliLaunchDarklyClient, true, "")
	req.Equal("string", stringFlag)

	intFlag := ld.IntVariation("testInt", ctx, config.CliLaunchDarklyClient, true, 5)
	req.Equal(1, intFlag)

	jsonFlag := ld.JsonVariation("testJson", ctx, config.CliLaunchDarklyClient, true, map[string]string{})
	req.Equal(map[string]any{"key": "val"}, jsonFlag)
}

func (suite *LaunchDarklyTestSuite) TestCcloudFlags() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		ccloudClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:      version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	boolFlag := ld.BoolVariation("testBool", ctx, config.CcloudStagLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)
}

func (suite *LaunchDarklyTestSuite) TestCcloudFlagsCached() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		ccloudClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:      version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	intFlag := ld.IntVariation("testInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(1, intFlag)

	ctx.FeatureFlags.CcloudValues["testInt"] = 123
	intFlag = ld.IntVariation("testInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(123, intFlag)
}

func (suite *LaunchDarklyTestSuite) TestCcloudFlagsCachedExpired() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		ccloudClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:      version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	intFlag := ld.IntVariation("testInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(1, intFlag)

	ctx.FeatureFlags.CcloudValues["testInt"] = 123
	intFlag = ld.IntVariation("testInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(123, intFlag)

	// cache expired, fetch value from mock
	ctx.FeatureFlags.LastUpdateTime = 0
	intFlag = ld.IntVariation("testInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(1, intFlag)
}

func (suite *LaunchDarklyTestSuite) TestCliFlagsCachedExpired() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		cliClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:   version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	boolFlag := ld.BoolVariation("testBool", ctx, config.CliLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)

	// updating the cached value
	ctx.FeatureFlags.CliValues["testBool"] = false
	boolFlag = ld.BoolVariation("testBool", ctx, config.CliLaunchDarklyClient, true, false)
	req.Equal(false, boolFlag)

	// expiring the cache: updated at 0
	ctx.FeatureFlags.LastUpdateTime = 0
	boolFlag = ld.BoolVariation("testBool", ctx, config.CliLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)
}

func (suite *LaunchDarklyTestSuite) TestMultipleCcloudFlags() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		ccloudClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:      version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(config.AuthenticatedCloudConfigMock().Context(), nil)
	req := require.New(suite.T())

	boolFlag := ld.BoolVariation("testBool", ctx, config.CcloudStagLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)

	intFlag := ld.IntVariation("testAnotherInt", ctx, config.CcloudStagLaunchDarklyClient, true, 0)
	req.Equal(99, intFlag)
}

// Flag variation tests using cached flag values
func (suite *LaunchDarklyTestSuite) TestBoolVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.BoolVariation("testBool", suite.ctx, config.CliLaunchDarklyClient, true, false)
	req.Equal(true, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestIntVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.IntVariation("testInt", suite.ctx, config.CliLaunchDarklyClient, true, 0)
	req.Equal(3, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestStringVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.StringVariation("testString", suite.ctx, config.CliLaunchDarklyClient, true, "")
	req.Equal("value", evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestJsonVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}

	evaluatedFlag := ld.JsonVariation("testJson", suite.ctx, config.CliLaunchDarklyClient, true, nil)
	req.Equal(suite.ctx.FeatureFlags.CliValues["testJson"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestContextToLDUser() {
	req := require.New(suite.T())
	ld := launchDarklyManager{version: version.NewVersion("1.2", "", "")}

	user := ld.contextToLDUser(suite.ctx)
	resourceId, _ := user.GetCustom("user.resource_id")
	req.Equal(config.MockUserResourceId, resourceId.StringValue())
	ver, _ := user.GetCustom("cli.version")
	req.Equal("v1.2", ver.StringValue())
	orgResourceId, _ := user.GetCustom("org.resource_id")
	req.Equal(config.MockOrgResourceId, orgResourceId.StringValue())
	environmentId, _ := user.GetCustom("environment.id")
	req.Equal(config.MockEnvironmentId, environmentId.StringValue())
}

func TestLaunchDarklySuite(t *testing.T) {
	suite.Run(t, new(LaunchDarklyTestSuite))
}
