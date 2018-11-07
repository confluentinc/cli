package apiKey

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/api-key"
)

type ApiKey interface {
	Create(ctx context.Context, key *schedv1.ApiKey) (*schedv1.ApiKey, error)
	Delete(ctx context.Context, key *schedv1.ApiKey) error
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl ApiKey
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewApiKeyClient(c)}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterApiKeyServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["apiKey"] = &Plugin{}
}
