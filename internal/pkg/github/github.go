package github

import (
	"context"

	"github.com/blang/semver"
	"github.com/google/go-github/v50/github"
)

const (
	Owner = "confluentinc"
	Repo  = "cli"
)

func GetLatestRelease() (semver.Version, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(context.Background(), Owner, Repo)
	if err != nil {
		return semver.Version{}, err
	}

	return semver.ParseTolerant(release.GetTagName())
}
