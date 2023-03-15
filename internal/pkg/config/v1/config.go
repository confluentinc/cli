package v1

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/utils"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

const (
	defaultConfigFileFmt = "%s/.confluent/config.json"
	emptyFieldIndicator  = "EMPTY"
)

var (
	ver, _ = version.NewVersion("1.0.0")
)

const signupSuggestion = `If you need a Confluent Cloud account, sign up with "confluent cloud-signup".`

var (
	RequireCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	RequireCloudLoginOrgUnsuspendedErr = errors.NewErrorWithSuggestions(
		"you must unsuspend your organization to use this command",
		errors.SuspendedOrganizationSuggestions,
	)
	RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr = errors.NewErrorWithSuggestions(
		"you must unsuspend your organization to use this command",
		errors.EndOfFreeTrialSuggestions,
	)
	RequireCloudLoginOrOnPremErr = errors.NewErrorWithSuggestions(
		"you must log in to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	RequireNonAPIKeyCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	RequireNonAPIKeyCloudLoginOrOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password or log in to Confluent Platform to use this command",
		"Log in with \"confluent login\" or \"confluent login --url <mds-url>\".\n"+signupSuggestion,
	)
	RequireNonCloudLogin = errors.NewErrorWithSuggestions(
		"you must log out of Confluent Cloud to use this command",
		"Log out with \"confluent logout\".\n",
	)
	RequireOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Platform to use this command",
		`Log in with "confluent login --url <mds-url>".`,
	)
	RequireUpdatesEnabledErr = errors.NewErrorWithSuggestions(
		"you must enable updates to use this command",
		"WARNING: To guarantee compatibility, enabling updates is not recommended for Confluent Platform users.\n"+`In ~/.confluent/config.json, set "disable_updates": false`,
	)
)

// Config represents the CLI configuration.
type Config struct {
	*config.BaseConfig

	DisableUpdateCheck  bool                        `json:"disable_update_check"`
	DisableUpdates      bool                        `json:"disable_updates"`
	DisablePlugins      bool                        `json:"disable_plugins"`
	DisablePluginsOnce  bool                        `json:"disable_plugins_once,omitempty"`
	DisableFeatureFlags bool                        `json:"disable_feature_flags"`
	NoBrowser           bool                        `json:"no_browser"`
	Platforms           map[string]*Platform        `json:"platforms,omitempty"`
	Credentials         map[string]*Credential      `json:"credentials,omitempty"`
	Contexts            map[string]*Context         `json:"contexts,omitempty"`
	ContextStates       map[string]*ContextState    `json:"context_states,omitempty"`
	CurrentContext      string                      `json:"current_context"`
	AnonymousId         string                      `json:"anonymous_id,omitempty"`
	SavedCredentials    map[string]*LoginCredential `json:"saved_credentials,omitempty"`
	LocalPorts          *LocalPorts                 `json:"local_ports,omitempty"`

	// The following configurations are not persisted between runs

	IsTest  bool              `json:"-"`
	Version *pversion.Version `json:"-"`

	overwrittenAccount     *ccloudv1.Account
	overwrittenCurrContext string
	overwrittenActiveKafka string
}

func (c *Config) SetOverwrittenAccount(acct *ccloudv1.Account) {
	if c.overwrittenAccount == nil {
		c.overwrittenAccount = acct
	}
}

func (c *Config) SetOverwrittenCurrContext(contextName string) {
	if contextName == "" {
		contextName = emptyFieldIndicator
	}
	if c.overwrittenCurrContext == "" {
		c.overwrittenCurrContext = contextName
	}
}

func (c *Config) SetOverwrittenActiveKafka(clusterId string) {
	if clusterId == "" {
		clusterId = emptyFieldIndicator
	}
	if c.overwrittenActiveKafka == "" {
		c.overwrittenActiveKafka = clusterId
	}
}

