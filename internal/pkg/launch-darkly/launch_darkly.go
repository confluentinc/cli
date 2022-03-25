//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/launch_darkly.go --pkg mock --selfpkg github.com/confluentinc/cli launch_darkly.go LaunchDarklyManager

package launchdarkly

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/confluentinc/cli/internal/pkg/utils"
	test_server "github.com/confluentinc/cli/test/test-server"

	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/confluentinc/cli/internal/pkg/cmd"
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
	Manager    FeatureFlagManager // Global LD Manager
	attributes = []string{"user.resource_id", "org.resource_id", "environment.id", "cli.version", "cluster.id", "cluster.physicalClusterId"}
)

type FeatureFlagManager interface {
	BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool
	StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string
	IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int
	JsonVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{}
}

type LaunchDarklyManager struct {
	client                 *sling.Sling
	flagVals               map[string]interface{}
	flagValsAreForAnonUser bool
	version                *version.Version
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

func (ld *LaunchDarklyManager) BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	flagVal, ok := flagValInterface.(bool)
	if !ok {
		logUnexpectedValueTypeMsg(key, ld.flagVals[key], "bool")
		return defaultVal
	}
	return flagVal
}

func (ld *LaunchDarklyManager) StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	if flagVal, ok := flagValInterface.(string); ok {
		return flagVal
	}
	logUnexpectedValueTypeMsg(key, ld.flagVals[key], "int")
	return defaultVal
}

func (ld *LaunchDarklyManager) IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int {
	flagValInterface := ld.generalVariation(key, ctx, defaultVal)
	if val, ok := flagValInterface.(int); ok {
		return val
	}
	if val, ok := flagValInterface.(float64); ok { // for test since Unmarshal uses float64
		return int(val)
	}
	logUnexpectedValueTypeMsg(key, ld.flagVals[key], "int")
	return defaultVal
}

func (ld *LaunchDarklyManager) JsonVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{} {
	flagVal := ld.generalVariation(key, ctx, defaultVal)
	return flagVal
}

func (ld *LaunchDarklyManager) generalVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{} {
	user, isAnonUser := ld.contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if !ld.areCachedFlagsAvailable(isAnonUser) {
		err := ld.fetchFlags(user, isAnonUser)
		if err != nil {
			log.CliLogger.Debug(err.Error())
			return defaultVal
		}
	}
	if _, ok := ld.flagVals[key]; ok {
		return ld.flagVals[key]
	} else {
		log.CliLogger.Debugf("unable to find value for requested flag \"%s\"", key)
		return defaultVal
	}
}

func logUnexpectedValueTypeMsg(key string, value interface{}, expectedType string) {
	log.CliLogger.Debugf(`value for flag \"%s\" was expected to be type %s but was type %T`, key, expectedType, value)
}

func (ld *LaunchDarklyManager) fetchFlags(user lduser.User, isAnonUser bool) error {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return fmt.Errorf("error encoding user: %w", err)
	}
	var resp *http.Response
	resp, err = ld.client.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&ld.flagVals, err)
	if err != nil {
		log.CliLogger.Debug(resp)
		return fmt.Errorf("error fetching feature flags: %w", err)
	}
	ld.flagValsAreForAnonUser = isAnonUser
	return nil
}

func (ld *LaunchDarklyManager) areCachedFlagsAvailable(isAnonUser bool) bool {
	return len(ld.flagVals) > 0 && ld.flagValsAreForAnonUser == isAnonUser
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return b64.URLEncoding.EncodeToString(userBytes), nil
}

func (ld *LaunchDarklyManager) contextToLDUser(ctx *cmd.DynamicContext) (lduser.User, bool) {
	var userBuilder lduser.UserBuilder
	custom := ldvalue.ValueMapBuild()
	anonUser := false
	if ctx == nil || ctx.Context == nil {
		anonUser = true
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
		return userBuilder.Build(), anonUser
	}
	user := ctx.GetUser()
	// Basic user info
	if user != nil && user.ResourceId != "" {
		userResourceId := ctx.State.Auth.User.ResourceId
		userBuilder = lduser.NewUserBuilder(userResourceId)
		setCustomAttribute(custom, "user.resource_id", ldvalue.String(userResourceId))
	} else {
		anonUser = true
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
	}

	if ld.version != nil && ld.version.Version != "" {
		setCustomAttribute(custom, "cli.version", ldvalue.String(ld.version.Version))
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
	return userBuilder.Build(), anonUser
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
