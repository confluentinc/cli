package connect

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
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

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewConnectClient(c)}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterConnectServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["connect"] = &Plugin{}
}
