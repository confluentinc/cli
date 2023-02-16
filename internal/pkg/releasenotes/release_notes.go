package releasenotes

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/go-netrc/netrc"
	"github.com/google/go-github/v50/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
)

const releaseNotesFilename = "release-notes.rst"

var (
	s3ReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      "Confluent CLI",
		sectionHeaderFormat: "%s\n-------------",
	}
	docsReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName:      "|confluent-cli|",
		sectionHeaderFormat: "**%s**",
	}
)

const (
	owner = "confluentinc"
	repo  = "cli"
)

const (
	major = iota
	minor
	patch
)

var sections = []string{
	"Breaking Changes",
	"New Features",
	"Bug Fixes",
}

type ReleaseNotes struct {
	version string
	bump    string

	major []string
	minor []string
	patch []string
}

func New() *ReleaseNotes {
	return &ReleaseNotes{
		major: []string{},
		minor: []string{},
		patch: []string{},
	}
}

func NewFromBody(body string) *ReleaseNotes {
	r := New()

	currentSection := ""
	for _, line := range strings.Split(body, "\n") {
		if utils.Contains(sections, line) {
			currentSection = line
		}

		if strings.HasPrefix(line, "- ") {
			note := strings.TrimPrefix(line, "- ")

			if note == "<PLACEHOLDER>" {
				continue
			}

			switch currentSection {
			case majorSectionTitle:
				r.major = append(r.major, note)
			case minorSectionTitle:
				r.minor = append(r.minor, note)
			case patchSectionTitle:
				r.patch = append(r.patch, note)
			}
		}

		if line == "" {
			currentSection = ""
		}
	}

	return r
}

func (r *ReleaseNotes) Merge(other *ReleaseNotes) {
	r.major = append(r.major, other.major...)
	r.minor = append(r.minor, other.minor...)
	r.patch = append(r.patch, other.patch...)
}

func (r *ReleaseNotes) GetBump() (string, error) {
	if len(r.major) > 0 {
		return "major", nil
	}
	if len(r.minor) > 0 {
		return "minor", nil
	}
	if len(r.patch) > 0 {
		return "patch", nil
	}

	return "", fmt.Errorf("no updates found")
}

func (r *ReleaseNotes) ReadFromGithub() error {
	ctx := context.Background()

	// Authenticate to avoid being rate-limited
	githubToken, err := getGithubToken()
	if err != nil {
		return err
	}
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})))

	latestRelease, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return err
	}

	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, nil)
	if err != nil {
		return err
	}
	var latestReleaseTagSha string
	for _, tag := range tags {
		if tag.GetName() == latestRelease.GetTagName() {
			latestReleaseTagSha = tag.GetCommit().GetSHA()
		}
	}

	done := false
	page := 0

	for !done {
		opts := &github.CommitsListOptions{ListOptions: github.ListOptions{Page: page}}
		commits, _, err := client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return err
		}

		for _, commit := range commits {
			if commit.GetSHA() == latestReleaseTagSha {
				done = true
				break
			}

			// Search for PRs
			issuesSearchResult, _, err := client.Search.Issues(ctx, commit.GetSHA(), nil)
			if err != nil {
				return err
			}

			if len(issuesSearchResult.Issues) > 0 {
				body := issuesSearchResult.Issues[0].GetBody()
				r.Merge(NewFromBody(body))
			}
		}

		page++
	}

	latestVersion, err := version.NewSemver(latestRelease.GetTagName())
	if err != nil {
		return err
	}

	bump, err := r.GetBump()
	if err != nil {
		return err
	}
	r.bump = bump

	r.version = bumpVersion(latestVersion, bump)

	return nil
}

func (r *ReleaseNotes) Write(releaseNotesPath string) error {
	_ = os.Mkdir("release-notes", 0777)

	s3ReleaseNotesBuilder := NewReleaseNotesBuilder(r.version, s3ReleaseNotesBuilderParams)
	s3ReleaseNotes := s3ReleaseNotesBuilder.buildS3ReleaseNotes(r)

	if err := writeFile(filepath.Join("release-notes", "latest-release.rst"), s3ReleaseNotes); err != nil {
		return err
	}

	docsReleaseNotesBuilder := NewReleaseNotesBuilder(r.version, docsReleaseNotesBuilderParams)
	docsReleaseNotes := docsReleaseNotesBuilder.buildDocsReleaseNotes(r)
	updatedDocsPage, err := buildDocsPage(releaseNotesPath, docsPageHeader, docsReleaseNotes)
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join("release-notes", releaseNotesFilename), updatedDocsPage); err != nil {
		return err
	}

	if err := writeFile(filepath.Join("release-notes", "bump.txt"), r.bump); err != nil {
		return err
	}
	if err := writeFile(filepath.Join("release-notes", "version.txt"), r.version); err != nil {
		return err
	}

	return nil
}

func getGithubToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	machine, err := netrc.FindMachine(filepath.Join(home, ".netrc"), "github.com")
	if err != nil {
		return "", err
	}

	return machine.Password, nil
}

func bumpVersion(v *version.Version, bump string) string {
	segments := v.Segments()

	switch bump {
	case "major":
		segments[major]++
		segments[minor] = 0
		segments[patch] = 0
	case "minor":
		segments[minor]++
		segments[patch] = 0
	case "patch":
		segments[patch]++
	}

	return fmt.Sprintf("%d.%d.%d", segments[major], segments[minor], segments[patch])
}

func buildDocsPage(releaseNotesPath, docsHeader, latestReleaseNotes string) (string, error) {
	docsUpdateHandler := NewDocsUpdateHandler(docsHeader, filepath.Join(releaseNotesPath, releaseNotesFilename))
	return docsUpdateHandler.getUpdatedDocsPage(latestReleaseNotes)
}

func writeFile(filePath, fileContent string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, fileContent)
	return err
}
