//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/launch_darkly.go --pkg mock --selfpkg github.com/confluentinc/cli launch_darkly.go LaunchDarklyManager

package launchdarkly

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"

	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"

	"github.com/confluentinc/cli/internal/pkg/utils"
	test_server "github.com/confluentinc/cli/test/test-server"

	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

const (
	baseURL         = "https://confluent.cloud/ldapi/sdk/eval/%s/"
	userPath        = "users/%s"
	prodEnvClientId = "61af57740127630ce47de5be"
	testEnvClientId = "61af57740127630ce47de5bd"
)

var (
	Manager    featureFlagManager // Global LD Manager
	attributes = []string{"user.resource_id", "org.resource_id", "environment.id", "cli.version", "cluster.id", "cluster.physicalClusterId"}
)

type featureFlagManager interface {
	BoolVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal bool) bool
	StringVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal string) string
	IntVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal int) int
	JsonVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal interface{}) interface{}
}

type LaunchDarklyManager struct {
	client  *sling.Sling
	version *version.Version
}

func InitManager(version *version.Version, isTest bool) {
	basePath := fmt.Sprintf(baseURL, prodEnvClientId)
	if isTest {
		basePath = test_server.TestCloudURL.Path + "/ldapi/sdk/eval/1234"
	} else if os.Getenv("XX_LAUNCH_DARKLY_TEST_ENV") != "" {
		basePath = fmt.Sprintf(baseURL, testEnvClientId)
	}
	Manager = &LaunchDarklyManager{
		client:  sling.New().Base(basePath),
		version: version,
	}
}

func (ld *LaunchDarklyManager) BoolVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal bool) bool {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	flagVal, ok := flagValInterface.(bool)
	if !ok {
		logUnexpectedValueTypeMsg(key, flagValInterface, "bool")
		return defaultVal
	}
	return flagVal
}

func (ld *LaunchDarklyManager) StringVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal string) string {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	if flagVal, ok := flagValInterface.(string); ok {
		return flagVal
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *LaunchDarklyManager) IntVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal int) int {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	if val, ok := flagValInterface.(int); ok {
		return val
	}
	if val, ok := flagValInterface.(float64); ok { // for test since Unmarshal uses float64
		return int(val)
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *LaunchDarklyManager) JsonVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal interface{}) interface{} {
	flagVal := ld.generalVariation(key, ctx, defaultVal)
	return flagVal
}

func (ld *LaunchDarklyManager) generalVariation(key string, ctx *dynamicconfig.DynamicContext, defaultVal interface{}) interface{} {
	user := ld.contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	var flagVals map[string]interface{}
	var err error
	if !areCachedFlagsAvailable(ctx, user) {
		flagVals, err = ld.fetchFlags(user)
		if err != nil {
			log.CliLogger.Debug(err.Error())
			return defaultVal
		}
		writeFlagsToConfig(ctx, flagVals, user)
	} else {
		flagVals = ctx.GetLDFlags()
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

func (ld *LaunchDarklyManager) fetchFlags(user lduser.User) (map[string]interface{}, error) {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return nil, fmt.Errorf("error encoding user: %w", err)
	}
	var resp *http.Response
	var flagVals map[string]interface{}
	resp, err = ld.client.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&flagVals, err)
	if err != nil {
		log.CliLogger.Debug(resp)
		return flagVals, fmt.Errorf("error fetching feature flags: %w", err)
	}
	return flagVals, nil
}

func areCachedFlagsAvailable(ctx *dynamicconfig.DynamicContext, user lduser.User) bool {
	if ctx == nil || ctx.Context == nil || ctx.LDConfig == nil {
		return false
	}
	// only use cached flags if they were fetched for the same LD User
	if !ctx.LDConfig.User.Equal(user) {
		return false
	}

	flagExpirationTime := int64(time.Hour.Seconds())

	isExpired := ctx.LDConfig.FlagUpdateTime < time.Now().Unix()-flagExpirationTime
	return !isExpired && len(ctx.LDConfig.FlagValues) > 0
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return b64.URLEncoding.EncodeToString(userBytes), nil
}

func (ld *LaunchDarklyManager) contextToLDUser(ctx *dynamicconfig.DynamicContext) lduser.User {
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

func writeFlagsToConfig(ctx *dynamicconfig.DynamicContext, vals map[string]interface{}, user lduser.User) {
	if ctx == nil {
		return
	} else if ctx.LDConfig == nil {
		ctx.LDConfig = &v1.LaunchDarkly{}
	}
	ctx.LDConfig.FlagValues = vals
	ctx.LDConfig.FlagUpdateTime = time.Now().Unix()
	ctx.LDConfig.User = user
	_ = ctx.Save()
}
