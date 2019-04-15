package ksql

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
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
	ret, err := c.Client.KSQL.List(ctx, cluster)
	return ret, errors.ConvertAPIError(err)
}

func (c *KSQL) Describe(ctx context.Context, cluster *ksqlv1.KSQLCluster) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Describe()")
	ret, err := c.Client.KSQL.Describe(ctx, cluster)
	return ret, errors.ConvertAPIError(err)
}

func (c *KSQL) Create(ctx context.Context, config *ksqlv1.KSQLClusterConfig) (*ksqlv1.KSQLCluster, error) {
	c.Logger.Log("msg", "ksql.Create()")
	ret, err := c.Client.KSQL.Create(ctx, config)
	return ret, errors.ConvertAPIError(err)
}

func (c *KSQL) Delete(ctx context.Context, cluster *ksqlv1.KSQLCluster) error {
	c.Logger.Log("msg", "ksql.Delete()")
	err := c.Client.KSQL.Delete(ctx, cluster)
	return errors.ConvertAPIError(err)
}
