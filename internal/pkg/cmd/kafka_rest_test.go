package cmd

import (
	"testing"

	"github.com/stretchr/testify/suite"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type KafkaRestTestSuite struct {
	suite.Suite
}

func (suite *KafkaRestTestSuite) TestInvalidGetBearerToken() {
	req := suite.Require()
	emptyState := config.ContextState{}
	_, err := pauth.GetDataplaneToken(&emptyState, "invalidhost")
	req.NotNil(err)
}

func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
