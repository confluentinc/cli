package v3

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
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

func newContext(name string, platform *v2.Platform, credential *v2.Credential,
	kafkaClusters map[string]*v1.KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*v2.SchemaRegistryCluster, state *v2.ContextState, config *Config) (*Context, error) {
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
	err := ctx.validate()
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (c *Context) validateKafkaClusterConfig(cluster *v1.KafkaClusterConfig) error {
	if cluster.ID == "" {
		return fmt.Errorf("cluster under context '%s' has no %s", c.Name, "id")
	}
	if _, ok := cluster.APIKeys[cluster.APIKey]; cluster.APIKey != "" && !ok {
		_, _ = fmt.Fprintf(os.Stderr, "Current API key '%s' of cluster '%s' under context '%s' is not found.\n"+
			"Removing current API key setting for the cluster.\n"+
			"You can re-add the API key with 'ccloud api-key store' and set current API key with 'ccloud api-key use'.\n",
			cluster.APIKey, cluster.Name, c.Name)
		cluster.APIKey = ""
		err := c.Save()
		if err != nil {
			return fmt.Errorf("unable to reset invalid active API key")
		}
	}
	return c.validateApiKeysDict(cluster)
}

func (c *Context) validateApiKeysDict(cluster *v1.KafkaClusterConfig) error {
	missingKey := false
	mismatchKey := false
	missingSecret := false
	for k, pair := range cluster.APIKeys {
		if pair.Key == "" {
			delete(cluster.APIKeys, k)
			missingKey = true
			continue
		}
		if k != pair.Key {
			delete(cluster.APIKeys, k)
			mismatchKey = true
			continue
		}
		if pair.Secret == "" {
			delete(cluster.APIKeys, k)
			missingSecret = true
		}
	}
	if missingKey || mismatchKey || missingSecret {
		c.printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret, cluster)
		err := c.Save()
		if err != nil {
			return fmt.Errorf("unable to clear invalid API key pairs")
		}
	}
	return nil
}

func (c *Context) printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret bool, cluster *v1.KafkaClusterConfig) {
	var problems []string
	if missingKey {
		problems = append(problems, "'API key missing'")
	}
	if mismatchKey {
		problems = append(problems, "'key of the dictionary does not match API key of the pair'")
	}
	if missingSecret {
		problems = append(problems, "'API secret missing'")
	}
	problemString := strings.Join(problems, ", ")
	_, _ = fmt.Fprintf(os.Stderr, "There are malformed API key secret pair entries in the dictionary for cluster '%s' under context '%s'.\n"+
		"The issues are the following: "+problemString+".\n"+
		"Deleting the malformed entries.\n"+
		"You can re-add the API key secret pair with 'ccloud api-key store'\n",
		cluster.Name, c.Name)
}

func (c *Context) validate() error {
	if c.Name == "" {
		return errors.New("one of the existing contexts has no name")
	}
	if c.CredentialName == "" || c.Credential == nil {
		return &errors.UnspecifiedCredentialError{ContextName: c.Name}
	}
	if c.PlatformName == "" || c.Platform == nil {
		return &errors.UnspecifiedPlatformError{ContextName: c.Name}
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

func (c *Context) HasMDSLogin() bool {
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

func (c *Context) hasLogin() bool {
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
		return errors.Wrap(err, "unable to delete user auth")
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
