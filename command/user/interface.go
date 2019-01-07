package user

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/user"
)

type User interface {
	CreateServiceAccount(ctx context.Context, user *orgv1.User) (*orgv1.User, error)
	UpdateServiceAccount(ctx context.Context, user *orgv1.User) error
	DeactivateServiceAccount(ctx context.Context, user *orgv1.User) error
	GetServiceAccounts(ctx context.Context, user *orgv1.User) ([]*orgv1.User, error)
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl User
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewUserClient(c)}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterUserServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["user"] = &Plugin{}
}
