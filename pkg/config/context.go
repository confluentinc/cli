package config

import (
	"fmt"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v3/pkg/auth/sso"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

// Context represents a specific CLI context.
type Context struct {
	Name                string                         `json:"name"`
	NetrcMachineName    string                         `json:"netrc_machine_name"`
	PlatformName        string                         `json:"platform"`
	CredentialName      string                         `json:"credential"`
	CurrentEnvironment  string                         `json:"current_environment,omitempty"`
	Environments        map[string]*EnvironmentContext `json:"environments,omitempty"`
	KafkaClusterContext *KafkaClusterContext           `json:"kafka_cluster_context"`
	LastOrgId           string                         `json:"last_org_id,omitempty"`
	FeatureFlags        *FeatureFlags                  `json:"feature_flags,omitempty"`

	// Deprecated
	SchemaRegistryClusters map[string]*SchemaRegistryCluster `json:"schema_registry_clusters,omitempty"`

	Platform   *Platform     `json:"-"`
	Credential *Credential   `json:"-"`
	State      *ContextState `json:"-"`
	Config     *Config       `json:"-"`
}

func newContext(name string, platform *Platform, credential *Credential, kafkaClusters map[string]*KafkaClusterConfig, kafka string, state *ContextState, config *Config, organizationId, environmentId string) (*Context, error) {
	ctx := &Context{
		Name:               name,
		NetrcMachineName:   name,
		Platform:           platform,
		PlatformName:       platform.Name,
		Credential:         credential,
		CredentialName:     credential.Name,
		CurrentEnvironment: environmentId,
		Environments:       map[string]*EnvironmentContext{},
		State:              state,
		Config:             config,
		LastOrgId:          organizationId,
	}
	ctx.KafkaClusterContext = NewKafkaClusterContext(ctx, kafka, kafkaClusters)
	if err := ctx.validate(); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (c *Context) validate() error {
	if c.Name == "" {
		return errors.NewCorruptedConfigError(errors.NoNameContextErrorMsg, "", c.Config.Filename)
	}
	if c.CredentialName == "" || c.Credential == nil {
		return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, c.Name, c.Config.Filename)
	}
	if c.PlatformName == "" || c.Platform == nil {
		return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, c.Name, c.Config.Filename)
	}
	if c.Environments == nil {
		c.Environments = map[string]*EnvironmentContext{}
	}
	if c.State == nil {
		c.State = new(ContextState)
	}
	c.KafkaClusterContext.Validate()
	return nil
}

func (c *Context) Save() error {
	return c.Config.Save()
}

func (c *Context) HasBasicMDSLogin() bool {
	return c.GetCredentialType() == Username && c.GetAuthToken() != ""
}

func (c *Context) hasBasicCloudLogin() bool {
	return c.GetCredentialType() == Username && c.GetAuthToken() != "" && c.GetCurrentEnvironment() != ""
}

func (c *Context) DeleteUserAuth() error {
	if c.State == nil {
		return nil
	}

	c.State.Auth = nil
	c.State.AuthToken = ""
	c.State.AuthRefreshToken = ""

	if err := c.Save(); err != nil {
		return fmt.Errorf("unable to delete user auth: %w", err)
	}

	return nil
}

func (c *Context) UpdateAuthTokens(token, refreshToken string) error {
	c.State.AuthToken = token
	c.State.AuthRefreshToken = refreshToken
	return c.Save()
}

func (c *Context) IsCloud(isTest bool) bool {
	if isTest && c.PlatformName == testserver.TestCloudUrl.String() {
		return true
	}

	for _, hostname := range []string{"confluent.cloud", "confluentgov-internal.com", "confluentgov.com", "cpdev.cloud"} {
		if strings.Contains(c.PlatformName, hostname) {
			return true
		}
	}
	return false
}

func (c *Context) GetCredentialType() CredentialType {
	if c != nil && c.Credential != nil {
		return c.Credential.CredentialType
	}
	return None
}

func (c *Context) GetPlatform() *Platform {
	if c != nil {
		return c.Platform
	}
	return nil
}

func (c *Context) GetPlatformServer() string {
	if platform := c.GetPlatform(); platform != nil {
		return platform.Server
	}
	return ""
}

func (c *Context) GetState() *ContextState {
	if c != nil {
		return c.State
	}
	return nil
}

func (c *Context) GetAuth() *AuthConfig {
	if c.GetState() != nil {
		return c.State.Auth
	}
	return nil
}

func (c *Context) SetAuth(auth *AuthConfig) {
	if c.GetState() == nil {
		c.State = new(ContextState)
	}
	c.GetState().Auth = auth
}

func (c *Context) GetUser() *ccloudv1.User {
	if auth := c.GetAuth(); auth != nil {
		return auth.User
	}
	return nil
}

func (c *Context) GetCurrentOrganization() string {
	if c != nil {
		return c.LastOrgId
	}
	return ""
}

func (c *Context) GetOrganization() *ccloudv1.Organization {
	if auth := c.GetAuth(); auth != nil {
		return auth.Organization
	}
	return nil
}

func (c *Context) GetSuspensionStatus() *ccloudv1.SuspensionStatus {
	return c.GetOrganization().GetSuspensionStatus()
}

func (c *Context) GetCurrentEnvironment() string {
	if c != nil && c.CurrentEnvironment != "" {
		return c.CurrentEnvironment
	}

	if auth := c.GetAuth(); auth != nil {
		return auth.Account.GetId()
	}

	return ""
}

