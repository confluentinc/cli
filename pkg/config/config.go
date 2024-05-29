package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/secret"
	"github.com/confluentinc/cli/v3/pkg/utils"
	pversion "github.com/confluentinc/cli/v3/pkg/version"
)

const emptyFieldIndicator = "EMPTY"

const signupSuggestion = "If you need a Confluent Cloud account, sign up with `confluent cloud-signup`."

var (
	RequireCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud to use this command",
		"Log in with `confluent login`.\n"+signupSuggestion,
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
		"Log in with `confluent login`.\n"+signupSuggestion,
	)
	RequireNonAPIKeyCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password to use this command",
		"Log in with `confluent login`.\n"+signupSuggestion,
	)
	RequireNonAPIKeyCloudLoginOrOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password or log in to Confluent Platform to use this command",
		"Log in with `confluent login` or `confluent login --url <mds-url>`.\n"+signupSuggestion,
	)
	RequireNonCloudLogin = errors.NewErrorWithSuggestions(
		"you must log out of Confluent Cloud to use this command",
		"Log out with `confluent logout`.\n",
	)
	RequireOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Platform to use this command",
		"Log in to Confluent Platform with `confluent login --url <mds-url>`.",
	)
	RunningOnPremCommandInCloudErr = errors.NewErrorWithSuggestions(
		"this is not a Confluent Cloud command. You must log in to Confluent Platform to use this command",
		"Log in to Confluent Platform with `confluent login --url <mds-url>`.\n"+`Use the "--help" flag to see available commands.`,
	)
)

// Config represents the CLI configuration.
type Config struct {
	DisableFeatureFlags bool `json:"disable_feature_flags"`
	DisablePlugins      bool `json:"disable_plugins"`
	DisablePluginsOnce  bool `json:"disable_plugins_once,omitempty"`
	DisableUpdateCheck  bool `json:"disable_update_check"`
	DisableUpdates      bool `json:"disable_updates,omitempty"`
	EnableColor         bool `json:"enable_color"`

	Platforms        map[string]*Platform        `json:"platforms,omitempty"`
	Credentials      map[string]*Credential      `json:"credentials,omitempty"`
	CurrentContext   string                      `json:"current_context"`
	Contexts         map[string]*Context         `json:"contexts,omitempty"`
	ContextStates    map[string]*ContextState    `json:"context_states,omitempty"`
	SavedCredentials map[string]*LoginCredential `json:"saved_credentials,omitempty"`
	LocalPorts       *LocalPorts                 `json:"local_ports,omitempty"`

	// Deprecated
	AnonymousId string `json:"anonymous_id,omitempty"`
	NoBrowser   bool   `json:"no_browser,omitempty"`
	Ver         string `json:"version,omitempty"`

	// The following configurations are not persisted between runs

	IsTest   bool              `json:"-"`
	Version  *pversion.Version `json:"-"`
	Filename string            `json:"-"`

	overwrittenCurrentContext      string
	overwrittenCurrentEnvironment  string
	overwrittenCurrentKafkaCluster string
}

func (c *Config) SetOverwrittenCurrentContext(context string) {
	if context == "" {
		context = emptyFieldIndicator
	}
	if c.overwrittenCurrentContext == "" {
		c.overwrittenCurrentContext = context
	}
}

func (c *Config) SetOverwrittenCurrentEnvironment(environmentId string) {
	if c.overwrittenCurrentEnvironment == "" {
		c.overwrittenCurrentEnvironment = environmentId
	}
}

func (c *Config) SetOverwrittenCurrentKafkaCluster(clusterId string) {
	if clusterId == "" {
		clusterId = emptyFieldIndicator
	}
	if c.overwrittenCurrentKafkaCluster == "" {
		c.overwrittenCurrentKafkaCluster = clusterId
	}
}

func New() *Config {
	return &Config{
		Platforms:        make(map[string]*Platform),
		Credentials:      make(map[string]*Credential),
		Contexts:         make(map[string]*Context),
		ContextStates:    make(map[string]*ContextState),
		SavedCredentials: make(map[string]*LoginCredential),
		Version:          new(pversion.Version),
	}
}

