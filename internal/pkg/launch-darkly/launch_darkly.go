package launch_darkly

import (
	b64 "encoding/base64"
	"fmt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/dghubble/sling"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"os"
	"regexp"
	"strconv"
	//ld "gopkg.in/launchdarkly/go-server-sdk.v4"
	uuid "github.com/satori/go.uuid"
)

const (
	baseURL = "https://clientsdk.launchdarkly.com/sdk/eval/%s/"
	userPath = "users/%s"
	LdProductionEnvClientId = "61af57740127630ce47de5be"
	LdTestEnvClientId = "61af57740127630ce47de5bd"
)

type LaunchDarklyManager interface {
	BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool
	StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string
	IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int
}

type LaunchDarklyManagerImpl struct {
	logger		*log.Logger
	client		*sling.Sling
	flagVals	map[string]interface{}
	flagValsAreForAnonUser bool
}

var LdManager LaunchDarklyManager

func InitLaunchDarklyManager(logger *log.Logger, isTest  bool) {
	// TODO if isTest, return a mock
	var basePath string
	if os.Getenv("XX_LD_TEST_ENV") != "" {
		basePath = fmt.Sprintf(baseURL, LdTestEnvClientId)
	} else {
		basePath = fmt.Sprintf(baseURL, LdProductionEnvClientId)
	}
	LdManager = &LaunchDarklyManagerImpl{
		logger: logger,
		client: sling.New().Base(basePath),
	}
}

func (l *LaunchDarklyManagerImpl) BoolVariation(key string, ctx *cmd.DynamicContext, defaultVal bool) bool {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if l.isCacheAvailable(isAnonUser) {
		flagVal, ok := l.flagVals[key].(bool)
		if !ok {
			l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
			return defaultVal
		}
		return flagVal
	}
	err := l.fetchLDFlags(user, isAnonUser)
	if err != nil {
		l.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := l.flagVals[key].(bool)
	if !ok {
		l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
		return defaultVal
	}
	return flagVal
}

func (l *LaunchDarklyManagerImpl) StringVariation(key string, ctx *cmd.DynamicContext, defaultVal string) string {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if l.isCacheAvailable(isAnonUser) {
		flagVal, ok := l.flagVals[key].(string)
		if !ok {
			l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
			return defaultVal
		}
		return flagVal
	}
	err := l.fetchLDFlags(user, isAnonUser)
	if err != nil {
		l.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := l.flagVals[key].(string)
	if !ok {
		l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
		return defaultVal
	}
	return flagVal
}

func (l *LaunchDarklyManagerImpl) IntVariation(key string, ctx *cmd.DynamicContext, defaultVal int) int {
	user, isAnonUser := contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	if l.isCacheAvailable(isAnonUser) {
		flagVal, ok := l.flagVals[key].(int)
		if !ok {
			l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
			return defaultVal
		}
		return flagVal
	}
	err := l.fetchLDFlags(user, isAnonUser)
	if err != nil {
		l.logger.Debug(err.Error())
		return defaultVal
	}
	flagVal, ok := l.flagVals[key].(int)
	if !ok {
		l.logger.Debug(fmt.Sprintf("value for flag \"%s\" was expected to be type %s but was type %T", key, "bool", l.flagVals[key]))
		return defaultVal
	}
	return flagVal
}

func (l *LaunchDarklyManagerImpl) fetchLDFlags(user lduser.User, isAnonUser bool) error {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return errors.Wrap(errors.New("Error encoding user: "), err.Error())
	}
	resp, err := l.client.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&l.flagVals, err)
	if err != nil {
		l.logger.Debug(resp)
		return errors.Wrap(errors.New("Error fetching feature flags: "), err.Error())
	}
	l.flagValsAreForAnonUser = isAnonUser
	return nil
}

func (l *LaunchDarklyManagerImpl) isCacheAvailable(isAnonUser bool) bool {
	return len(l.flagVals) > 0 && (l.flagValsAreForAnonUser == isAnonUser)
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
	//custom := make(map[string]ldvalue.Value, 0)
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
		key := uuid.Must(uuid.NewV4()).String()
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
	// environment info (neé account)
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