func (c *Context) GetCurrentEnvironmentContext() *EnvironmentContext {
	if id := c.GetCurrentEnvironment(); id != "" {
		return c.Environments[id]
	}
	return nil
}

func (c *Context) SetCurrentEnvironment(id string) {
	c.CurrentEnvironment = id

	if id != "" {
		c.AddEnvironment(id)
	}

	if auth := c.GetAuth(); auth != nil {
		auth.Account = nil
		auth.Accounts = nil
	}
}

func (c *Context) AddEnvironment(id string) {
	if _, ok := c.Environments[id]; !ok {
		c.Environments[id] = &EnvironmentContext{}
	}
}

func (c *Context) DeleteEnvironment(id string) {
	delete(c.Environments, id)
}

func (c *Context) GetCurrentFlinkComputePool() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentFlinkComputePool
	}
	return ""
}

func (c *Context) SetCurrentFlinkComputePool(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentFlinkComputePool = id
	return nil
}

func (c *Context) GetCurrentFlinkCloudProvider() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentFlinkCloudProvider
	}
	return ""
}

func (c *Context) SetCurrentFlinkCloudProvider(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentFlinkCloudProvider = id
	return nil
}
func (c *Context) GetCurrentFlinkRegion() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentFlinkRegion
	}
	return ""
}

func (c *Context) SetCurrentFlinkRegion(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentFlinkRegion = id
	return nil
}

func (c *Context) GetCurrentFlinkCatalog() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentFlinkCatalog
	}
	return ""
}

func (c *Context) SetCurrentFlinkCatalog(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentFlinkCatalog = id
	return nil
}

func (c *Context) GetCurrentFlinkDatabase() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentFlinkDatabase
	}
	return ""
}

func (c *Context) SetCurrentFlinkDatabase(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentFlinkDatabase = id
	return nil
}

func (c *Context) GetCurrentServiceAccount() string {
	if ctx := c.GetCurrentEnvironmentContext(); ctx != nil {
		return ctx.CurrentServiceAccount
	}
	return ""
}

func (c *Context) SetCurrentServiceAccount(id string) error {
	ctx := c.GetCurrentEnvironmentContext()
	if ctx == nil {
		return fmt.Errorf("no environment found")
	}

	ctx.CurrentServiceAccount = id
	return nil
}

func (c *Context) GetAuthToken() string {
	if state := c.GetState(); state != nil {
		return state.AuthToken
	}
	return ""
}

func (c *Context) GetAuthRefreshToken() string {
	if state := c.GetState(); state != nil {
		return state.AuthRefreshToken
	}
	return ""
}

func (c *Context) GetLDFlags(client LaunchDarklyClient) map[string]any {
	if c.FeatureFlags == nil {
		return map[string]any{}
	}

	switch client {
	case CcloudDevelLaunchDarklyClient, CcloudStagLaunchDarklyClient, CcloudProdLaunchDarklyClient:
		return c.FeatureFlags.CcloudValues
	default:
		return c.FeatureFlags.CliValues
	}
}

func (c *Context) GetNetrcMachineName() string {
	if c != nil {
		return c.NetrcMachineName
	}
	return ""
}

func (c *Context) RefreshSession(client *ccloudv1.Client) error {
	if c.IsSso() {
		idToken, refreshToken, err := sso.RefreshTokens(client.BaseURL, c.GetAuthRefreshToken())
		if err != nil {
			return err
		}

		req := &ccloudv1.AuthenticateRequest{
			IdToken:       idToken,
			OrgResourceId: c.GetCurrentOrganization(),
		}

		var res *ccloudv1.AuthenticateReply
		if sso.IsOkta(client.BaseURL) {
			res, err = client.Auth.OktaLogin(req)
		} else {
			res, err = client.Auth.Login(req)
		}
		if err != nil {
			return err
		}

		c.State.AuthToken = res.GetToken()
		c.State.AuthRefreshToken = refreshToken
	} else {
		req := &ccloudv1.AuthenticateRequest{
			RefreshToken:  c.GetAuthRefreshToken(),
			OrgResourceId: c.GetCurrentOrganization(),
		}

		res, err := client.Auth.Login(req)
		if err != nil {
			return err
		}

		c.State.AuthToken = res.GetToken()
		c.State.AuthRefreshToken = res.GetRefreshToken()
	}

	return nil
}

func (c *Context) IsSso() bool {
	return c.GetUser().GetAuthType() == ccloudv1.AuthType_AUTH_TYPE_SSO || c.GetUser().GetSocialConnection() != ""
}

func printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret bool, cluster *KafkaClusterConfig, contextName string) {
	var problems []string
	if missingKey {
		problems = append(problems, "API key missing")
	}
	if mismatchKey {
		problems = append(problems, "key of the dictionary does not match API key of the pair")
	}
	if missingSecret {
		problems = append(problems, "API secret missing")
	}

	output.ErrPrintf(false, "There are malformed API key pair entries in the dictionary for cluster \"%s\" under context \"%s\".\n", cluster.ID, contextName)
	output.ErrPrintf(false, "The issues are the following: %s.\n", strings.Join(problems, ", "))
	output.ErrPrintln(false, "Deleting the malformed entries.")
	output.ErrPrintf(false, "You can re-add the API key pair with `confluent api-key store --resource %s`\n", cluster.ID)
}
