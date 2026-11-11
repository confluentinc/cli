//go:build live_test && all

package live

func (s *CLILiveTestSuite) TestAlwaysFailLive() {
	s.T().Fatal("intentional failure to test release pipeline update")
}
