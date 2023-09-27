//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/launch_darkly.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 launch_darkly.go LaunchDarklyManager

package featureflags

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/dghubble/sling"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	ppanic "github.com/confluentinc/cli/v3/pkg/panic-recovery"
	"github.com/confluentinc/cli/v3/pkg/version"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

const (
	baseURL  = "%s/ldapi/sdk/eval/%s/"
	userPath = "users/%s"
)

const (
	devel = "devel.cpdev.cloud"
	stag  = "stag.cpdev.cloud"
)

const (
	cliProdEnvClientId     = "61af57740127630ce47de5be"
	cliTestEnvClientId     = "61af57740127630ce47de5bd"
	ccloudProdEnvClientId  = "5c636508aa445d32c86f26b1"
	ccloudStagEnvClientId  = "5c63651f1df21432a45fc773"
	ccloudDevelEnvClientId = "5c63653912b6db32db950445"
)

// Manager is a global feature flag manager
var Manager launchDarklyManager

var attributes = []string{"user.resource_id", "org.resource_id", "environment.id", "cli.version", "cluster.id", "cluster.physicalClusterId", "cli.command", "cli.flags"}

type launchDarklyManager struct {
	cliClient             *sling.Sling
	ccloudClient          *sling.Sling
	Command               *cobra.Command
	flags                 []string
	hideTimeoutWarning    bool
	isDisabled            bool
	timeoutWarningPrinted bool
	version               *version.Version
}

func Init(cfg *config.Config) {
	cliBasePath := fmt.Sprintf(baseURL, auth.CCloudURL, cliTestEnvClientId)
	ccloudBasePath := fmt.Sprintf(baseURL, auth.CCloudURL, ccloudProdEnvClientId)
	if cfg.IsTest {
		cliBasePath = fmt.Sprintf(baseURL, testserver.TestCloudUrl.String(), "1234")
		ccloudBasePath = cliBasePath
	} else {
		switch cfg.Context().GetPlatform().GetName() {
		case devel:
			ccloudBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, ccloudDevelEnvClientId)
		case stag:
			ccloudBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, ccloudStagEnvClientId)
		default:
			cliBasePath = fmt.Sprintf(baseURL, auth.CCloudURL, cliProdEnvClientId)
		}
	}

	Manager = launchDarklyManager{
		cliClient:             sling.New().Base(cliBasePath),
		ccloudClient:          sling.New().Base(ccloudBasePath),
		hideTimeoutWarning:    cfg.IsTest,
		isDisabled:            cfg.DisableFeatureFlags,
		timeoutWarningPrinted: false,
		version:               cfg.Version,
	}
}

// GetCcloudLaunchDarklyClient resolves to a LaunchDarkly client based on the string platform name that is passed in.
func GetCcloudLaunchDarklyClient(platformName string) config.LaunchDarklyClient {
	switch platformName {
	case "stag.cpdev.cloud":
		return config.CcloudStagLaunchDarklyClient
	case "devel.cpdev.cloud":
		return config.CcloudDevelLaunchDarklyClient
	default:
		return config.CcloudProdLaunchDarklyClient
	}
}

func (ld *launchDarklyManager) SetCommandAndFlags(cmd *cobra.Command, args []string) {
	fullCmd, flagsAndArgs, _ := cmd.Find(args)
	flags := ppanic.ParseFlags(fullCmd, flagsAndArgs)
	ld.Command = fullCmd
	ld.flags = flags
}

func (ld *launchDarklyManager) BoolVariation(key string, ctx *dynamicconfig.DynamicContext, client config.LaunchDarklyClient, shouldCache, defaultVal bool) bool {
	flagValInterface := ld.generalVariation(key, ctx, client, shouldCache, defaultVal)
	flagVal, ok := flagValInterface.(bool)
	if !ok {
		logUnexpectedValueTypeMsg(key, flagValInterface, "bool")
		return defaultVal
	}
	return flagVal
}

