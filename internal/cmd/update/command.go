package update

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/v50/github"
	update "github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	pgithub "github.com/confluentinc/cli/internal/pkg/github"
	"github.com/confluentinc/cli/internal/pkg/output"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*pcmd.CLICommand
	version string
}

func New(prerunner pcmd.PreRunner, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       fmt.Sprintf("Update the %s.", pversion.FullCLIName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireUpdatesEnabled},
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		version:    version,
	}
	cmd.RunE = c.update

	cmd.Flags().BoolP("yes", "y", false, "Update without prompting.")
	cmd.Flags().Bool("major", false, "Allow major version updates.")
	cmd.Flags().Bool("no-verify", false, "Skip checksum verification of new binary.")

	return cmd
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	major, err := cmd.Flags().GetBool("major")
	if err != nil {
		return err
	}

	noVerify, err := cmd.Flags().GetBool("no-verify")
	if err != nil {
		return err
	}

	output.ErrPrintln("Checking for updates...")

	current, err := semver.ParseTolerant(c.version)
	if err != nil {
		return err
	}

	client := github.NewClient(nil)

	// Intentionally do not paginate. By default there are 30 entries per page, and printing more than 30 release notes would be overkill.
	releases, _, err := client.Repositories.ListReleases(context.Background(), pgithub.Owner, pgithub.Repo, nil)
	if err != nil {
		return err
	}

	releases = getRelevantReleases(releases, current, major)

	if len(releases) == 0 {
		output.Println("Already up to date.")
		return nil
	}

	latestVersion, err := getReleaseVersion(releases[len(releases)-1])
	if err != nil {
		return err
	}

	if major && latestVersion.Major <= current.Major {
		output.Println("No major version updates are available.")
		return nil
	}

	if !major && latestVersion.LTE(current) {
		output.Println("The only available update is a major version update. Use `confluent update --major` to accept the update.")
		return nil
	}

	output.Println("New version of confluent is available")
	output.Printf("Current Version: v%s\n", current)
	output.Printf("Latest Version:  v%s\n", latestVersion)

	for _, release := range releases {
		output.Println()
		output.Println(release.GetBody())
	}

	if !yes {
		ok, err := verify()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	assets, _, err := client.Repositories.ListReleaseAssets(context.Background(), pgithub.Owner, pgithub.Repo, releases[len(releases)-1].GetID(), nil)
	if err != nil {
		return err
	}

	var binaryReader io.Reader
	var bytes int64
	var checksum []byte

	binary := fmt.Sprintf("confluent_%s_%s_%s", latestVersion.String(), runtime.GOOS, runtime.GOARCH)

	for _, asset := range assets {
		if asset.GetName() == binary {
			binaryReader, bytes, err = downloadReleaseAsset(client, asset)
			if err != nil {
				return err
			}
		} else if !noVerify && asset.GetName() == fmt.Sprintf("confluent_%s_checksums.txt", latestVersion.String()) {
			reader, _, err := downloadReleaseAsset(client, asset)
			if err != nil {
				return err
			}

			checksums, err := io.ReadAll(reader)
			if err != nil {
				return err
			}

			checksum, err = findChecksum(string(checksums), binary)
			if err != nil {
				return err
			}
		}
	}

	if binaryReader == nil {
		return fmt.Errorf("could not find release")
	}
	if !noVerify && checksum == nil {
		return fmt.Errorf("could not find checksum")
	}

	var opts update.Options
	if !noVerify {
		opts.Checksum = checksum
	}

	start := time.Now()
	if err := update.Apply(binaryReader, opts); err != nil {
		return err
	}

	mb := convertBytesToMegabytes(bytes)
	s := time.Since(start).Seconds()
	output.Printf("Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, s, mb/s)

	return nil
}

// getRelevantReleases gets an ordered list of releases that have happened since the last update.
func getRelevantReleases(releases []*github.RepositoryRelease, current semver.Version, major bool) []*github.RepositoryRelease {
	var lo int
	var hi int

	var found bool

	for i, release := range releases {
		version, err := getReleaseVersion(release)
		if err != nil {
			continue
		}

		if !found && (major || version.Major == current.Major) {
			lo = i
			found = true
		}

		hi = i
		if version.LTE(current) {
			break
		}
	}

	reversed := make([]*github.RepositoryRelease, hi-lo)
	j := 0
	for i := hi - 1; i >= lo; i-- {
		reversed[j] = releases[i]
		j++
	}

	return reversed
}

func getReleaseVersion(release *github.RepositoryRelease) (semver.Version, error) {
	return semver.ParseTolerant(release.GetTagName())
}

func downloadReleaseAsset(client *github.Client, asset *github.ReleaseAsset) (io.ReadCloser, int64, error) {
	_, url, err := client.Repositories.DownloadReleaseAsset(context.Background(), pgithub.Owner, pgithub.Repo, asset.GetID(), nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, resp.ContentLength, nil
}

func verify() (bool, error) {
	f := form.New(form.Field{
		ID:        "update",
		Prompt:    "Do you want to download and install this update?",
		IsYesOrNo: true,
	})
	if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
		return false, err
	}
	return f.Responses["update"].(bool), nil
}

func findChecksum(checksums, filename string) ([]byte, error) {
	for _, line := range strings.Split(checksums, "\n") {
		if x := strings.SplitN(line, "  ", 2); len(x) == 2 && x[1] == filename {
			checksum := make([]byte, len(x[0])/2)
			if _, err := hex.Decode(checksum, []byte(x[0])); err != nil {
				return nil, err
			}
			return checksum, nil
		}
	}
	return nil, nil
}

func convertBytesToMegabytes(b int64) float64 {
	return float64(b) / 1024.0 / 1024.0
}
