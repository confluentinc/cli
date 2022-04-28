package launchdarkly

import (
	"testing"
	"time"

	"github.com/dghubble/sling"

	"github.com/confluentinc/cli/internal/pkg/version"
	test_server "github.com/confluentinc/cli/test/test-server"

	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type LaunchDarklyTestSuite struct {
	suite.Suite
	authContext *dynamicconfig.DynamicContext
}

func (suite *LaunchDarklyTestSuite) SetupTest() {
	suite.authContext = dynamicconfig.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	ld := LaunchDarklyManager{}
	suite.authContext.LDConfig = &v1.LaunchDarkly{
		FlagValues: map[string]interface{}{"testJson": struct {
			key string
			val string
		}{key: "key", val: "val"}, "testBool": true, "testInt": 3, "testString": "value"},
		FlagUpdateTime: time.Now().Unix(),
		User:           ld.contextToLDUser(suite.authContext),
	}
}

func (suite *LaunchDarklyTestSuite) TestFlags() {
	server := test_server.StartTestCloudServer(suite.T(), false)
	defer server.Close()
	flagManager := LaunchDarklyManager{
		client:  sling.New().Base(server.GetCloudUrl() + "/ldapi/sdk/eval/1234/"),
		version: version.NewVersion("v1.2", "", "", ""),
	}
	ctx := dynamicconfig.NewDynamicContext(v1.AuthenticatedCloudConfigMock().Context(), nil, nil)
	req := require.New(suite.T())

	boolFlag := flagManager.BoolVariation("testBool", ctx, false)
	req.Equal(true, boolFlag)

	stringFlag := flagManager.StringVariation("testString", ctx, "")
	req.Equal("string", stringFlag)

	intFlag := flagManager.IntVariation("testInt", ctx, 5)
	req.Equal(1, intFlag)

	jsonFlag := flagManager.JsonVariation("testJson", ctx, map[string]string{})
	req.Equal(map[string]interface{}{"key": "val"}, jsonFlag)
}

// Flag variation tests using cached flag values
func (suite *LaunchDarklyTestSuite) TestBoolVariation() {
	req := require.New(suite.T())
	flagManager := LaunchDarklyManager{}
	evaluatedFlag := flagManager.BoolVariation("testBool", suite.authContext, false)
	req.Equal(true, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestIntVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{}
	evaluatedFlag := flagMananger.IntVariation("testInt", suite.authContext, 0)
	req.Equal(3, evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestStringVariation() {
	req := require.New(suite.T())
	flagMananger := LaunchDarklyManager{}
	evaluatedFlag := flagMananger.StringVariation("testString", suite.authContext, "")
	req.Equal("value", evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestJsonVariation() {
	req := require.New(suite.T())
	flagManager := LaunchDarklyManager{}

	evaluatedFlag := flagManager.JsonVariation("testJson", suite.authContext, nil)
	req.Equal(suite.authContext.LDConfig.FlagValues["testJson"], evaluatedFlag)
}

func (suite *LaunchDarklyTestSuite) TestContextToLDUser() {
	req := require.New(suite.T())
	flagManager := LaunchDarklyManager{version: version.NewVersion("v1.2", "", "", "")}

	user := flagManager.contextToLDUser(suite.authContext)
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
