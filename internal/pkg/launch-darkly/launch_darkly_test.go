package launchdarkly

import (
	"github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"testing"

	"github.com/dghubble/sling"

	test_server "github.com/confluentinc/cli/test/test-server"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type LaunchDarklyTestSuite struct {
	suite.Suite
	flagManager LaunchDarklyManager
}

func (suite *LaunchDarklyTestSuite) SetupTest() {
	suite.flagManager = LaunchDarklyManager{
		client:  sling.New().Base(test_server.TestCloudURL.Path + "/ldapi/sdk/eval/1234/"),
		version: version.NewVersion("v1.2", "", "", ""),
	}
}

func (suite *LaunchDarklyTestSuite) TestFlags() {
	server := test_server.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	flagManager := LaunchDarklyManager{
		client:  sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version: version.NewVersion("v1.2", "", "", ""),
	}
	ctx := dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	req := require.New(suite.T())

	boolFlag := flagManager.BoolVariation("testBool", ctx, false)
	req.Equal(true, boolFlag)

	flagManager.flagValsAreForAnonUser = true // reset so cache isn't used
	stringFlag := flagManager.StringVariation("testString", ctx, "")
	req.Equal("string", stringFlag)

	flagManager.flagValsAreForAnonUser = true
	intFlag := flagManager.IntVariation("testInt", ctx, 5)
	req.Equal(1, intFlag)

	flagManager.flagValsAreForAnonUser = true
	jsonFlag := flagManager.JsonVariation("testJson", ctx, map[string]string{})
	req.Equal(map[string]interface{}{"key": "val"}, jsonFlag)
}

// Flag variation tests using cached flag values
func (suite *LaunchDarklyTestSuite) TestBoolVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{
		flagVals: map[string]interface{}{"test": true},
	}
	// evaluate cached flags for anon user
	ctx := dynamic_config.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.BoolVariation("test", ctx, false)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
	// evaluate cached flags for logged in user
	ctx = dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.BoolVariation("test", ctx, false)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestIntVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{
		flagVals: map[string]interface{}{"test": 3},
	}
	// evaluate cached flags for anon user
	ctx := dynamic_config.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.IntVariation("test", ctx, 0)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
	// evaluate cached flags for logged in user
	ctx = dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.IntVariation("test", ctx, 0)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestStringVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{
		flagVals: map[string]interface{}{"test": "value"},
	}
	// evaluate cached flags for anon user
	ctx := dynamic_config.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.StringVariation("test", ctx, "")
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
	// evaluate cached flags for logged in user
	ctx = dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.StringVariation("test", ctx, "")
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestJsonVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{
		flagVals: map[string]interface{}{"test": struct {
			key string
			val string
		}{key: "key", val: "val"}},
	}
	// evaluate cached flags for anon user
	ctx := dynamic_config.NewDynamicContext(nil, nil, nil)
	flagMananger.flagValsAreForAnonUser = true
	evaluatedFlag := flagMananger.JsonVariation("test", ctx, nil)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
	// evaluate cached flags for logged in user
	ctx = dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	flagMananger.flagValsAreForAnonUser = false
	evaluatedFlag = flagMananger.JsonVariation("test", ctx, nil)
	req.Equal(flagMananger.flagVals["test"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestContextToLDUser() {
	req := require.New(suite.T())
	ctx := dynamic_config.NewDynamicContext(nil, nil, nil)
	user, anon := suite.flagManager.contextToLDUser(ctx)
	req.True(user.GetAnonymous())
	req.True(anon)

	ctx = dynamic_config.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
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