func (c *Config) DecryptContextStates() error {
	if context := c.Context(); context != nil {
		state := c.ContextStates[context.Name]
		if state != nil {
			if err := state.DecryptAuthToken(context.Name); err != nil {
				return err
			}
			if err := state.DecryptAuthRefreshToken(context.Name); err != nil {
				return err
			}
		}
		context.State = state
	}
	return c.Validate()
}

func (c *Config) DecryptCredentials() error {
	if credentials := c.Credentials; c.Credentials != nil {
		for _, credential := range credentials {
			if credential.APIKeyPair != nil {
				if err := credential.APIKeyPair.DecryptSecret(); err != nil {
					return err
				}
			}
		}
	}
	return c.Validate()
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
				return fmt.Errorf("unable to save configuration file: %w", err)
			}
			return nil
		}
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
	}

	if err := json.Unmarshal(input, c); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, filename, err)
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
			return errors.NewCorruptedConfigError(`context "%s" missing KafkaClusterContext`, context.Name, c.Filename)
		}
		context.KafkaClusterContext.Context = context
		context.State = c.ContextStates[context.Name]
	}

	if runtime.GOOS == "windows" && !c.DisablePluginsOnce {
		c.DisablePlugins = true
		c.DisablePluginsOnce = true
		_ = c.Save()
	}

	return c.Validate()
}

// Save writes the CLI config to disk.
func (c *Config) Save() error {
	tempKafkaCluster := c.resolveOverwrittenKafkaCluster()
	tempEnvironment := c.resolveOverwrittenCurrentEnvironment()
	tempContext := c.resolveOverwrittenContext()
	var tempAuthToken string
	var tempAuthRefreshToken string
	tempCredentials := map[string]string{}

	if c.Context() != nil {
		tempAuthToken = c.Context().GetState().AuthToken
		tempAuthRefreshToken = c.Context().GetState().AuthRefreshToken
		if err := c.encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken); err != nil {
			return err
		}
	}

	if c.Credentials != nil {
		for name, credential := range c.Credentials {
			if credential.APIKeyPair != nil {
				tempCredentials[name] = credential.APIKeyPair.Secret
			}
		}
		if err := c.encryptCredentialsAPISecret(); err != nil {
			return err
		}
	}

	if err := c.Validate(); err != nil {
		return err
	}

	cfg, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	filename := c.GetFilename()

	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return fmt.Errorf("unable to create config directory %s: %w", filename, err)
	}

	if err := os.WriteFile(filename, cfg, 0600); err != nil {
		return fmt.Errorf("unable to write config to file %s: %w", filename, err)
	}

	c.restoreOverwrittenContext(tempContext)
	c.restoreOverwrittenEnvironment(tempEnvironment)
	c.restoreOverwrittenKafkaCluster(tempKafkaCluster)
	c.restoreOverwrittenAuthToken(tempAuthToken)
	c.restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken)
	c.restoreOverwrittenCredentials(tempCredentials)

	return nil
}

