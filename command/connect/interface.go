package connect

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	proto "github.com/confluentinc/cli/shared/connect"
)

type Connect interface {
	List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error)
	Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error)
	DescribeS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error)
	CreateS3Sink(ctx context.Context, config *proto.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error)
	UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error)
	Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Connect
}
