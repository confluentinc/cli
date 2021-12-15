package launch_darkly

import (
	b64 "encoding/base64"
	"fmt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"os"
	"regexp"
	"strconv"
)

const (
	baseURL = "https://clientsdk.launchdarkly.com/sdk/eval/%s/"
	userPath = "users/%s"
	prodEnvClientId = "61af57740127630ce47de5be"
	testEnvClientId = "61af57740127630ce47de5bd"
)

type LaunchDarklyManager interface {
	BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool
	StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string
	IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int
}

type FeatureFlagManager struct {
	logger		*log.Logger
	client		*sling.Sling
	flagVals	map[string]interface{}
	flagValsAreForAnonUser bool
}

var LdManager LaunchDarklyManager

func InitManager(logger *log.Logger, isTest  bool) {
	// TODO if isTest, return a mock
	var basePath string
	if os.Getenv("XX_LD_TEST_ENV") != "" {
		basePath = fmt.Sprintf(baseURL, testEnvClientId)
	} else {
		basePath = fmt.Sprintf(baseURL, prodEnvClientId)
	}
	LdManager = &FeatureFlagManager{
		logger: logger,
		client: sling.New().Base(basePath),
	}
}

func (f *FeatureFlagManager) BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if f.isCacheAvailable(isAnonUser) {
		flagVal, ok := f.flagVals[key].(bool)
		if !ok {
			f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
			return defaultVal
		}
		return flagVal
	}
	err := f.fetchFlags(user, isAnonUser)
	if err != nil {
		f.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := f.flagVals[key].(bool)
	if !ok {
		f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if f.isCacheAvailable(isAnonUser) {
		flagVal, ok := f.flagVals[key].(string)
		if !ok {
			f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
			return defaultVal
		}
		return flagVal
	}
	err := f.fetchFlags(user, isAnonUser)
	if err != nil {
		f.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := f.flagVals[key].(string)
	if !ok {
		f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if f.isCacheAvailable(isAnonUser) {
		flagVal, ok := f.flagVals[key].(int)
		if !ok {
			f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
			return defaultVal
		}
		return flagVal
	}
	err := f.fetchFlags(user, isAnonUser)
	if err != nil {
		f.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := f.flagVals[key].(int)
	if !ok {
		f.logger.Debugf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", f.flagVals[key])
		return defaultVal
	}
	return flagVal
}

func (f *FeatureFlagManager) fetchFlags(user lduser.User, isAnonUser bool) error {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return fmt.Errorf("error encoding user: %w", err)
	}
	resp, err := f.client.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&f.flagVals, err)
	if err != nil {
		f.logger.Debug(resp)
		return fmt.Errorf("error fetching feature flags: %w", err)
	}
	f.flagValsAreForAnonUser = isAnonUser
	return nil
}

func (f *FeatureFlagManager) isCacheAvailable(isAnonUser bool) bool {
	return len(f.flagVals) > 0 && f.flagValsAreForAnonUser == isAnonUser
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return b64.URLEncoding.EncodeToString(userBytes), nil
}

func contextToLDUser(ctx *cmd.DynamicContext) (lduser.User, bool) {
	var userBuilder lduser.UserBuilder
	custom := ldvalue.ValueMapBuild()
	anonUser := false
	var user *orgv1.User // TODO change to internal config user struct when available
	if ctx != nil && ctx.State != nil && ctx.State.Auth != nil {
		user = ctx.State.Auth.User
	}
	// Basic user info
	if user != nil && user.Id != 0 {
		userID := strconv.Itoa(int(ctx.State.Auth.User.Id))
		userBuilder = lduser.NewUserBuilder(userID)
		custom.Set("user.id", ldvalue.String(userID))
	} else {
		anonUser = true
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
	}

	if user != nil && user.Email != "" {
		userBuilder = userBuilder.Email(user.Email).AsPrivateAttribute()
	}
	var organization *orgv1.Organization // TODO change to internal config org struct when available
	if ctx != nil && ctx.State != nil && ctx.State.Auth != nil {
		organization = ctx.State.Auth.Organization
	}
	// org info
	if organization != nil && (organization.Id != 0 || organization.ResourceId != "") {
		custom.Set("org.id", ldvalue.Int(int(organization.Id)))
		custom.Set("org.resource_id", ldvalue.String(organization.ResourceId))
		plan := organization.Plan
		if plan != nil {
			custom.Set("org.product_level", ldvalue.String(plan.ProductLevel.String()))
			if plan.Billing != nil {
				custom.Set("org.billing_method", ldvalue.String(plan.Billing.Method.String()))
			}
		}
	}
	var account *orgv1.Account // TODO change to internal config account/env struct when available
	if ctx != nil && ctx.State != nil && ctx.State.Auth != nil {
		account = ctx.State.Auth.Account
	}
	// environment (account) info
	if account != nil && account.Id != "" {
		custom.Set("environment.id", ldvalue.String(account.Id))
	}
	// cluster info
	var cluster *v1.KafkaClusterConfig
	if ctx != nil {
		cluster, _ = ctx.GetKafkaClusterForCommand()
	}
	if cluster != nil {
		if cluster.ID != "" {
			custom.Set("cluster.id", ldvalue.String(cluster.ID))
		}
		if cluster.Bootstrap != "" {
			physicalClusterId := parsePkcFromBootstrap(cluster.Bootstrap)
			custom.Set("cluster.physicalClusterId", ldvalue.String(physicalClusterId))
		}
	}
	customValueMap := custom.Build()
	if customValueMap.Count() > 0 {
		userBuilder.CustomAll(customValueMap)
	}
	return userBuilder.Build(), anonUser
}

func parsePkcFromBootstrap(bootstrap string) string {
	r := regexp.MustCompile(`(?P<Pkc>pkc-.+?(?=\.))`)
	return r.FindString(bootstrap)
}