func (ld *launchDarklyManager) StringVariation(key string, ctx *dynamicconfig.DynamicContext, client config.LaunchDarklyClient, shouldCache bool, defaultVal string) string {
	flagValInterface := ld.generalVariation(key, ctx, client, shouldCache, defaultVal)
	if flagVal, ok := flagValInterface.(string); ok {
		return flagVal
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *launchDarklyManager) IntVariation(key string, ctx *dynamicconfig.DynamicContext, client config.LaunchDarklyClient, shouldCache bool, defaultVal int) int {
	flagValInterface := ld.generalVariation(key, ctx, client, shouldCache, defaultVal)
	if val, ok := flagValInterface.(int); ok {
		return val
	}
	if val, ok := flagValInterface.(float64); ok { // for test since Unmarshal uses float64
		return int(val)
	}
	logUnexpectedValueTypeMsg(key, flagValInterface, "int")
	return defaultVal
}

func (ld *launchDarklyManager) JsonVariation(key string, ctx *dynamicconfig.DynamicContext, client config.LaunchDarklyClient, shouldCache bool, defaultVal any) any {
	return ld.generalVariation(key, ctx, client, shouldCache, defaultVal)
}

func (ld *launchDarklyManager) generalVariation(key string, ctx *dynamicconfig.DynamicContext, client config.LaunchDarklyClient, shouldCache bool, defaultVal any) any {
	if ld.isDisabled {
		return defaultVal
	}

	user := ld.contextToLDUser(ctx)
	// Check if cached flags are available
	// Check if cached flags are for same auth status (anon or not anon) as current ctx so that we know the values are valid based on targeting
	var flagVals map[string]any
	var err error
	if !areCachedFlagsAvailable(ctx, user, client, key) {
		flagVals, err = ld.fetchFlags(user, client)
		if err != nil {
			log.CliLogger.Debug(err.Error())
			return defaultVal
		}
		if shouldCache {
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

func logUnexpectedValueTypeMsg(key string, value any, expectedType string) {
	log.CliLogger.Debugf(`value for flag \"%s\" was expected to be type %s but was type %T`, key, expectedType, value)
}

func (ld *launchDarklyManager) fetchFlags(user lduser.User, client config.LaunchDarklyClient) (map[string]any, error) {
	userEnc, err := getBase64EncodedUser(user)
	if err != nil {
		return nil, fmt.Errorf("error encoding user: %w", err)
	}
	var resp *http.Response
	var flagVals map[string]any
	switch client {
	case config.CcloudProdLaunchDarklyClient, config.CcloudStagLaunchDarklyClient, config.CcloudDevelLaunchDarklyClient:
		resp, err = ld.ccloudClient.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&flagVals, err)
	// default is "cli" client
	default:
		resp, err = ld.cliClient.New().Get(fmt.Sprintf(userPath, userEnc)).Receive(&flagVals, err)
	}
	if err != nil {
		log.CliLogger.Debug(resp)
		if !ld.hideTimeoutWarning && !ld.timeoutWarningPrinted {
			output.ErrPrintln(false, "WARNING: Failed to fetch feature flags.")
			output.ErrPrintln(false, errors.ComposeSuggestionsMessage("Check connectivity to https://confluent.cloud or disable feature flags using `confluent configuration update disable_feature_flags true`."))
			ld.timeoutWarningPrinted = true
		}

		return flagVals, fmt.Errorf("error fetching feature flags: %w", err)
	}
	return flagVals, nil
}

func areCachedFlagsAvailable(ctx *dynamicconfig.DynamicContext, user lduser.User, client config.LaunchDarklyClient, key string) bool {
	if ctx == nil || ctx.Context == nil || ctx.FeatureFlags == nil {
		return false
	}

	flags := ctx.FeatureFlags

	// only use cached flags if they were fetched for the same LD User
	if !flags.User.Equal(user) {
		return false
	}

	switch client {
	case config.CcloudDevelLaunchDarklyClient, config.CcloudStagLaunchDarklyClient, config.CcloudProdLaunchDarklyClient:
		// for ccloud only single keys are stored, so we need to check if the specific flag is present
		if _, ok := flags.CcloudValues[key]; !ok {
			return false
		}
	default:
		if _, ok := flags.CliValues[key]; !ok {
			return false
		}
	}

	return cacheExpired(flags)
}

func cacheExpired(flags *config.FeatureFlags) bool {
	return flags.LastUpdateTime+int64(time.Hour.Seconds()) > time.Now().Unix()
}

func getBase64EncodedUser(user lduser.User) (string, error) {
	userBytes, err := user.MarshalJSON()
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(userBytes), nil
}

func (ld *launchDarklyManager) contextToLDUser(ctx *dynamicconfig.DynamicContext) lduser.User {
	var userBuilder lduser.UserBuilder
	custom := ldvalue.ValueMapBuild()

	if ld.version != nil && ld.version.Version != "" {
		setCustomAttribute(custom, "cli.version", ldvalue.String(ld.version.Version))
	}

	if ld.Command != nil {
		setCustomAttribute(custom, "cli.command", ldvalue.String(ld.Command.CommandPath()))
	}

	if ld.flags != nil {
		setCustomAttribute(custom, "cli.flags", ldvalue.CopyArbitraryValue(ld.flags))
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
	// Basic user info
	if id := ctx.GetUser().GetResourceId(); id != "" {
		userBuilder = lduser.NewUserBuilder(id)
		setCustomAttribute(custom, "user.resource_id", ldvalue.String(id))
	} else {
		key := uuid.New().String()
		userBuilder = lduser.NewUserBuilder(key).Anonymous(true)
	}
	// org info
	if id := ctx.GetCurrentOrganization(); id != "" {
		setCustomAttribute(custom, "org.resource_id", ldvalue.String(id))
	}
	// environment (account) info
	if id := ctx.GetCurrentEnvironment(); id != "" {
		setCustomAttribute(custom, "environment.id", ldvalue.String(id))
	}
	customValueMap := custom.Build()
	if customValueMap.Count() > 0 {
		userBuilder.CustomAll(customValueMap)
	}
	return userBuilder.Build()
}

func setCustomAttribute(custom ldvalue.ValueMapBuilder, key string, value ldvalue.Value) {
	if !slices.Contains(attributes, key) {
		panic(fmt.Sprintf(errors.UnsupportedCustomAttributeErrorMsg, key))
	}
	custom.Set(key, value)
}

func writeFlagsToConfig(ctx *dynamicconfig.DynamicContext, key string, vals map[string]any, user lduser.User, client config.LaunchDarklyClient) {
	if ctx == nil {
		return
	}

	if ctx.FeatureFlags == nil {
		ctx.FeatureFlags = new(config.FeatureFlags)
	}

	switch client {
	case config.CcloudDevelLaunchDarklyClient, config.CcloudStagLaunchDarklyClient, config.CcloudProdLaunchDarklyClient:
		// only store the target feature flag for ccloud clients
		if ctx.FeatureFlags.CcloudValues == nil {
			ctx.FeatureFlags.CcloudValues = make(map[string]any)
		}
		if v, ok := vals[key]; ok {
			ctx.FeatureFlags.CcloudValues[key] = v
		}
	default:
		ctx.FeatureFlags.CliValues = vals
	}

	ctx.FeatureFlags.LastUpdateTime = time.Now().Unix()
	ctx.FeatureFlags.User = user

	_ = ctx.Save()
}
