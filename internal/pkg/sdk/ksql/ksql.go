package ksql

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.KSQL = (*KSQL)(nil)

type KSQL struct {
	Client *ccloud.Client
	Logger *log.Logger
}

func New(client *ccloud.Client, logger *log.Logger) *KSQL {
	return &KSQL{Client: client, Logger: logger}
}

func (c *KSQL) List(ctx context.Context, cluster *ksqlv1.KSQLCluster) ([]*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.List()")
	return c.Client.KSQL.List(ctx, cluster)
}

func (c *KSQL) Describe(ctx context.Context, cluster *ksqlv1.KSQLCluster) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Describe()")
	return c.Client.KSQL.Describe(ctx, cluster)
}

func (c *KSQL) Create(ctx context.Context, config *ksqlv1.KSQLClusterConfig) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Create()")
	return c.Client.KSQL.Create(ctx, config)
}

func (c *KSQL) Delete(ctx context.Context, cluster *ksqlv1.KSQLCluster) error {
	c.Logger.Log("msg", "ksql.Delete()")
	return c.Client.KSQL.Delete(ctx, cluster)
}
