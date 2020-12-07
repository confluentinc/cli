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
			args:    fmt.Sprintf("price list --cloud %s --region %s", exampleCloud, exampleRegion),
			fixture: "price/list.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
