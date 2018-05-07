package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/connect"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Connect
}

type Connect interface {
	List(ctx context.Context) ([]*schedv1.ConnectCluster, error)
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewConnectClient(c)}, nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterConnectServer(s, &GRPCServer{p.Impl})
	return nil
}

func init() {
	shared.PluginMap["connect"] = &Plugin{}
}
