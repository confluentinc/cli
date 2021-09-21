package v3

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	testserver "github.com/confluentinc/cli/test/test-server"
)

// Context represents a specific CLI context.
type Context struct {
	Name                   string                               `json:"name" hcl:"name"`
	Platform               *v2.Platform                         `json:"-" hcl:"-"`
	PlatformName           string                               `json:"platform" hcl:"platform"`
	Credential             *v2.Credential                       `json:"-" hcl:"-"`
	CredentialName         string                               `json:"credential" hcl:"credential"`
	KafkaClusterContext    *KafkaClusterContext                 `json:"kafka_cluster_context" hcl:"kafka_cluster_config"`
	SchemaRegistryClusters map[string]*v2.SchemaRegistryCluster `json:"schema_registry_clusters" hcl:"schema_registry_clusters"`
	State                  *v2.ContextState                     `json:"-" hcl:"-"`
	Logger                 *log.Logger                          `json:"-" hcl:"-"`
	Config                 *Config                              `json:"-" hcl:"-"`
}

func newContext(name string, platform *v2.Platform, credential *v2.Credential, kafkaClusters map[string]*v1.KafkaClusterConfig, kafka string, schemaRegistryClusters map[string]*v2.SchemaRegistryCluster, state *v2.ContextState, config *Config) (*Context, error) {
	ctx := &Context{
		Name:                   name,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             credential,
		CredentialName:         credential.Name,
		SchemaRegistryClusters: schemaRegistryClusters,
		State:                  state,
		Logger:                 config.Logger,
		Config:                 config,
	}
	ctx.KafkaClusterContext = NewKafkaClusterContext(ctx, kafka, kafkaClusters)
	return ctx, ctx.validate()
}

func (c *Context) validate() error {
	if c.Name == "" {
		return errors.NewCorruptedConfigError(errors.NoNameContextErrorMsg, "", c.Config.Filename, c.Logger)
	}
	if c.CredentialName == "" || c.Credential == nil {
		return errors.NewCorruptedConfigError(errors.UnspecifiedCredentialErrorMsg, c.Name, c.Config.Filename, c.Logger)
	}
	if c.PlatformName == "" || c.Platform == nil {
		return errors.NewCorruptedConfigError(errors.UnspecifiedPlatformErrorMsg, c.Name, c.Config.Filename, c.Logger)
	}
	if c.SchemaRegistryClusters == nil {
		c.SchemaRegistryClusters = map[string]*v2.SchemaRegistryCluster{}
	}
	if c.State == nil {
		c.State = new(v2.ContextState)
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
	case v2.Username:
		return c.State != nil && c.State.AuthToken != ""
	case v2.APIKey:
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
	case v2.Username:
		return c.State != nil && c.State.AuthToken != "" && c.State.Auth != nil && c.State.Auth.Account != nil && c.State.Auth.Account.Id != ""
	case v2.APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, c.Name))
	}
}

func (c *Context) DeleteUserAuth() error {
	if c.State == nil {
		return nil
	}
	c.State.AuthToken = ""
	c.State.Auth = nil
	err := c.Save()
	if err != nil {
		return errors.Wrap(err, errors.DeleteUserAuthErrorMsg)
	}
	return nil
}

func (c *Context) GetCurrentEnvironmentId() string {
	// non environment contexts
	if c.State.Auth == nil {
		return ""
	}
	return c.State.Auth.Account.Id
}

func (c *Context) UpdateAuthToken(token string) error {
	c.State.AuthToken = token
	err := c.Save()
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) IsCloud(isTest bool) bool {
	if isTest && c.PlatformName == testserver.TestCloudURL.String() {
		return true
	}

	for _, hostname := range CCloudHostnames {
		if strings.Contains(c.PlatformName, hostname) {
			return true
		}
	}
	return false
}

func printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret bool, cluster *v1.KafkaClusterConfig, contextName string) {
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
	_, _ = fmt.Fprintf(os.Stderr, errors.APIKeysMapAutofixMsg, cluster.ID, contextName, problemString, cluster.ID)
}
