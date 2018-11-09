package kafka

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

// Name description used for registering/disposing GRPC components
const Name = "confluent-kafka-plugin"

// Kafka describes the shared interface between the GRPC server(plugin) and the GRPC client
type Kafka interface {
	CreateAPIKey(ctx context.Context, apiKey *schedv1.ApiKey) (*schedv1.ApiKey, error)
	List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error)
	Create(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error)
	Delete(ctx context.Context, cluster *schedv1.KafkaCluster) error
	ListTopics(ctx context.Context) (*ListKafkaTopicReply, error)
	DescribeTopic(ctx context.Context, conf *KafkaAPITopicRequest) (*KafkaTopicDescription, error)
	CreateTopic(ctx context.Context, conf *KafkaAPITopicRequest) (*KafkaAPIResponse, error)
	DeleteTopic(ctx context.Context, conf *KafkaAPITopicRequest) (*KafkaAPIResponse, error)
	UpdateTopic(ctx context.Context, conf *KafkaAPITopicRequest) (*KafkaAPIResponse, error)
	ListACL(ctx context.Context, conf *KafkaAPIACLFilterRequest) (*KafkaAPIACLFilterReply, error)
	CreateACL(ctx context.Context, conf *KafkaAPIACLRequest) (*KafkaAPIResponse, error)
	DeleteACL(ctx context.Context, conf *KafkaAPIACLFilterRequest) (*KafkaAPIResponse, error)
}

// NewKafkaAPITopicRequest returns an instance of KafkaAPITopicRequest
func NewKafkaAPITopicRequest(conf *KafkaTopicSpecification, validate bool) *KafkaAPITopicRequest {
	return &KafkaAPITopicRequest{
		Spec:     conf,
		Validate: validate,
	}
}

// Plugin mates an interface with Hashicorp plugin object
type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Kafka
}

// GRPCClient registers a GRPC client
func (p *Plugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewKafkaClient(c)}, nil
}

// GRPCServer registers a GRPC Server
func (p *Plugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterKafkaServer(s, &GRPCServer{p.Impl})
	return nil
}

// Check that Plugin satisfies GPRCPlugin interface.
var _ plugin.GRPCPlugin = &Plugin{}

func init() {
	shared.PluginMap[Name] = &Plugin{}
}
