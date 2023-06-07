//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/client.go --pkg mock --selfpkg github.com/confluentinc/cli client.go Client
package update

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
	"github.com/jonboulle/clockwork"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Client lets you check for updated application binaries and install them if desired
type Client interface {
	CheckForUpdates(cliName, currentVersion string, forceCheck bool) (string, string, error)
	GetLatestReleaseNotes(cliName, currentVersion string) (string, []string, error)
	PromptToDownload(cliName, currVersion, latestVersion string, releaseNotes string, confirm bool) bool
	UpdateBinary(cliName, version, path string, noVerify bool) error
}

type client struct {
	*ClientParams
	// @VisibleForTesting, defaults to the system clock
	clock clockwork.Clock
	// @VisibleForTesting, defaults to the OS filesystem
	fs pio.FileSystem
}

var _ Client = (*client)(nil)

// ClientParams are used to configure the update.Client
type ClientParams struct {
	Repository Repository
	Out        pio.File
	// Optional, if you want to disable checking for updates
	DisableCheck bool
	// Optional, if you wish to rate limit your update checks. The parent directories must exist.
	CheckFile string
	// Optional, defaults to checking once every 24h
	CheckInterval time.Duration
	OS            string
}

// NewClient returns a client for updating CLI binaries
func NewClient(params *ClientParams) *client {
	if params.CheckInterval == 0 {
		params.CheckInterval = 24 * time.Hour
	}
	if params.OS == "" {
		params.OS = runtime.GOOS
	}
	return &client{
		ClientParams: params,
		clock:        clockwork.NewRealClock(),
		fs:           &pio.RealFileSystem{},
	}
}

// CheckForUpdates checks for new versions in the repo
func (c *client) CheckForUpdates(cliName, currentVersion string, forceCheck bool) (string, string, error) {
	if c.DisableCheck {
		return "", "", nil
	}

	shouldCheck, err := c.readCheckFile()
	if err != nil {
		return "", "", err
	}
	if !shouldCheck && !forceCheck {
		return "", "", nil
	}

	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		return "", "", errors.Wrapf(err, errors.ParseVersionErrorMsg, cliName, currentVersion)
	}

	latestMajorVersion, latestMinorVersion, err := c.Repository.GetLatestMajorAndMinorVersion(cliName, currVersion)
	if err != nil {
		return "", "", err
	}

	// If there is a major version update for `ccloud`, it will be found under the name `confluent`.
	if cliName == "ccloud" {
		latestMajorVersion, _, err = c.Repository.GetLatestMajorAndMinorVersion("confluent", currVersion)
		if err != nil {
			return "", "", err
		}
	}

	// After fetching the latest version, we touch the file so that we don't make the request again for 24hrs.
	if err := c.touchCheckFile(); err != nil {
		return "", "", errors.Wrap(err, errors.TouchLastCheckFileErrorMsg)
	}

	var major, minor string
	if latestMajorVersion != nil && isLessThanVersion(currVersion, latestMajorVersion) {
		major = latestMajorVersion.Original()
	}
	if isLessThanVersion(currVersion, latestMinorVersion) {
		minor = latestMinorVersion.Original()
	}
	return major, minor, nil
}

func (c *client) GetLatestReleaseNotes(cliName, currentVersion string) (string, []string, error) {
	latestReleaseNotesVersions, err := c.Repository.GetLatestReleaseNotesVersions(cliName, currentVersion)
	if err != nil {
		return "", nil, err
	}

	var latestVersion string
	allReleaseNotes := make([]string, len(latestReleaseNotesVersions))

	for i, releaseNotesVersion := range latestReleaseNotesVersions {
		releaseNotes, err := c.Repository.DownloadReleaseNotes(cliName, releaseNotesVersion.String())
		if err != nil {
			return "", nil, err
		}

		latestVersion = releaseNotesVersion.Original()
		allReleaseNotes[i] = releaseNotes
	}

	return latestVersion, allReleaseNotes, nil
}

// SemVer considers x.x.x-yyy to be less (older) than x.x.x
func isLessThanVersion(curr, latest *version.Version) bool {
	splitCurList := strings.Split(curr.String(), "-")
	splitLatestList := strings.Split(latest.String(), "-")

	truncatedCur, _ := version.NewVersion(splitCurList[0])
	truncatedLatest, _ := version.NewVersion(splitLatestList[0])

	compareNum := truncatedCur.Compare(truncatedLatest)
	if compareNum < 0 {
		return true
	} else if compareNum == 0 {
		if len(splitCurList) > 1 && len(splitLatestList) == 1 {
			return false
		} else if len(splitCurList) == 1 && len(splitLatestList) > 1 {
			return true
		}
		return curr.LessThan(latest)
	}
	return false
}

