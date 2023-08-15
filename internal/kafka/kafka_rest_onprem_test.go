package kafka

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type KafkaRestTestSuite struct {
	suite.Suite
}

func (suite *KafkaRestTestSuite) TestSetServerURL() {
	req := suite.Require()
	cmd := cobra.Command{Use: "command"}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	cmd.Flags().CountP("verbose", "v", "verbosity")
	client := kafkarestv3.NewAPIClient(kafkarestv3.NewConfiguration())

	SetServerURL(&cmd, client, "localhost:8090")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	SetServerURL(&cmd, client, "localhost:8090/kafka/v3/")
	req.Equal("http://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	SetServerURL(&cmd, client, "localhost:8090/")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "path")
	SetServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "")
	_ = cmd.Flags().Set("ca-cert-path", "path")
	SetServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)
}

func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