func New() *Config {
	return &Config{
		BaseConfig:       config.NewBaseConfig(ver),
		Platforms:        make(map[string]*Platform),
		Credentials:      make(map[string]*Credential),
		Contexts:         make(map[string]*Context),
		ContextStates:    make(map[string]*ContextState),
		SavedCredentials: make(map[string]*LoginCredential),
		AnonymousId:      uuid.New().String(),
		Version:          new(pversion.Version),
	}
}

// Load reads the CLI config from disk.
// Save a default version if none exists yet.
func (c *Config) Load() error {
	filename := c.GetFilename()
	input, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Save a default version if none exists yet.
			if err := c.Save(); err != nil {
				return errors.Wrapf(err, errors.UnableToCreateConfigErrorMsg)
			}
			return nil
		}
		return errors.Wrapf(err, errors.UnableToReadConfigErrorMsg, filename)
	}
	err = json.Unmarshal(input, c)
	if c.Ver.Compare(ver) < 0 {
		return errors.Errorf(errors.ConfigNotUpToDateErrorMsg, c.Ver, ver)
	} else if c.Ver.Compare(ver) > 0 {
		if c.Ver.Equal(version.Must(version.NewVersion("3.0.0"))) {
			// The user is a CP user who downloaded the v2 CLI instead of running `confluent update --major`,
			// so their config files weren't merged and migrated. Migrate this config to avoid an error.
			c.Ver = config.Version{Version: version.Must(version.NewVersion("1.0.0"))}
			for name := range c.Contexts {
				c.Contexts[name].NetrcMachineName = name
			}
		} else {
			return errors.Errorf(errors.InvalidConfigVersionErrorMsg, c.Ver)
		}
	}
	if err != nil {
		return errors.Wrapf(err, errors.ParseConfigErrorMsg, filename)
	}
	for _, context := range c.Contexts {
		// Some "pre-validation"
		if context.Name == "" {
			return errors.NewCorruptedConfigError(errors.NoNameContextErrorMsg, "", c.Filename)
		}
		if context.CredentialName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.Filename)
		}
		if context.PlatformName == "" {
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.Filename)
		}
		context.Credential = c.Credentials[context.CredentialName]
		context.Platform = c.Platforms[context.PlatformName]
		context.Config = c
		if context.KafkaClusterContext == nil {
			return errors.NewCorruptedConfigError(errors.MissingKafkaClusterContextErrorMsg, context.Name, c.Filename)
		}
		context.KafkaClusterContext.Context = context

		state := c.ContextStates[context.Name]
		if state != nil {
			err = state.DecryptContextStateAuthToken(context.Name)
			if err != nil {
				return err
			}
			err = state.DecryptContextStateAuthRefreshToken(context.Name)
			if err != nil {
				return err
			}
		}
		context.State = state
	}
	return c.Validate()
}

// Save writes the CLI config to disk.
func (c *Config) Save() error {
	tempKafka := c.resolveOverwrittenKafka()
	tempAccount := c.resolveOverwrittenAccount()
	tempContext := c.resolveOverwrittenContext()
	var tempAuthToken string
	var tempAuthRefreshToken string

	if c.Context() != nil {
		tempAuthToken = c.Context().GetState().AuthToken
		tempAuthRefreshToken = c.Context().GetState().AuthRefreshToken
		err := c.encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken)
		if err != nil {
			return err
		}
	}

	if err := c.Validate(); err != nil {
		return err
	}

	cfg, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrapf(err, errors.MarshalConfigErrorMsg)
	}

	filename := c.GetFilename()

	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return errors.Wrapf(err, errors.CreateConfigDirectoryErrorMsg, filename)
	}

	if err := os.WriteFile(filename, cfg, 0600); err != nil {
		return errors.Wrapf(err, errors.CreateConfigFileErrorMsg, filename)
	}

	c.restoreOverwrittenContext(tempContext)
	c.restoreOverwrittenAccount(tempAccount)
	c.restoreOverwrittenKafka(tempKafka)
	c.restoreOverwrittenAuthToken(tempAuthToken)
	c.restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken)

	return nil
}

