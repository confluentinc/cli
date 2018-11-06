package connect

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

// Name description used for registering/disposing GRPC components
const Name = "confluent-connect-plugin"

// Connect describes the shared interface between the GRPC server(plugin) and the GRPC client
type Connect interface {
	List(ctx context.Context, cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, error)
	Describe(ctx context.Context, cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, error)
	DescribeS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error)
	CreateS3Sink(ctx context.Context, config *ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, error)
	UpdateS3Sink(ctx context.Context, cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, error)
	Delete(ctx context.Context, cluster *schedv1.ConnectCluster) error
}

// Plugin mates an interface with Hashicorp plugin object
type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Connect
}

// GRPCClient registers a GRPC client
func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewConnectClient(c)}, nil
}

// GRPCServer registers a GRPC Server
func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterConnectServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["confluent-connect-plugin"] = &Plugin{}
}
