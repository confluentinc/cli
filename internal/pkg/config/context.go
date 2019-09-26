package config

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/errors"
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
	SchemaRegistryEndpoint string      `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	SrCredentials          *APIKeyPair `json:"schema_registry_credentials" hcl:"schema_registry_credentials"`
}

type ContextState struct {
	Auth      *AuthConfig `json:"auth" hcl:"auth"`
	AuthToken string      `json:"auth_token" hcl:"auth_token"`
}

// CfgContext represents a specific CLI context.
type Context struct {
	Name           string      `json:"name" hcl:"name"`
	Platform       *Platform   `json:"-" hcl:"-"`
	PlatformName   string      `json:"platform" hcl:"platform"`
	Credential     *Credential `json:"-" hcl:"-"`
	CredentialName string      `json:"credentials" hcl:"credentials"`
	// KafkaClusters store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	KafkaClusters map[string]*KafkaClusterConfig `json:"kafka_clusters" hcl:"kafka_clusters"`
	// Kafka is your active Kafka cluster and references a key in the KafkaClusters map
	Kafka string `json:"kafka_cluster" hcl:"kafka_cluster"`
	// SR map keyed by environment-id
	SchemaRegistryClusters map[string]*SchemaRegistryCluster `json:"schema_registry_cluster" hcl:"schema_registry_cluster"`
	State                  *ContextState                     `json:"-" hcl:"-"`
}

func newContext(name string, platform *Platform, credential *Credential,
	kafkaClusters map[string]*KafkaClusterConfig, kafka string,
	schemaRegistryClusters map[string]*SchemaRegistryCluster, state *ContextState) *Context {
	return &Context{
		Name:                   name,
		Platform:               platform,
		PlatformName:           platform.Name,
		Credential:             credential,
		CredentialName:         credential.Name,
		KafkaClusters:          kafkaClusters,
		Kafka:                  kafka,
		SchemaRegistryClusters: schemaRegistryClusters,
		State: state,
	}
}

func (c *Context) validateKafkaClusterConfig(cluster *KafkaClusterConfig) error {
	if cluster.ID == "" {
		return fmt.Errorf("cluster under context \"%s\" has no %s", c.Name, "id")
	}
	if cluster.Name == "" {
		return fmt.Errorf("cluster under context \"%s\" has no %s", c.Name, "name")
	}
	if cluster.Bootstrap == "" {
		return fmt.Errorf("cluster \"%s\" under context \"%s\" has no %s",
			cluster.Name, c.Name, "bootstrap")
	}
	//if cluster.APIEndpoint == "" {
	//	return fmt.Errorf(missingFieldErrFmt, c.Name, "API endpoint")
	//}
	if cluster.APIKey == "" {
		return &errors.UnspecifiedAPIKeyError{ClusterID: cluster.ID}
	}
	if _, ok := cluster.APIKeys[cluster.APIKey]; !ok {
		return fmt.Errorf("current API key of cluster \"%s\" under context \"%s\" does not exist",
			cluster.Name, c.Name)
	}
	for _, pair := range cluster.APIKeys {
		if pair.Key == "" {
			return fmt.Errorf("an API key of a key pair of cluster \"%s\" under context \"%s\" does not exist",
				cluster.Name, c.Name)
		}
		if pair.Secret == "" {
			return fmt.Errorf("an API secret of a key pair of cluster \"%s\" under context \"%s\" does not exist",
				cluster.Name, c.Name)
		}
	}
	return nil
}

func (c *Context) validateSRCluster(cluster *SchemaRegistryCluster) error {
	// envId validation?
	//srErrFmt := "SR cluster under context \"%s\" has no %s"
	//if cluster.SchemaRegistryEndpoint == "" {
	//	return fmt.Errorf(srErrFmt, c.Name, "endpoint")
	//}
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

func (c *Context) Validate() error {
	if c.Name == "" {
		return errors.New("context has no name")
	}
	if c.CredentialName == "" {
		return &errors.UnspecifiedCredentialError{ContextName: c.Name}
	}
	if c.PlatformName == "" {
		return &errors.UnspecifiedPlatformError{ContextName: c.Name}
	}
	// envId validation?
	if _, ok := c.KafkaClusters[c.Kafka]; c.Kafka != "" && !ok {
		return fmt.Errorf("context \"%s\" has a nonexistent active kafka cluster", c.Name)
	}
	if c.SchemaRegistryClusters == nil {
		c.SchemaRegistryClusters = make(map[string]*SchemaRegistryCluster)
	}
	for envId, sr := range c.SchemaRegistryClusters {
		if sr == nil {
			c.SchemaRegistryClusters[envId] = new(SchemaRegistryCluster)
		}

	}
	for _, cluster := range c.KafkaClusters {
		err := c.validateKafkaClusterConfig(cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) SetActiveCluster(clusterId string) error {
	if _, ok := c.KafkaClusters[clusterId]; !ok {
		return fmt.Errorf("cluster \"%s\" does not exist in context \"%s\"", clusterId, c.Name)
	}
	c.Kafka = clusterId
	return nil
}

func (c *Context) KafkaClusterConfig() *KafkaClusterConfig {
	kafka := c.Kafka
	if kafka == "" {
		return nil
	}
	return c.KafkaClusters[kafka]
}

// SchemaRegistryCluster returns the SchemaRegistryCluster of the CfgContext,
// or an empty SchemaRegistryCluster if there is none set, 
// or an ErrNotLoggedIn if the user is not logged in.
func (c *Context) schemaRegistryCluster() (*SchemaRegistryCluster, error) {
	if !c.hasLogin() {
		return nil, errors.ErrNotLoggedIn
	}
	srCluster, ok := c.SchemaRegistryClusters[c.State.Auth.Account.Id]
	if !ok {
		srCluster = new(SchemaRegistryCluster)
	}
	return srCluster, nil
}

func (c *Context) ActiveKafkaCluster() *KafkaClusterConfig {
	return c.KafkaClusters[c.Kafka]
}

func (c *Context) hasLogin() bool {
	credType := c.Credential.CredentialType
	switch credType {
	case Username:
		state := c.State
		return state.AuthToken != "" && state.Auth != nil &&
			state.Auth.Account != nil && state.Auth.Account.Id != ""
	case APIKey:
		return false
	default:
		panic(fmt.Sprintf("unknown credential type %d in context \"%s\"", credType, c.Name))
	}
}

func (c *Context) authenticatedState() (*ContextState, error) {
	if !c.hasLogin() {
		return nil, errors.ErrNotLoggedIn
	}
	return c.State, nil
}