func (c *Config) encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken string) error {
	if c.Context().GetState().Salt == nil || c.Context().GetState().Nonce == nil {
		salt, err := secret.GenerateRandomBytes(secret.SaltLength)
		if err != nil {
			return err
		}
		nonce, err := secret.GenerateRandomBytes(secret.NonceLength)
		if err != nil {
			return err
		}
		c.Context().GetState().Salt = salt
		c.Context().GetState().Nonce = nonce
	}
	if tempAuthToken != "" {
		encryptedAuthToken, err := secret.Encrypt(c.Context().Name, tempAuthToken, c.Context().GetState().Salt, c.Context().GetState().Nonce)
		if err != nil {
			return err
		}
		c.Context().GetState().AuthToken = encryptedAuthToken
	}
	if tempAuthRefreshToken != "" {
		encryptedAuthRefreshToken, err := secret.Encrypt(c.Context().Name, tempAuthRefreshToken, c.Context().GetState().Salt, c.Context().GetState().Nonce)
		if err != nil {
			return err
		}
		c.Context().State.AuthRefreshToken = encryptedAuthRefreshToken
	}
	return nil
}

// If active Kafka cluster has been overwritten by flag value; if so, replace with previous active kafka
// Return the flag value so that it can be restored after writing to file so that continued execution uses flag value
// This prevents flags from updating state
func (c *Config) resolveOverwrittenKafka() string {
	ctx := c.Context()
	var tempKafka string
	if c.overwrittenActiveKafka != "" && ctx != nil && ctx.KafkaClusterContext != nil {
		if c.overwrittenActiveKafka == emptyFieldIndicator {
			c.overwrittenActiveKafka = ""
		}
		tempKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		ctx.KafkaClusterContext.SetActiveKafkaCluster(c.overwrittenActiveKafka)
	}
	return tempKafka
}

// Restore the flag cluster back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenKafka(tempKafka string) {
	ctx := c.Context()
	if tempKafka != "" {
		ctx.KafkaClusterContext.SetActiveKafkaCluster(tempKafka)
	}
}

func (c *Config) restoreOverwrittenAuthToken(tempAuthToken string) {
	ctx := c.Context()
	if tempAuthToken != "" {
		ctx.GetState().AuthToken = tempAuthToken
	}
}

func (c *Config) restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken string) {
	ctx := c.Context()
	if tempAuthRefreshToken != "" {
		ctx.GetState().AuthRefreshToken = tempAuthRefreshToken
	}
}

// Switch the initial config context back into the struct so that it is saved and not the flag value
// Return the overwriting flag context value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenContext() string {
	var tempContext string
	if c.overwrittenCurrContext != "" && c != nil {
		if c.overwrittenCurrContext == emptyFieldIndicator {
			c.overwrittenCurrContext = ""
		}
		tempContext = c.CurrentContext
		c.CurrentContext = c.overwrittenCurrContext
	}
	return tempContext
}

// Restore the flag context back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenContext(tempContext string) {
	if tempContext != "" {
		c.CurrentContext = tempContext
	}
}

// Switch the initial config account back into the struct so that it is saved and not the flag value
// Return the overwriting flag account value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenAccount() *ccloudv1.Account {
	ctx := c.Context()
	var tempAccount *ccloudv1.Account
	if c.overwrittenAccount != nil && ctx != nil && ctx.State != nil && ctx.State.Auth != nil {
		tempAccount = ctx.GetEnvironment()
		ctx.State.Auth.Account = c.overwrittenAccount
	}
	return tempAccount
}

// Restore the flag account back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenAccount(tempAccount *ccloudv1.Account) {
	ctx := c.Context()
	if tempAccount != nil {
		ctx.State.Auth.Account = tempAccount
	}
}

