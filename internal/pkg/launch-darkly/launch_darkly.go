//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/launch_darkly.go --pkg mock --selfpkg github.com/confluentinc/cli launch_darkly.go LaunchDarklyManager

package launch_darkly

import (
	b64 "encoding/base64"
	"fmt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	//"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/version"
	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"os"
	"regexp"
)

const (
	baseURL = "https://confluent.cloud/ldapi/sdk/eval/%s/"
	userPath = "users/%s"
	prodEnvClientId = "61af57740127630ce47de5be"
	testEnvClientId = "61af57740127630ce47de5bd"
)

var (
	LdManager LaunchDarklyManager // Global LdManager
	attributes = []string{"user.resource_id", "org.resource_id", "environment.id", "cli.version", "cluster.id" , "cluster.physicalClusterId"}
)


type LaunchDarklyManager interface {
	BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool
	StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string
	IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int
	JsonVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{}
}

type FeatureFlagManager struct {
	client		*sling.Sling
	flagVals	map[string]interface{}
	flagValsAreForAnonUser bool
	version		*version.Version
}

func InitManager(version *version.Version, isTest  bool) {
	if isTest {
	//	LdManager = &mock.LaunchDarklyManager{}
	}
	var basePath string
	if os.Getenv("XX_LD_TEST_ENV") != "" {
		basePath = fmt.Sprintf(baseURL, testEnvClientId)
	} else {
		basePath = fmt.Sprintf(baseURL, prodEnvClientId)
	}
	LdManager = &FeatureFlagManager{
		client: 	sling.New().Base(basePath),
		version: 	version,
	}
}

func (f *FeatureFlagManager) BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool {
	flagValInterface := f.generalVariation(key, ctx, defaultVal)
	flagVal, ok := flagValInterface.(bool)
	if !ok {
		log.CliLogger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string {
	flagValInterface := f.generalVariation(key, ctx, defaultVal)
	flagVal, ok := flagValInterface.(string)
	if !ok {
		log.CliLogger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "string", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int {
	flagValInterface := f.generalVariation(key, ctx, defaultVal)
	flagVal, ok := flagValInterface.(int)
	if !ok {
		log.CliLogger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "int", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) JsonVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{} {
	flagVal:= f.generalVariation(key, ctx, defaultVal)
	return flagVal
}

func (f *FeatureFlagManager) generalVariation(key string, ctx *cmd.DynamicContext, defaultVal interface{}) interface{} {
	user, isAnonUser := f.contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if f.areCachedFlagsAvailable(isAnonUser) {
		return f.flagVals[key]
	}
	err := f.fetchFlags(user, isAnonUser)
	if err != nil {
		log.CliLogger.Debug(err.Error())
		return defaultVal
	}
	return f.flagVals[key]
}

func (f *FeatureFlagManager) fetchFlags(user lduser.User, isAnonUser bool) error {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return fmt.Errorf("error encoding user: %w", err)
	}
	resp, err := f.client.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&f.flagVals, err)
	if err != nil {
		log.CliLogger.Debug(resp)
		return fmt.Errorf("error fetching feature flags: %w", err)
	}
	f.flagValsAreForAnonUser = isAnonUser
	return nil
}

func (f *FeatureFlagManager) areCachedFlagsAvailable(isAnonUser bool) bool {
	return len(f.flagVals) > 0 && f.flagValsAreForAnonUser == isAnonUser
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return b64.URLEncoding.EncodeToString(userBytes), nil
}

func (f *FeatureFlagManager) contextToLDUser(ctx *cmd.DynamicContext) (lduser.User, bool) {
	var userBuilder lduser.UserBuilder
	custom := ldvalue.ValueMapBuild()
	anonUser := false
	if ctx == nil || ctx.Context == nil {
		anonUser = true
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
		return userBuilder.Build(), anonUser
	}
	var user *orgv1.User
	if ctx.State != nil && ctx.State.Auth != nil {
		user = ctx.State.Auth.User
	}
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

	if f.version != nil && f.version.Version != "" {
		setCustomAttribute(custom, "cli.version", ldvalue.String(f.version.Version))
	}

	var organization *orgv1.Organization
	if ctx.State != nil && ctx.State.Auth != nil {
		organization = ctx.State.Auth.Organization
	}
	// org info
	if organization != nil && organization.ResourceId != "" {
		setCustomAttribute(custom, "org.resource_id", ldvalue.String(organization.ResourceId))
	}
	var account *orgv1.Account
	if ctx.State != nil && ctx.State.Auth != nil {
		account = ctx.State.Auth.Account
	}
	// environment (account) info
	if account != nil && account.Id != "" {
		setCustomAttribute(custom, "environment.id", ldvalue.String(account.Id))
	}
	// cluster info
	var cluster *v1.KafkaClusterConfig
	if ctx != nil {
		cluster, _ = ctx.GetKafkaClusterForCommand()
	}
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

func setCustomAttribute(custom ldvalue.ValueMapBuilder, key string, value ldvalue.Value) error {
	found := false
	for _, attribute := range attributes {
		if key == attribute {
			found = true
			break // key is an accepted targeting attribute
		}
	}
	if !found {
		panic(fmt.Sprintf(errors.UnsupportedCustomAttributeErrorMsg, key))
	}
	custom.Set(key, value)
	return nil
}

func parsePkcFromBootstrap(bootstrap string) string {
	r := regexp.MustCompile("pkc-([^.]+)")
	return r.FindString(bootstrap)
}