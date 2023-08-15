package mock

import "github.com/confluentinc/cli/v3/pkg/version"

func NewVersionMock() *version.Version {
	return &version.Version{
		Binary:    "",
		Name:      "mock-cli",
		Version:   "-1.2.3",
		Commit:    "commit-abc",
		BuildDate: "2019-08-19T00:00:00+00:00",
		UserAgent: "mock-user",
		ClientID:  "mock-client-id",
	}
}
