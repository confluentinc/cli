package schema_registry

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	schemaregistryv1 "github.com/confluentinc/ccloudapis/schemaregistry/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.SchemaRegistry = (*SchemaRegistry)(nil)

type SchemaRegistry struct {
	Client *ccloud.Client
	Logger *log.Logger
}

func New(client *ccloud.Client, logger *log.Logger) *SchemaRegistry {
	return &SchemaRegistry{Client: client, Logger: logger}
}

func (c *SchemaRegistry) CreateSchemaRegistryCluster(ctx context.Context, config *schemaregistryv1.SchemaRegistryClusterConfig) (*schemaregistryv1.SchemaRegistryCluster, error) {
	return c.Client.SchemaRegistry.CreateSchemaRegistryCluster(ctx, config)
}

func (c *SchemaRegistry) GetSchemaRegistryClusters(ctx context.Context, cluster *schemaregistryv1.SchemaRegistryCluster) ([]*schemaregistryv1.SchemaRegistryCluster, error) {
	return c.Client.SchemaRegistry.GetSchemaRegistryClusters(ctx, cluster)
}

func (c *SchemaRegistry) GetSchemaRegistryCluster(ctx context.Context, cluster *schemaregistryv1.SchemaRegistryCluster) (*schemaregistryv1.SchemaRegistryCluster, error) {
	return c.Client.SchemaRegistry.GetSchemaRegistryCluster(ctx, cluster)
}

func (c *SchemaRegistry) DeleteSchemaRegistryCluster(
	ctx context.Context, cluster *schemaregistryv1.SchemaRegistryCluster) error {
	return c.Client.SchemaRegistry.DeleteSchemaRegistryCluster(ctx, cluster)
}
