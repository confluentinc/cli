package ksql

import (
	"context"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

type Ksql interface {
	List(ctx context.Context, cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error)
	Create(ctx context.Context, config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error)
	Delete(ctx context.Context, cluster *schedv1.KSQLCluster) error
}
