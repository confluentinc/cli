//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/launch_darkly.go --pkg mock --selfpkg github.com/confluentinc/cli launch_darkly.go LaunchDarklyManager

package featureflags

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
	testserver "github.com/confluentinc/cli/test/test-server"
)

const (
	baseURL  = "%s/ldapi/sdk/eval/%s/"
	userPath = "users/%s"
)

const (
	cliProdEnvClientId     = "61af57740127630ce47de5be"
	cliTestEnvClientId     = "61af57740127630ce47de5bd"
	ccloudProdEnvClientId  = "5c636508aa445d32c86f26b1"
	ccloudStagEnvClientId  = "5c63651f1df21432a45fc773"
	ccloudDevelEnvClientId = "5c63653912b6db32db950445"
	ccloudCpdEnvClientId   = "5c6365ffaa445d32c86f26c0"
)

// Manager is a global feature flag manager
var Manager featureFlagManager

var attributes = []string{"user.resource_id", "org.resource_id", "environment.id", "cli.version", "cluster.id", "cluster.physicalClusterId"}

type featureFlagManager interface {
	BoolVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal bool) bool
	StringVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal string) string
	IntVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal int) int
	JsonVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal interface{}) interface{}
}

type launchDarklyManager struct {
	cliClient    *sling.Sling
	ccloudClient func(v1.LaunchDarklyClient) *sling.Sling
	version      *version.Version
}

func Init(version *version.Version, isTest bool) {
	cliBasePath := fmt.Sprintf(baseURL, auth.CCloudURL, cliProdEnvClientId)
	if isTest {
		cliBasePath = fmt.Sprintf(baseURL, testserver.TestCloudURL.Path, "1234")
	} else if os.Getenv("XX_LAUNCH_DARKLY_TEST_ENV") != "" {
		cliBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, cliTestEnvClientId)
	}

	ccloudClientProvider := func(client v1.LaunchDarklyClient) *sling.Sling {
		if isTest || os.Getenv("XX_LAUNCH_DARKLY_TEST_ENV") != "" {
			return sling.New().Base(fmt.Sprintf(baseURL, auth.CCloudURL, ccloudCpdEnvClientId))
		}

		var ccloudBasePath string
		switch client {
		case v1.CcloudDevelLaunchDarklyClient:
			ccloudBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, ccloudDevelEnvClientId)
		case v1.CcloudStagLaunchDarklyClient:
			ccloudBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, ccloudStagEnvClientId)
		default:
			ccloudBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, ccloudProdEnvClientId)
		}

		return sling.New().Base(ccloudBasePath)
	}

	Manager = &launchDarklyManager{
		cliClient:    sling.New().Base(cliBasePath),
		ccloudClient: ccloudClientProvider,
		version:      version,
	}
}

func (ld *launchDarklyManager) BoolVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal bool) bool {
	flagValInterface := ld.generalVariation(key, ctx, client, cacheFlag, defaultVal)
	flagVal, ok := flagValInterface.(bool)
	if !ok {
		logUnexpectedValueTypeMsg(key, flagValInterface, "bool")
		return defaultVal
	}
	return flagVal
}

func (ld *launchDarklyManager) StringVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal string) string {
	flagValInterface := ld.generalVariation(key, ctx, client, cacheFlag, defaultVal)
	if flagVal, ok := flagValInterface.(string); ok {
		return flagVal
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *launchDarklyManager) IntVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal int) int {
	flagValInterface := ld.generalVariation(key, ctx, client, cacheFlag, defaultVal)
	if val, ok := flagValInterface.(int); ok {
		return val
	}
	if val, ok := flagValInterface.(float64); ok { // for test since Unmarshal uses float64
		return int(val)
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *launchDarklyManager) JsonVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal interface{}) interface{} {
	flagVal := ld.generalVariation(key, ctx, client, cacheFlag, defaultVal)
	return flagVal
}

func (ld *launchDarklyManager) generalVariation(key string, ctx *dynamicconfig.DynamicContext, client v1.LaunchDarklyClient, cacheFlag bool, defaultVal interface{}) interface{} {
	user := ld.contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	var flagVals map[string]interface{}
	var err error
	if !areCachedFlagsAvailable(ctx, user, client) {
		flagVals, err = ld.fetchFlags(user, client)
		if err != nil {
			log.CliLogger.Debug(err.Error())
			return defaultVal
		}
		if cacheFlag {
			writeFlagsToConfig(ctx, key, flagVals, user, client)
		}
	} else {
		flagVals = ctx.GetLDFlags(client)
	}
	if _, ok := flagVals[key]; ok {
		return flagVals[key]
	} else {
		log.CliLogger.Debugf("unable to find value for requested flag \"%s\"", key)
		return defaultVal
	}
}

func logUnexpectedValueTypeMsg(key string, value interface{}, expectedType string) {
	log.CliLogger.Debugf(`value for flag \"%s\" was expected to be type %s but was type %T`, key, expectedType, value)
}

