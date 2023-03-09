package kafka

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/stretchr/testify/suite"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

	setServerURL(&cmd, client, "localhost:8090")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	setServerURL(&cmd, client, "localhost:8090/kafka/v3/")
	req.Equal("http://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	setServerURL(&cmd, client, "localhost:8090/")
	req.Equal("http://localhost:8090/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "path")
	setServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)

	_ = cmd.Flags().Set("client-cert-path", "")
	_ = cmd.Flags().Set("ca-cert-path", "path")
	setServerURL(&cmd, client, "localhost:8090/kafka")
	req.Equal("https://localhost:8090/kafka/v3", client.GetConfig().BasePath)
}

func TestKafkaRestTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KafkaRestTestSuite))
}
