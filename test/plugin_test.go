package test

import (
	"github.com/stretchr/testify/require"
	"os"
)

func (s *CLITestSuite) TestPlugin() {
	path := os.Getenv("PATH")
	err := os.Setenv("PATH", "test/fixtures/input/plugin")
	require.NoError(s.T(), err)
	defer func() {
		err := os.Setenv("PATH", path)
		require.NoError(s.T(), err)
	}()

	tests := []CLITest{
		{args: "", fixture: ""},
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}