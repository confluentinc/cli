package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

type KafkaRestTestSuite struct {
	suite.Suite
}

func (suite *KafkaRestTestSuite) TestBootstrapServersToRestURL() {
	req := suite.Require()

	r, err := bootstrapServersToRestURL("localhost:9092")
	req.Nil(err)
	req.Equal(r, "https://localhost:8090/kafka/v3")

	_, err = bootstrapServersToRestURL("loc")
	req.NotNil(err)

	_, err = bootstrapServersToRestURL("localhost9092")
	req.NotNil(err)

	_, err = bootstrapServersToRestURL("localhost:344")
	req.NotNil(err)
}

func (suite *KafkaRestTestSuite) TestInvalidGetBearerToken() {
	req := suite.Require()
	emptyState := v2.ContextState{}
	_, err := getBearerToken(&emptyState, "invalidhost")
	req.NotNil(err)
}

func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
