package v1

import (
	"fmt"
	"strings"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	testserver "github.com/confluentinc/cli/test/test-server"
)

// Context represents a specific CLI context.
type Context struct {
	Name                   string                            `json:"name" hcl:"name"`
	NetrcMachineName       string                            `json:"netrc_machine_name" hcl:"netrc_machine_name"`
	Platform               *Platform                         `json:"-" hcl:"-"`
	PlatformName           string                            `json:"platform" hcl:"platform"`
	Credential             *Credential                       `json:"-" hcl:"-"`
	CredentialName         string                            `json:"credential" hcl:"credential"`
	KafkaClusterContext    *KafkaClusterContext              `json:"kafka_cluster_context" hcl:"kafka_cluster_config"`
	SchemaRegistryClusters map[string]*SchemaRegistryCluster `json:"schema_registry_clusters" hcl:"schema_registry_clusters"`
	State                  *ContextState                     `json:"-" hcl:"-"`
	Config                 *Config                           `json:"-" hcl:"-"`
	LastOrgId              string                            `json:"last_org_id" hcl:"last_org_id"`
	FeatureFlags           *FeatureFlags                     `json:"feature_flags,omitempty" hcl:"feature_flags,omitempty"`
}

func newContext(name string, platform *Platform, credential *Credential,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster, state *ContextState, config *Config,
	orgResourceId string) (*Context, error) {
	ctx := &Context{
		Name:                   name,
		NetrcMachineName:       name,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             credential,
		CredentialName:         credential.Name,
		SchemaRegistryClusters: schemaRegistryClusters,
		State:                  state,
		Config:                 config,
		LastOrgId:              orgResourceId,
	}
	ctx.KafkaClusterContext = NewKafkaClusterContext(ctx, kafka, kafkaClusters)
	err := ctx.validate()
	if err != nil {
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
	if c.SchemaRegistryClusters == nil {
		c.SchemaRegistryClusters = map[string]*SchemaRegistryCluster{}
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
	if c.Credential == nil {
		return false
	}

	credType := c.Credential.CredentialType
	switch credType {
	case Username:
		return c.GetAuthToken() != ""
	case APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, c.Name))
	}
}

func (c *Context) hasBasicCloudLogin() bool {
	if c.Credential == nil {
		return false
	}

	credType := c.Credential.CredentialType
	switch credType {
	case Username:
		return c.GetAuthToken() != "" && c.GetEnvironment().GetId() != ""
	case APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, c.Name))
	}
}

func (c *Context) DeleteUserAuth() error {
	if c.State == nil {
		return nil
	}

	c.State.Auth = nil
	c.State.AuthToken = ""
	c.State.AuthRefreshToken = ""

	err := c.Save()
	return errors.Wrap(err, errors.DeleteUserAuthErrorMsg)
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

	for _, hostname := range ccloudv2.Hostnames {
		if strings.Contains(c.PlatformName, hostname) {
			return true
		}
	}
	return false
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
	if c.State != nil {
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

func (c *Context) GetOrganization() *ccloudv1.Organization {
	if auth := c.GetAuth(); auth != nil {
		return auth.Organization
	}
	return nil
}

func (c *Context) GetSuspensionStatus() *ccloudv1.SuspensionStatus {
	return c.GetOrganization().GetSuspensionStatus()
}

func (c *Context) GetEnvironment() *ccloudv1.Account {
	if auth := c.GetAuth(); auth != nil {
		return auth.Account
	}
	return nil
}

func (c *Context) SetEnvironment(environment *ccloudv1.Account) {
	if c.GetAuth() == nil {
		c.SetAuth(new(AuthConfig))
	}
	c.GetAuth().Account = environment
}

func (c *Context) GetEnvironments() []*ccloudv1.Account {
	if auth := c.GetAuth(); auth != nil {
		return auth.Accounts
	}
	return nil
}

func (c *Context) SetEnvironments(environments []*ccloudv1.Account) {
	if c.GetAuth() == nil {
		c.SetAuth(new(AuthConfig))
	}
	c.GetAuth().Accounts = environments
}

func (c *Context) GetAuthToken() string {
	if state := c.GetState(); state != nil {
		return state.AuthToken
	}
	return ""
}

func (c *Context) GetAuthRefreshToken() string {
	if c.State != nil {
		return c.State.AuthRefreshToken
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
		return c.FeatureFlags.Values
	}
}

func (c *Context) GetNetrcMachineName() string {
	if c != nil {
		return c.NetrcMachineName
	}
	return ""
}

func printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret bool, cluster *KafkaClusterConfig, contextName string) {
	var problems []string
	if missingKey {
		problems = append(problems, errors.APIKeyMissingMsg)
	}
	if mismatchKey {
		problems = append(problems, errors.KeyPairMismatchMsg)
	}
	if missingSecret {
		problems = append(problems, errors.APISecretMissingMsg)
	}
	problemString := strings.Join(problems, ", ")
	output.ErrPrintf(errors.APIKeysMapAutofixMsg, cluster.ID, contextName, problemString, cluster.ID)
}
