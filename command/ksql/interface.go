package ksql

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/ksql"
)

type Ksql interface {
	List(ctx context.Context, cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error)
	Create(ctx context.Context, config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error)
	Delete(ctx context.Context, cluster *schedv1.KSQLCluster) error
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Ksql
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewKsqlClient(c)}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterKsqlServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["ksql"] = &Plugin{}
}
