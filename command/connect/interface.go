package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/connect"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Connect interface {
	List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error)
	Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error)
	CreateS3Sink(ctx context.Context, config *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error)
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