func (c *Config) Validate() error {
	// Validate that current context exists.
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			log.CliLogger.Trace("current context does not exist")
			return errors.NewCorruptedConfigError(errors.CurrentContextNotExistErrorMsg, c.CurrentContext, c.Filename)
		}
	}
	// Validate that every context:
	// 1. Has no hanging references between the context and the config.
	// 2. Is mapped by name correctly in the config.
	for _, context := range c.Contexts {
		err := context.validate()
		if err != nil {
			log.CliLogger.Trace("context validation error")
			return err
		}
		if _, ok := c.Credentials[context.CredentialName]; !ok {
			log.CliLogger.Trace("unspecified credential error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, context.Name, c.Filename)
		}
		if _, ok := c.Platforms[context.PlatformName]; !ok {
			log.CliLogger.Trace("unspecified platform error")
			return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, context.Name, c.Filename)
		}
		if _, ok := c.ContextStates[context.Name]; !ok {
			c.ContextStates[context.Name] = new(ContextState)
		}
		if !c.IsTest && !reflect.DeepEqual(*c.ContextStates[context.Name], *context.State) {
			log.CliLogger.Tracef("state of context %s in config does not match actual state of context", context.Name)
			return errors.NewCorruptedConfigError(errors.ContextStateMismatchErrorMsg, context.Name, c.Filename)
		}
	}
	// Validate that all context states are mapped to an existing context.
	for contextName := range c.ContextStates {
		if _, ok := c.Contexts[contextName]; !ok {
			log.CliLogger.Trace("context state mapped to nonexistent context")
			return errors.NewCorruptedConfigError(errors.ContextStateNotMappedErrorMsg, contextName, c.Filename)
		}
	}

	return nil
}

// DeleteContext deletes the specified context, and returns an error if it's not found.
func (c *Config) DeleteContext(name string) error {
	if _, err := c.FindContext(name); err != nil {
		return err
	}
	delete(c.Contexts, name)
	delete(c.ContextStates, name)

	if name == c.CurrentContext {
		c.CurrentContext = ""
	}

	return c.Save()
}

// FindContext finds a context by name, and returns nil if not found.
func (c *Config) FindContext(name string) (*Context, error) {
	context, ok := c.Contexts[name]
	if !ok {
		return nil, fmt.Errorf(errors.ContextDoesNotExistErrorMsg, name)
	}
	return context, nil
}

func (c *Config) AddContext(name, platformName, credentialName string, kafkaClusters map[string]*KafkaClusterConfig, kafka string, schemaRegistryClusters map[string]*SchemaRegistryCluster, state *ContextState, orgResourceId string) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf(errors.ContextAlreadyExistsErrorMsg, name)
	}

	credential, ok := c.Credentials[credentialName]
	if !ok {
		return fmt.Errorf(errors.CredentialNotFoundErrorMsg, credentialName)
	}

	platform, ok := c.Platforms[platformName]
	if !ok {
		return fmt.Errorf(errors.PlatformNotFoundErrorMsg, platformName)
	}

	ctx, err := newContext(name, platform, credential, kafkaClusters, kafka, schemaRegistryClusters, state, c, orgResourceId)
	if err != nil {
		return err
	}

	c.Contexts[name] = ctx
	c.ContextStates[name] = ctx.State

	if err := c.Validate(); err != nil {
		return err
	}

	return c.Save()
}

