package connect

/*import (
	"context"
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Compile-time check for Interface adherence
var _ ccloud.Connect = (*Connect)(nil)

type Connect struct {
	Client *ccloud.Client
	Logger *log.Logger
}

func New(client *ccloud.Client, logger *log.Logger) *Connect {
	return &Connect{Client: client, Logger: logger}
}

func (c *Connect) List(ctx context.Context, cluster *connectv1.ConnectCluster) ([]*connectv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.List()")
	return c.Client.Connect.List(ctx, cluster)
}

func (c *Connect) Describe(ctx context.Context, cluster *connectv1.ConnectCluster) (*connectv1.ConnectCluster, error) {
	c.Logger.Log("msg", "connect.Describe()")
	return c.Client.Connect.Describe(ctx, cluster)
}

func (c *Connect) DescribeS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.DescribeS3Sink()")
	return c.Client.Connect.DescribeS3Sink(ctx, cluster)
}

func (c *Connect) CreateS3Sink(ctx context.Context, cfg *connectv1.ConnectS3SinkClusterConfig) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.CreateS3Sink()")
	config := &connectv1.ConnectS3SinkClusterConfig{
		Name:           cfg.Name,
		AccountId:      cfg.AccountId,
		KafkaClusterId: cfg.KafkaClusterId,
		Servers:        cfg.Servers,
		Options:        cfg.Options,
	}

	// Resolve kafka user email -> ID
	user, err := c.Client.User.Describe(ctx, &orgv1.User{Email: cfg.UserEmail})
	if err != nil {
		return nil, err
	}
	config.KafkaUserId = user.Id

	// Create the connect cluster
	return c.Client.Connect.CreateS3Sink(ctx, config)
}

func (c *Connect) UpdateS3Sink(ctx context.Context, cluster *connectv1.ConnectS3SinkCluster) (*connectv1.ConnectS3SinkCluster, error) {
	c.Logger.Log("msg", "connect.UpdateS3Sink()")
	return c.Client.Connect.UpdateS3Sink(ctx, cluster)
}

func (c *Connect) Delete(ctx context.Context, cluster *connectv1.ConnectCluster) error {
	c.Logger.Log("msg", "connect.Delete()")
	return c.Client.Connect.Delete(ctx, cluster)
}*/
