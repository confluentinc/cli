package kafka

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
	proto "github.com/confluentinc/cli/shared/kafka"
)

type Kafka interface {
	List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error)
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Kafka
}

func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewKafkaClient(c)}, nil
}

func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterKafkaServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap["kafka"] = &Plugin{}
}