// CreateContext creates a new context.
func (c *Config) CreateContext(name, bootstrapURL, apiKey, apiSecret string) error {
	apiKeyPair := &APIKeyPair{
		Key:    apiKey,
		Secret: apiSecret,
	}
	apiKeys := map[string]*APIKeyPair{
		apiKey: apiKeyPair,
	}
	kafkaClusterCfg := &KafkaClusterConfig{
		ID:        "anonymous-id",
		Name:      "anonymous-cluster",
		Bootstrap: bootstrapURL,
		APIKeys:   apiKeys,
		APIKey:    apiKey,
	}
	kafkaClusters := map[string]*KafkaClusterConfig{
		kafkaClusterCfg.ID: kafkaClusterCfg,
	}
	platform := &Platform{Server: bootstrapURL}

	// Inject credential and platforms name for now, until users can provide custom names.
	platform.Name = strings.TrimPrefix(platform.Server, "https://")

	// Hardcoded for now, since username/password isn't implemented yet.
	credential := &Credential{
		Username:       "",
		Password:       "",
		APIKeyPair:     apiKeyPair,
		CredentialType: APIKey,
	}

	switch credential.CredentialType {
	case Username:
		credential.Name = fmt.Sprintf("%s-%s", &credential.CredentialType, credential.Username)
	case APIKey:
		credential.Name = fmt.Sprintf("%s-%s", &credential.CredentialType, credential.APIKeyPair.Key)
	default:
		return errors.Errorf(errors.UnknownCredentialTypeErrorMsg, credential.CredentialType)
	}

	if err := c.SaveCredential(credential); err != nil {
		return err
	}

	if err := c.SavePlatform(platform); err != nil {
		return err
	}

	return c.AddContext(name, platform.Name, credential.Name, kafkaClusters, kafkaClusterCfg.ID, nil, nil, "")
}

// UseContext sets the current context, if it exists.
func (c *Config) UseContext(name string) error {
	if _, err := c.FindContext(name); err != nil {
		return err
	}
	c.CurrentContext = name
	return c.Save()
}

func (c *Config) SaveCredential(credential *Credential) error {
	if credential.Name == "" {
		return errors.New(errors.NoNameCredentialErrorMsg)
	}
	c.Credentials[credential.Name] = credential
	return c.Save()
}

func (c *Config) SaveLoginCredential(ctxName string, loginCredential *LoginCredential) error {
	if ctxName == "" {
		return errors.New(errors.SavedCredentialNoContextErrorMsg)
	}
	c.SavedCredentials[ctxName] = loginCredential
	return c.Save()
}

func (c *Config) SavePlatform(platform *Platform) error {
	if platform.Name == "" {
		return errors.New(errors.NoNamePlatformErrorMsg)
	}
	c.Platforms[platform.Name] = platform
	return c.Save()
}

// Context returns the current context.
func (c *Config) Context() *Context {
	if c == nil {
		return nil
	}
	return c.Contexts[c.CurrentContext]
}

// CredentialType returns the credential type used in the current context: API key, username & password, or neither.
func (c *Config) CredentialType() CredentialType {
	if c.hasAPIKeyLogin() {
		return APIKey
	}

	if c.HasBasicLogin() {
		return Username
	}

	return None
}

// hasAPIKeyLogin returns true if the user has valid API Key credentials.
func (c *Config) hasAPIKeyLogin() bool {
	ctx := c.Context()
	return ctx != nil && ctx.Credential != nil && ctx.Credential.CredentialType == APIKey
}

// HasBasicLogin returns true if the user has valid username & password credentials.
func (c *Config) HasBasicLogin() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	if c.IsCloudLogin() {
		return ctx.hasBasicCloudLogin()
	} else {
		return ctx.HasBasicMDSLogin()
	}
}

func (c *Config) ResetAnonymousId() error {
	c.AnonymousId = uuid.New().String()
	return c.Save()
}

func (c *Config) GetFilename() string {
	if c.Filename == "" {
		homedir, _ := os.UserHomeDir()
		c.Filename = filepath.FromSlash(fmt.Sprintf(defaultConfigFileFmt, homedir))
	}
	return c.Filename
}

func (c *Config) CheckIsOnPremLogin() error {
	ctx := c.Context()
	if ctx != nil && ctx.PlatformName != "" && !c.isCloud() {
		return nil
	}
	return RequireOnPremLoginErr
}

func (c *Config) CheckIsCloudLogin() error {
	if !c.isCloud() {
		return RequireCloudLoginErr
	}

	if c.isContextStatePresent() && c.isOrgSuspended() {
		if c.isLoginBlockedByOrgSuspension() {
			return RequireCloudLoginOrgUnsuspendedErr
		} else {
			return RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr
		}
	}

	return nil
}

