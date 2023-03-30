package featureflags

import (
	"testing"
	"time"

	"github.com/dghubble/sling"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

type LaunchDarklyTestSuite struct {
	suite.Suite
	ctx *dynamicconfig.DynamicContext
}

func (suite *LaunchDarklyTestSuite) SetupTest() {
	suite.ctx = dynamicconfig.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)

	type kv struct {
		key string
		val string
	}

	ld := launchDarklyManager{}
	suite.ctx.FeatureFlags = &v1.FeatureFlags{
		Values:         map[string]any{"testJson": kv{key: "key", val: "val"}, "testBool": true, "testInt": 3, "testString": "value"},
		LastUpdateTime: time.Now().Unix(),
		User:           ld.contextToLDUser(suite.ctx),
	}
}

func (suite *LaunchDarklyTestSuite) TestFlags() {
	server := testserver.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	ld := launchDarklyManager{
		cliClient: sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version:   version.NewVersion("v1.2", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	req := require.New(suite.T())

	boolFlag := ld.BoolVariation("testBool", ctx, v1.CliLaunchDarklyClient, true, false)
	req.Equal(true, boolFlag)

	stringFlag := ld.StringVariation("testString", ctx, v1.CliLaunchDarklyClient, true, "")
	req.Equal("string", stringFlag)

	intFlag := ld.IntVariation("testInt", ctx, v1.CliLaunchDarklyClient, true, 5)
	req.Equal(1, intFlag)

	jsonFlag := ld.JsonVariation("testJson", ctx, v1.CliLaunchDarklyClient, true, map[string]string{})
	req.Equal(map[string]any{"key": "val"}, jsonFlag)
}

// Flag variation tests using cached flag values
func (suite *LaunchDarklyTestSuite) TestBoolVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.BoolVariation("testBool", suite.ctx, v1.CliLaunchDarklyClient, true, false)
	req.Equal(true, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestIntVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.IntVariation("testInt", suite.ctx, v1.CliLaunchDarklyClient, true, 0)
	req.Equal(3, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestStringVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}
	evaluatedFlag := ld.StringVariation("testString", suite.ctx, v1.CliLaunchDarklyClient, true, "")
	req.Equal("value", evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestJsonVariation() {
	req := require.New(suite.T())
	ld := launchDarklyManager{}

	evaluatedFlag := ld.JsonVariation("testJson", suite.ctx, v1.CliLaunchDarklyClient, true, nil)
	req.Equal(suite.ctx.FeatureFlags.Values["testJson"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestContextToLDUser() {
	req := require.New(suite.T())
	ld := launchDarklyManager{version: version.NewVersion("v1.2", "", "")}

	user := ld.contextToLDUser(suite.ctx)
	resourceId, _ := user.GetCustom("user.resource_id")
	req.Equal(v1.MockUserResourceId, resourceId.StringValue())
	ver, _ := user.GetCustom("cli.version")
	req.Equal("v1.2", ver.StringValue())
	orgResourceId, _ := user.GetCustom("org.resource_id")
	req.Equal(v1.MockOrgResourceId, orgResourceId.StringValue())
	environmentId, _ := user.GetCustom("environment.id")
	req.Equal(v1.MockEnvironmentId, environmentId.StringValue())
}

func TestLaunchDarklySuite(t *testing.T) {
	suite.Run(t, new(LaunchDarklyTestSuite))
}
