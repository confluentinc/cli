package ksql

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

// Name description used for registering/disposing GRPC components
const Name = "confluent-ksql-plugin"

// Ksql describes the shared interface between the GRPC server(plugin) and the GRPC client
type Ksql interface {
	List(ctx context.Context, cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, error)
	Create(ctx context.Context, config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, error)
	Delete(ctx context.Context, cluster *schedv1.KSQLCluster) error
}

// Plugin mates an interface with Hashicorp plugin object
type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Ksql
}

// GRPCClient registers a GRPC client
func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewKsqlClient(c)}, nil
}

// GRPCServer registers a GRPC Server
func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterKsqlServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["confluent-ksql-plugin"] = &Plugin{}
}