func (c *Config) CheckIsCloudLoginAllowFreeTrialEnded() error {
	if !c.isCloud() {
		return RequireCloudLoginErr
	}

	if c.isContextStatePresent() && c.isLoginBlockedByOrgSuspension() {
		return RequireCloudLoginOrgUnsuspendedErr
	}

	return nil
}

func (c *Config) CheckIsCloudLoginOrOnPremLogin() error {
	isCloudLoginErr := c.CheckIsCloudLogin()
	isOnPremLoginErr := c.CheckIsOnPremLogin()

	if !(isCloudLoginErr == nil || isOnPremLoginErr == nil) {
		// return org suspension errors
		if isCloudLoginErr != nil && isCloudLoginErr != RequireCloudLoginErr {
			return isCloudLoginErr
		}
		return RequireCloudLoginOrOnPremErr
	}

	return nil
}

func (c *Config) CheckIsNonAPIKeyCloudLogin() error {
	isCloudLoginErr := c.CheckIsCloudLogin()

	if !(c.CredentialType() != APIKey && isCloudLoginErr == nil) {
		// return org suspension errors
		if isCloudLoginErr != nil && isCloudLoginErr != RequireCloudLoginErr {
			return isCloudLoginErr
		}
		return RequireNonAPIKeyCloudLoginErr
	}

	return nil
}

func (c *Config) CheckIsNonAPIKeyCloudLoginOrOnPremLogin() error {
	isNonAPIKeyCloudLoginErr := c.CheckIsNonAPIKeyCloudLogin()
	isOnPremLoginErr := c.CheckIsOnPremLogin()

	if !(isNonAPIKeyCloudLoginErr == nil || isOnPremLoginErr == nil) {
		// return org suspension errors
		if isNonAPIKeyCloudLoginErr != nil && isNonAPIKeyCloudLoginErr != RequireCloudLoginErr && isNonAPIKeyCloudLoginErr != RequireNonAPIKeyCloudLoginErr {
			return isNonAPIKeyCloudLoginErr
		}
		return RequireNonAPIKeyCloudLoginOrOnPremLoginErr
	}

	return nil
}

func (c *Config) CheckIsNonCloudLogin() error {
	if c.isCloud() {
		return RequireNonCloudLogin
	}
	return nil
}

func (c *Config) CheckAreUpdatesEnabled() error {
	if c.DisableUpdates {
		return RequireUpdatesEnabledErr
	}
	return nil
}

func (c *Config) IsCloudLogin() bool {
	return c.CheckIsCloudLogin() == nil
}

func (c *Config) IsOnPremLogin() bool {
	return c.CheckIsOnPremLogin() == nil
}

func (c *Config) isCloud() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	return ctx.IsCloud(c.IsTest)
}

func (c *Config) isContextStatePresent() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	if ctx.GetOrganization() == nil {
		log.CliLogger.Trace("current context state is not set up properly for checking org suspension status")
		return false
	}

	return true
}

func (c *Config) isOrgSuspended() bool {
	return utils.IsOrgSuspended(c.Context().GetSuspensionStatus())
}

func (c *Config) isLoginBlockedByOrgSuspension() bool {
	return utils.IsLoginBlockedByOrgSuspension(c.Context().GetSuspensionStatus())
}

func (c *Config) GetLastUsedOrgId() string {
	if ctx := c.Context(); ctx != nil && ctx.LastOrgId != "" {
		return ctx.LastOrgId
	}
	return os.Getenv("CONFLUENT_CLOUD_ORGANIZATION_ID")
}

func (c *Config) GetCloudClientV2(unsafeTrace bool) *ccloudv2.Client {
	ctx := c.Context()
	return ccloudv2.NewClient(ctx.GetPlatformServer(), c.IsTest, ctx.GetAuthToken(), c.Version.UserAgent, unsafeTrace)
}