func (c *Config) encryptCredentialsAPISecret() error {
	for _, credential := range c.Credentials {
		if credential.APIKeyPair != nil {
			err := credential.APIKeyPair.EncryptSecret()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) encryptContextStateTokens(tempAuthToken, tempAuthRefreshToken string) error {
	if c.Context().GetState().Salt == nil || c.Context().GetState().Nonce == nil {
		salt, nonce, err := secret.GenerateSaltAndNonce()
		if err != nil {
			return err
		}
		c.Context().GetState().Salt = salt
		c.Context().GetState().Nonce = nonce
	}

	if regexp.MustCompile(authTokenRegex).MatchString(tempAuthToken) {
		encryptedAuthToken, err := secret.Encrypt(c.Context().Name, tempAuthToken, c.Context().GetState().Salt, c.Context().GetState().Nonce)
		if err != nil {
			return err
		}
		c.Context().GetState().AuthToken = encryptedAuthToken
	}

	// The Confluent Gov environment and the Confluent Platform MDS return a refresh token that does not match `authRefreshTokenRegex` and cannot be distinguished from an already encrypted refresh token.
	// We prefix encrypted tokens with "AES/GCM/NoPadding" to ensure that they are only encrypted once.
	isUnencryptedConfluentGov := !strings.HasPrefix(tempAuthRefreshToken, secret.AesGcm) && (strings.Contains(c.Context().PlatformName, "confluentgov.com") || strings.Contains(c.Context().PlatformName, "confluentgov-internal.com"))

	isUnencryptedConfluentPlatform := tempAuthRefreshToken != "" && !strings.HasPrefix(tempAuthRefreshToken, secret.AesGcm) && !c.Context().IsCloud(c.IsTest)

	if regexp.MustCompile(authRefreshTokenRegex).MatchString(tempAuthRefreshToken) || isUnencryptedConfluentGov || isUnencryptedConfluentPlatform {
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
func (c *Config) resolveOverwrittenKafkaCluster() string {
	ctx := c.Context()
	var tempKafka string
	if c.overwrittenCurrentKafkaCluster != "" && ctx != nil && ctx.KafkaClusterContext != nil {
		if c.overwrittenCurrentKafkaCluster == emptyFieldIndicator {
			c.overwrittenCurrentKafkaCluster = ""
		}
		tempKafka = ctx.KafkaClusterContext.GetActiveKafkaClusterId()
		ctx.KafkaClusterContext.SetActiveKafkaCluster(c.overwrittenCurrentKafkaCluster)
	}
	return tempKafka
}

// Restore the flag cluster back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenKafkaCluster(tempKafkaCluster string) {
	if tempKafkaCluster != "" {
		c.Context().KafkaClusterContext.SetActiveKafkaCluster(tempKafkaCluster)
	}
}

func (c *Config) restoreOverwrittenAuthToken(tempAuthToken string) {
	if tempAuthToken != "" {
		c.Context().GetState().AuthToken = tempAuthToken
	}
}

func (c *Config) restoreOverwrittenAuthRefreshToken(tempAuthRefreshToken string) {
	if tempAuthRefreshToken != "" {
		c.Context().GetState().AuthRefreshToken = tempAuthRefreshToken
	}
}

func (c *Config) restoreOverwrittenCredentials(tempApiSecrets map[string]string) {
	for name, secret := range tempApiSecrets {
		if secret != "" {
			c.Credentials[name].APIKeyPair.Secret = secret
		}
	}
}

// Switch the initial config context back into the struct so that it is saved and not the flag value
// Return the overwriting flag context value so that it can be restored after writing the file
func (c *Config) resolveOverwrittenContext() string {
	var tempContext string
	if c.overwrittenCurrentContext != "" && c != nil {
		if c.overwrittenCurrentContext == emptyFieldIndicator {
			c.overwrittenCurrentContext = ""
		}
		tempContext = c.CurrentContext
		c.CurrentContext = c.overwrittenCurrentContext
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
func (c *Config) resolveOverwrittenCurrentEnvironment() string {
	var tempEnvironment string
	if c.overwrittenCurrentEnvironment != "" {
		tempEnvironment = c.Context().GetCurrentEnvironment()
		c.Context().SetCurrentEnvironment(c.overwrittenCurrentEnvironment)
	}
	return tempEnvironment
}

// Restore the flag account back into the struct so that it is used for any execution after Save()
func (c *Config) restoreOverwrittenEnvironment(id string) {
	if id != "" {
		c.Context().SetCurrentEnvironment(id)
	}
}

func (c *Config) Validate() error {
	// Validate that current context exists.
	if c.CurrentContext != "" {
		if _, ok := c.Contexts[c.CurrentContext]; !ok {
			log.CliLogger.Trace("current context does not exist")
			return errors.NewCorruptedConfigError(`the current context "%s" does not exist`, c.CurrentContext, c.Filename)
		}
	}

	// Validate that every context:
	// 1. Has no hanging references between the context and the config.
	// 2. Is mapped by name correctly in the config.
	for _, context := range c.Contexts {
		if err := context.validate(); err != nil {
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
			return errors.NewCorruptedConfigError(`context state mismatch for context "%s"`, context.Name, c.Filename)
		}
	}

	// Validate that all context states are mapped to an existing context.
	for contextName := range c.ContextStates {
		if _, ok := c.Contexts[contextName]; !ok {
			log.CliLogger.Trace("context state mapped to nonexistent context")
			return errors.NewCorruptedConfigError(`context state mapping error for context "%s"`, contextName, c.Filename)
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

func (c *Config) AddContext(name, platformName, credentialName string, kafkaClusters map[string]*KafkaClusterConfig, kafka string, state *ContextState, organizationId, environmentId string) error {
	if _, ok := c.Contexts[name]; ok {
		return fmt.Errorf(errors.ContextAlreadyExistsErrorMsg, name)
	}

	credential, ok := c.Credentials[credentialName]
	if !ok {
		return fmt.Errorf(`credential "%s" not found`, credentialName)
	}

	platform, ok := c.Platforms[platformName]
	if !ok {
		return fmt.Errorf(`platform "%s" not found`, platformName)
	}

	ctx, err := newContext(name, platform, credential, kafkaClusters, kafka, state, c, organizationId, environmentId)
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

	// Hardcoded for now, since username/password isn't implemented yet.
	credential := &Credential{
		APIKeyPair:     apiKeyPair,
		CredentialType: APIKey,
		Name:           fmt.Sprintf("%s-%s", APIKey, apiKey),
	}

	if err := c.SaveCredential(credential); err != nil {
		return err
	}

	// Inject credential and platforms name for now, until users can provide custom names.
	platform := &Platform{
		Server: bootstrapURL,
		Name:   strings.TrimPrefix(bootstrapURL, "https://"),
	}

	if err := c.SavePlatform(platform); err != nil {
		return err
	}

	kafkaClusterCfg := &KafkaClusterConfig{
		ID:        "anonymous-id",
		Name:      "anonymous-cluster",
		Bootstrap: bootstrapURL,
		APIKeys:   map[string]*APIKeyPair{apiKey: apiKeyPair},
		APIKey:    apiKey,
	}
	kafkaClusters := map[string]*KafkaClusterConfig{kafkaClusterCfg.ID: kafkaClusterCfg}

	return c.AddContext(name, platform.Name, credential.Name, kafkaClusters, kafkaClusterCfg.ID, nil, "", "")
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
		return fmt.Errorf("credential must have a name")
	}
	c.Credentials[credential.Name] = credential
	return c.Save()
}

func (c *Config) SaveLoginCredential(ctxName string, loginCredential *LoginCredential) error {
	if ctxName == "" {
		return fmt.Errorf("saved credential must match a context")
	}
	c.SavedCredentials[ctxName] = loginCredential
	return c.Save()
}

func (c *Config) SavePlatform(platform *Platform) error {
	if platform.Name == "" {
		return fmt.Errorf("platform must have a name")
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
	return c.Context().GetCredentialType() == APIKey
}

// HasBasicLogin returns true if the user has valid username & password credentials.
func (c *Config) HasBasicLogin() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	if c.IsCloudLogin() {
		return ctx.HasLogin() && ctx.GetCurrentEnvironment() != ""
	} else {
		return ctx.HasLogin()
	}
}

func (c *Config) GetFilename() string {
	if c.Filename == "" {
		c.Filename = GetDefaultFilename()
	}
	return c.Filename
}

func GetDefaultFilename() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".confluent", "config.json")
}

func (c *Config) CheckIsOnPremLogin() error {
	ctx := c.Context()
	if ctx != nil && ctx.PlatformName != "" {
		if !c.isCloud() {
			return nil
		} else {
			return RunningOnPremCommandInCloudErr
		}
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

func (c *Config) IsCloudLogin() bool {
	return c.CheckIsCloudLogin() == nil
}

func (c *Config) HasGovHostname() bool {
	ctx := c.Context()
	if ctx == nil {
		return false
	}

	for _, hostname := range []string{"confluentgov-internal.com", "confluentgov.com"} {
		if strings.Contains(ctx.PlatformName, hostname) {
			return true
		}
	}

	return false
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

// Parse `--context` flag value into config struct
// Call ParseFlagsIntoContext which handles environment and cluster flags
func (c *Config) ParseFlagsIntoConfig(cmd *cobra.Command) error {
	if context, _ := cmd.Flags().GetString("context"); context != "" {
		if _, err := c.FindContext(context); err != nil {
			return err
		}
		c.SetOverwrittenCurrentContext(c.CurrentContext)
		c.CurrentContext = context
	}

	return nil
}
