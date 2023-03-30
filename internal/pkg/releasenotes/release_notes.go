package releasenotes

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"

	"github.com/confluentinc/go-netrc/netrc"

	"github.com/confluentinc/cli/internal/pkg/types"
)

const releaseNotesFilename = "release-notes.rst"

var (
	s3ReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName: "Confluent CLI",
		sectionHeaderFormat: func(title string) string {
			return title + "\n" + strings.Repeat("-", len(title))
		},
		codeSnippetFormat: "`",
	}
	docsReleaseNotesBuilderParams = &ReleaseNotesBuilderParams{
		cliDisplayName: "|confluent-cli|",
		sectionHeaderFormat: func(title string) string {
			return fmt.Sprintf("**%s**", title)
		},
		codeSnippetFormat: "``",
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
		line = strings.TrimSpace(line)

		if types.Contains(sections, line) {
			currentSection = line
			continue
		}

		if currentSection != "" {
			if strings.HasPrefix(line, "- ") {
				note := strings.TrimPrefix(line, "- ")
				if note == "PLACEHOLDER" {
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
			} else {
				currentSection = ""
			}
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
	page := 1

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

			pullRequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, commit.GetSHA(), nil)
			if err != nil {
				return err
			}

			if len(pullRequests) > 0 {
				releaseNotes := NewFromBody(pullRequests[0].GetBody())
				log.Printf("SHA: %s, Release Notes: %v\n", commit.GetSHA(), releaseNotes)
				r.Merge(releaseNotes)
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

	r.version = bumpVersion(latestVersion, bump)

	return nil
}

func (r *ReleaseNotes) Write(releaseNotesPath string) error {
	_ = os.Mkdir("release-notes", 0777)

	s3ReleaseNotesBuilder := NewReleaseNotesBuilder(r.version, s3ReleaseNotesBuilderParams)
	s3ReleaseNotes := s3ReleaseNotesBuilder.buildReleaseNotes(r)

	if err := writeFile(filepath.Join("release-notes", "latest-release.rst"), s3ReleaseNotes); err != nil {
		return err
	}

	docsReleaseNotesBuilder := NewReleaseNotesBuilder(r.version, docsReleaseNotesBuilderParams)
	docsReleaseNotes := docsReleaseNotesBuilder.buildReleaseNotes(r)
	updatedDocsPage, err := buildDocsPage(releaseNotesPath, docsPageHeader, docsReleaseNotes)
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join("release-notes", releaseNotesFilename), updatedDocsPage); err != nil {
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
