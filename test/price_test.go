package test

import (
	"fmt"
)

const (
	exampleCloud  = "aws"
	exampleRegion = "us-east-1"
)

func (s *CLITestSuite) TestPriceList() {
	tests := []CLITest{
		{
			Args:    fmt.Sprintf("price list --cloud %s --region %s", exampleCloud, exampleRegion),
			Fixture: "price/list.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}
