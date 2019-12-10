package config

import (
	"fmt"

	"github.com/confluentinc/ccloud-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key" hcl:"api_key"`
	Secret string `json:"api_secret" hcl:"api_secret"`
}

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID          string                 `json:"id" hcl:"id"`
	Name        string                 `json:"name" hcl:"name"`
	Bootstrap   string                 `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	APIEndpoint string                 `json:"api_endpoint,omitempty" hcl:"api_endpoint"`
	APIKeys     map[string]*APIKeyPair `json:"api_keys" hcl:"api_keys"`
	// APIKey is your active api key for this cluster and references a key in the APIKeys map
	APIKey string `json:"api_key,omitempty" hcl:"api_key"`
}

type SchemaRegistryCluster struct {
	Id                     string      `json:"id" hcl:"id"`
	SchemaRegistryEndpoint string      `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	SrCredentials          *APIKeyPair `json:"schema_registry_credentials" hcl:"schema_registry_credentials"`
}

type ContextState struct {
	Auth      *AuthConfig `json:"auth" hcl:"auth"`
	AuthToken string      `json:"auth_token" hcl:"auth_token"`
}

// Context represents a specific CLI context.
type Context struct {
	Name           string      `json:"name" hcl:"name"`
	Platform       *Platform   `json:"-" hcl:"-"`
	PlatformName   string      `json:"platform" hcl:"platform"`
	Credential     *Credential `json:"-" hcl:"-"`
	CredentialName string      `json:"credential" hcl:"credential"`
	// KafkaClusters store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	// Mapped by cluster id.
	KafkaClusters map[string]*KafkaClusterConfig `json:"kafka_clusters" hcl:"kafka_clusters"`
	// Kafka is your active Kafka cluster and references a key in the KafkaClusters map
	Kafka                     string `json:"kafka_cluster" hcl:"kafka_cluster"`
	UserSpecifiedKafkaCluster string `json:"-" hcl:"-"`
	// SR map keyed by environment-id.
	SchemaRegistryClusters           map[string]*SchemaRegistryCluster `json:"schema_registry_clusters" hcl:"schema_registry_clusters"`
	UserSpecifiedSchemaRegistryEnvId string                            `json:"-" hcl:"-"`
	State                            *ContextState                     `json:"-" hcl:"-"`
	Logger                           *log.Logger                       `json:"-" hcl:"-"`
	Version                          *version.Version                  `json:"-" hcl:"-"`
	Config                           *Config                           `json:"-" hcl:"-"`
}

func newContext(name string, platform *Platform, credential *Credential,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster, state *ContextState, config *Config) (*Context, error) {
	ctx := &Context{
		Name:                   name,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             credential,
		CredentialName:         credential.Name,
		KafkaClusters:          kafkaClusters,
		Kafka:                  kafka,
		SchemaRegistryClusters: schemaRegistryClusters,
		State:                  state,
		Logger:                 config.Logger,
		Version:                config.Version,
		Config:                 config,
	}
	err := ctx.validate()
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

// TODO: Save contexts after resolution.

func (c *Context) validateKafkaClusterConfig(cluster *KafkaClusterConfig) error {
	if cluster.ID == "" {
		return fmt.Errorf("cluster under context '%s' has no %s", c.Name, "id")
	}
	//if cluster.Name == "" {
	//	return fmt.Errorf("cluster under context '%s' has no %s", c.Name, "name")
	//}
	//if cluster.Bootstrap == "" {
	//	return fmt.Errorf("cluster '%s' under context '%s' has no %s",
	//		cluster.Name, c.Name, "bootstrap")
	//}
	//if cluster.APIEndpoint == "" {
	//	return fmt.Errorf("cluster '%s' under context '%s' has no %s", cluster.Name, c.Name, "API endpoint")
	//}
	if _, ok := cluster.APIKeys[cluster.APIKey]; cluster.APIKey != "" && !ok {
		return fmt.Errorf("current API key of cluster '%s' under context '%s' does not exist. "+
			"Please specify a valid API key",
			cluster.Name, c.Name)
	}
	for _, pair := range cluster.APIKeys {
		if pair.Key == "" {
			return fmt.Errorf("an API key of a key pair of cluster '%s' under context '%s' is missing. "+
				"Please add an API key",
				cluster.Name, c.Name)
		}
	}
	return nil
}

func (c *Context) validateSRCluster(cluster *SchemaRegistryCluster, accountId string) error {
	// envId validation?
	//srErrFmt := "SR cluster under context '%s' has no %s"
	//if cluster.SrCredentials == nil {
	//	return fmt.Errorf(srErrFmt, c.Name, "credentials")
	//}
	//if cluster.SrCredentials.Key == "" {
	//	return fmt.Errorf(srErrFmt, c.Name, "API key")
	//}
	//if cluster.SrCredentials.Secret == "" {
	//	return fmt.Errorf(srErrFmt, c.Name, "API secret")
	//}
	return nil
}

func (c *Context) validate() error {
	if c.Name == "" {
		return errors.New("one of the existing contexts has no name")
	}
	if c.CredentialName == "" {
		return &errors.UnspecifiedCredentialError{ContextName: c.Name}
	}
	if c.PlatformName == "" {
		return &errors.UnspecifiedPlatformError{ContextName: c.Name}
	}
	if _, ok := c.KafkaClusters[c.Kafka]; c.Kafka != "" && !ok {
		return fmt.Errorf("context '%s' has a nonexistent active kafka cluster", c.Name)
	}
	if c.SchemaRegistryClusters == nil {
		c.SchemaRegistryClusters = map[string]*SchemaRegistryCluster{}
	}
	if c.KafkaClusters == nil {
		c.KafkaClusters = map[string]*KafkaClusterConfig{}
	}
	if c.State == nil {
		c.State = new(ContextState)
	}
	// TODO: envId validation?
	for envId, sr := range c.SchemaRegistryClusters {
		if sr == nil {
			c.SchemaRegistryClusters[envId] = new(SchemaRegistryCluster)
		}
	}
	state, err := c.AuthenticatedState()
	if err == nil {
		accId := state.Auth.Account.Id
		if _, ok := c.SchemaRegistryClusters[accId]; !ok {
			c.SchemaRegistryClusters[accId] = new(SchemaRegistryCluster)
		}
	}
	for _, cluster := range c.KafkaClusters {
		err := c.validateKafkaClusterConfig(cluster)
		if err != nil {
			return err
		}
	}
	for accountId, cluster := range c.SchemaRegistryClusters {
		err := c.validateSRCluster(cluster, accountId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) SetActiveKafkaCluster(clusterId string, client *ccloud.Client) error {
	_, err := c.FindKafkaCluster(clusterId, client)
	if err != nil {
		return err
	}
	c.Kafka = clusterId
	return c.Save()
}

func (c *Context) SetUserSpecifiedKafkaCluster(clusterId string, client *ccloud.Client) error {
	_, err := c.FindKafkaCluster(clusterId, client)
	if err != nil {
		return err
	}
	c.UserSpecifiedKafkaCluster = clusterId
	return nil
}

// SchemaRegistryCluster returns the SchemaRegistryCluster of the Context,
// or an empty SchemaRegistryCluster if there is none set, 
// or an ErrNotLoggedIn if the user is not logged in.
func (c *Context) schemaRegistryCluster(client *ccloud.Client) (*SchemaRegistryCluster, error) {
	state, err := c.AuthenticatedState()
	if err != nil {
		return nil, err
	}
	accountId := state.Auth.Account.Id
	cluster, ok := c.SchemaRegistryClusters[accountId]
	if !ok {
		return nil, nil
	}
	if cluster.SchemaRegistryEndpoint == "" || cluster.Id == "" {
		resolver := NewResolver(c, client)
		resolvedCluster, err := resolver.ResolveSchemaRegistryByAccountId(accountId)
		if err != nil {
			return nil, err
		}
		cluster = resolvedCluster
		c.SchemaRegistryClusters[accountId] = cluster
		err = c.Save()
		if err != nil {
			return nil, err
		}
	}
	return cluster, nil
}

func (c *Context) ActiveKafkaCluster(client *ccloud.Client) (*KafkaClusterConfig, error) {
	var clusterId string
	if c.UserSpecifiedKafkaCluster != "" {
		clusterId = c.UserSpecifiedKafkaCluster
	} else {
		clusterId = c.Kafka
	}
	cluster, err := c.FindKafkaCluster(clusterId, client)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Context) FindKafkaCluster(clusterId string, client *ccloud.Client) (*KafkaClusterConfig, error) {
	if _, ok := c.KafkaClusters[clusterId]; !ok {
		if client == nil {
			return nil, errors.ErrNoKafkaContext
		}
		// Resolve cluster details.
		resolver := NewResolver(c, client)
		cluster, err := resolver.ResolveCluster(clusterId)
		if err != nil {
			return nil, err
		}
		c.KafkaClusters[clusterId] = cluster
		err = c.Save()
		if err != nil {
			return nil, err
		}
	}
	return c.KafkaClusters[clusterId], nil
}

func (c *Context) UseAPIKey(apiKey string, clusterId string, client *ccloud.Client) error {
	kcc, err := c.FindKafkaCluster(clusterId, client)
	if err != nil {
		return err
	}
	if _, ok := kcc.APIKeys[apiKey]; !ok {
		// Fetch API key error.
		ctxClient := NewContextClient(c, client)
		return ctxClient.FetchAPIKeyError(apiKey, clusterId)
	}
	kcc.APIKey = apiKey
	return c.Save()
}

func (c *Context) Save() error {
	return c.Config.Save()
}

func (c *Context) hasLogin() bool {
	credType := c.Credential.CredentialType
	switch credType {
	case Username:
		state := c.State
		return state != nil && state.AuthToken != "" && state.Auth != nil &&
			state.Auth.Account != nil && state.Auth.Account.Id != ""
	case APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context '%s'", credType, c.Name))
	}
}

func (c *Context) AuthenticatedState() (*ContextState, error) {
	if !c.hasLogin() {
		return nil, errors.ErrNotLoggedIn
	}
	return c.State, nil
}