// PromptToDownload displays an interactive CLI prompt to download the latest version
func (c *client) PromptToDownload(cliName, currVersion, latestVersion, releaseNotes string, confirm bool) bool {
	if confirm && !c.fs.IsTerminal(c.Out.Fd()) {
		log.CliLogger.Warn("disable confirm as stdout is not a tty")
		confirm = false
	}

	fmt.Fprintf(c.Out, errors.PromptToDownloadDescriptionMsg, cliName, currVersion, latestVersion, releaseNotes)

	if !confirm {
		return true
	}

	for {
		fmt.Fprint(c.Out, "Do you want to download and install this update? (y/n): ")

		reader := c.fs.NewBufferedReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		choice := strings.TrimRightFunc(input, unicode.IsSpace)

		switch choice {
		case "yes", "y", "Y":
			return true
		case "no", "n", "N":
			return false
		default:
			fmt.Fprintf(c.Out, "%s is not a valid choice\n", choice)
			continue
		}
	}
}

// UpdateBinary replaces the named binary at path with the desired version
func (c *client) UpdateBinary(cliName, version, path string, noVerify bool) error {
	downloadDir, err := c.fs.MkdirTemp("", cliName)
	if err != nil {
		return errors.Wrapf(err, errors.GetTempDirErrorMsg, cliName)
	}
	defer func() {
		if err := c.fs.RemoveAll(downloadDir); err != nil {
			log.CliLogger.Warnf("unable to clean up temp download dir %s: %s", downloadDir, err)
		}
	}()

	fmt.Fprintf(c.Out, "Downloading %s version %s...\n", cliName, version)
	startTime := c.clock.Now()

	payload, err := c.Repository.DownloadVersion(cliName, version, downloadDir)
	if err != nil {
		return errors.Wrapf(err, errors.DownloadVersionErrorMsg, cliName, version, downloadDir)
	}

	mb := float64(len(payload)) / 1024.0 / 1024.0
	timeSpent := c.clock.Now().Sub(startTime).Seconds()
	fmt.Fprintf(c.Out, "Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, timeSpent, mb/timeSpent)

	opts := update.Options{}
	if !noVerify {
		content, err := c.Repository.DownloadChecksums(cliName, version)
		if err != nil {
			return errors.Wrapf(err, "failed to download checksums file")
		}

		binary := getBinaryName(version, GetOs(), runtime.GOARCH)
		checksum, err := findChecksum(content, binary)
		if err != nil {
			return err
		}

		opts.Checksum = make([]byte, len(checksum)/2)
		if _, err := hex.Decode(opts.Checksum, []byte(checksum)); err != nil {
			return err
		}
	}

	return update.Apply(bytes.NewReader(payload), opts)
}

func getBinaryName(version, os, arch string) string {
	name := fmt.Sprintf("confluent_%s_%s_%s", version, os, arch)
	if os == "windows" {
		name += ".exe"
	}
	return name
}

func findChecksum(content, binary string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		x := strings.Split(line, "  ")
		if len(x) == 2 && x[1] == binary {
			return x[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", binary)
}

func (c *client) readCheckFile() (bool, error) {
	// If CheckFile is not provided, then we'll always perform the check
	if c.CheckFile == "" {
		return true, nil
	}
	info, err := c.fs.Stat(c.CheckFile)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	// if the file doesn't exist, check updates anyway -- indicates a new CLI install
	if os.IsNotExist(err) {
		return true, nil
	}
	// if the file was updated in the last (interval), don't check again
	if info.ModTime().After(c.clock.Now().Add(-1 * c.CheckInterval)) {
		return false, nil
	}
	return true, nil
}

func (c *client) touchCheckFile() error {
	// If CheckFile is not provided, then we'll skip touching
	if c.CheckFile == "" {
		return nil
	}

	if _, err := c.fs.Stat(c.CheckFile); os.IsNotExist(err) {
		if f, err := c.fs.Create(c.CheckFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := c.fs.Chtimes(c.CheckFile, c.clock.Now(), c.clock.Now()); err != nil {
		return err
	}
	return nil
}
