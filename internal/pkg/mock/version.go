package mock

import "github.com/confluentinc/cli/internal/pkg/version"

func NewVersionMock() *version.Version {
	return &version.Version{
		Binary:    "",
		Name:      "mock-cli",
		Version:   "-1.2.3",
		Commit:    "commit-abc",
		BuildDate: "2019-08-19T00:00:00+00:00",
		BuildHost: "mock-host",
		UserAgent: "mock-user",
		ClientID:  "mock-client-id",
	}
}