func (ld *launchDarklyManager) fetchFlags(user lduser.User, client v1.LaunchDarklyClient) (map[string]interface{}, error) {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return nil, fmt.Errorf("error encoding user: %w", err)
	}
	var resp *http.Response
	var flagVals map[string]interface{}
	switch client {
	case v1.CcloudProdLaunchDarklyClient, v1.CcloudStagLaunchDarklyClient, v1.CcloudDevelLaunchDarklyClient:
		resp, err = ld.ccloudClient(client).New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&flagVals, err)
	// default is "cli" client
	default:
		resp, err = ld.cliClient.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&flagVals, err)
	}
	if err != nil {
		log.CliLogger.Debug(resp)
		return flagVals, fmt.Errorf("error fetching feature flags: %w", err)
	}
	return flagVals, nil
}

func areCachedFlagsAvailable(ctx *dynamicconfig.DynamicContext, user lduser.User, client v1.LaunchDarklyClient) bool {
	if ctx == nil || ctx.Context == nil || ctx.FeatureFlags == nil {
		return false
	}

	flags := ctx.FeatureFlags

	// only use cached flags if they were fetched for the same LD User
	if !flags.User.Equal(user) {
		return false
	}

	switch client {
	case v1.CcloudDevelLaunchDarklyClient, v1.CcloudStagLaunchDarklyClient, v1.CcloudProdLaunchDarklyClient:
		if flags.CcloudValues == nil || len(flags.CcloudValues[client]) == 0 {
			return false
		}
	default:
		if len(flags.Values) == 0 {
			return false
		}
	}

	timeout := int64(time.Hour.Seconds())
	return flags.LastUpdateTime+timeout > time.Now().Unix()
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return b64.URLEncoding.EncodeToString(userBytes), nil
}

func (ld *launchDarklyManager) contextToLDUser(ctx *dynamicconfig.DynamicContext) lduser.User {
	var userBuilder lduser.UserBuilder
	custom := ldvalue.ValueMapBuild()

	if ld.version != nil && ld.version.Version != "" {
		setCustomAttribute(custom, "cli.version", ldvalue.String(ld.version.Version))
	}

	if ctx == nil || ctx.Context == nil {
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
		customValueMap := custom.Build()
		if customValueMap.Count() > 0 {
			userBuilder.CustomAll(customValueMap)
		}
		userBuilder.Key(key).Anonymous(true)
		return userBuilder.Build()
	}
	user := ctx.GetUser()
	// Basic user info
	if user != nil && user.ResourceId != "" {
		userResourceId := ctx.State.Auth.User.ResourceId
		userBuilder = lduser.NewUserBuilder(userResourceId)
		setCustomAttribute(custom, "user.resource_id", ldvalue.String(userResourceId))
	} else {
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
	}

	organization := ctx.GetOrganization()
	// org info
	if organization != nil && organization.ResourceId != "" {
		setCustomAttribute(custom, "org.resource_id", ldvalue.String(organization.ResourceId))
	}
	environment := ctx.GetEnvironment()
	// environment (account) info
	if environment != nil && environment.Id != "" {
		setCustomAttribute(custom, "environment.id", ldvalue.String(environment.Id))
	}
	// cluster info
	cluster, _ := ctx.GetKafkaClusterForCommand()
	if cluster != nil {
		if cluster.ID != "" {
			setCustomAttribute(custom, "cluster.id", ldvalue.String(cluster.ID))
		}
		if cluster.Bootstrap != "" {
			physicalClusterId := parsePkcFromBootstrap(cluster.Bootstrap)
			setCustomAttribute(custom, "cluster.physicalClusterId", ldvalue.String(physicalClusterId))
		}
	}
	customValueMap := custom.Build()
	if customValueMap.Count() > 0 {
		userBuilder.CustomAll(customValueMap)
	}
	return userBuilder.Build()
}

func setCustomAttribute(custom ldvalue.ValueMapBuilder, key string, value ldvalue.Value) {
	if !utils.Contains(attributes, key) {
		panic(fmt.Sprintf(errors.UnsupportedCustomAttributeErrorMsg, key))
	}
	custom.Set(key, value)
}

func parsePkcFromBootstrap(bootstrap string) string {
	r := regexp.MustCompile("pkc-([a-z0-9]+)")
	return r.FindString(bootstrap)
}

func writeFlagsToConfig(ctx *dynamicconfig.DynamicContext, key string, vals map[string]interface{}, user lduser.User, client v1.LaunchDarklyClient) {
	if ctx == nil {
		return
	}

	if ctx.FeatureFlags == nil {
		ctx.FeatureFlags = new(v1.FeatureFlags)
	}

	if ctx.FeatureFlags.CcloudValues == nil {
		ctx.FeatureFlags.CcloudValues = make(map[v1.LaunchDarklyClient]map[string]interface{})
	}

	switch client {
	case v1.CcloudDevelLaunchDarklyClient, v1.CcloudStagLaunchDarklyClient, v1.CcloudProdLaunchDarklyClient:
		// only store the target feature flag for ccloud clients
		if v, ok := vals[key]; ok {
			if ctx.FeatureFlags.CcloudValues[client] == nil {
				ctx.FeatureFlags.CcloudValues[client] = make(map[string]interface{})
			}
			ctx.FeatureFlags.CcloudValues[client][key] = v
		}
	default:
		ctx.FeatureFlags.Values = vals
	}

	ctx.FeatureFlags.LastUpdateTime = time.Now().Unix()
	ctx.FeatureFlags.User = user

	_ = ctx.Save()
}
