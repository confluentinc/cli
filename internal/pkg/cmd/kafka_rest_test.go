package cmd

import (
	"testing"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"

	"github.com/stretchr/testify/suite"
)

type KafkaRestTestSuite struct {
	suite.Suite
}

func (suite *KafkaRestTestSuite) TestInvalidGetBearerToken() {
	req := suite.Require()
	emptyState := v1.ContextState{}
	_, err := pauth.GetBearerToken(&emptyState, "invalidhost", "lkc-123")
	req.NotNil(err)
}

func TestKafkaRestTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaRestTestSuite))
}
