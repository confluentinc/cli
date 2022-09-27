package test

import "fmt"

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
		{
			args:    fmt.Sprintf("price list --cloud %s --region %s -o json", exampleCloud, exampleRegion),
			fixture: "price/list-json.golden",
		},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
