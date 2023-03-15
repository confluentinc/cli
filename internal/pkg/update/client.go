//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/client.go --pkg mock --selfpkg github.com/confluentinc/cli client.go Client
package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-version"
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
	VerifyChecksum(newBin, cliName, version string) error
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

func (c *client) VerifyChecksum(newBin, cliName, version string) error {
	// Step 1: Compute actual hash of downloaded file

	f, err := os.Open(newBin)
	if err != nil {
		return errors.Wrap(err, "failed to open new binary file for checksum verification")
	}

	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return errors.Wrap(err, "failed to load new binary file for checksum verification")
	}

	actualHash := fmt.Sprintf("%x", h.Sum(nil))

	// Step 2: Download expected checksum

	allChecksumsForVersion, err := c.Repository.DownloadChecksums(cliName, version)
	if err != nil {
		return errors.Wrapf(err, "failed to download checksums file")
	}

	log.CliLogger.Tracef("Actual hash of new binary: %s", actualHash)

	// Step 3: Check if checksums match

	// Don't filter checksums for the current platform, just
	// check if any platform's checksum is a match (this is arguably better
	// in case we add more platforms in the future)
	if !strings.Contains(allChecksumsForVersion, actualHash) {
		return errors.Errorf("checksum verification failed: new file's checksum is: %s, but not found in the list of valid checksums:\n%s", actualHash, allChecksumsForVersion)
	}

	log.CliLogger.Tracef("Checksum verification succeeded")

	return nil
}

// UpdateBinary replaces the named binary at path with the desired version
func (c *client) UpdateBinary(cliName, version, path string, noVerify bool) error {
	downloadDir, err := c.fs.MkdirTemp("", cliName)
	if err != nil {
		return errors.Wrapf(err, errors.GetTempDirErrorMsg, cliName)
	}
	defer func() {
		err = c.fs.RemoveAll(downloadDir)
		if err != nil {
			log.CliLogger.Warnf("unable to clean up temp download dir %s: %s", downloadDir, err)
		}
	}()

	fmt.Fprintf(c.Out, "Downloading %s version %s...\n", cliName, version)
	startTime := c.clock.Now()

	newBin, bytes, err := c.Repository.DownloadVersion(cliName, version, downloadDir)
	if err != nil {
		return errors.Wrapf(err, errors.DownloadVersionErrorMsg, cliName, version, downloadDir)
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := c.clock.Now().Sub(startTime).Seconds()
	fmt.Fprintf(c.Out, "Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, timeSpent, mb/timeSpent)

	if !noVerify {
		if err := c.VerifyChecksum(newBin, cliName, version); err != nil {
			return errors.Wrapf(err, "checksum verification failed for new binary")
		}
	}

	// On Windows, we have to move the old binary out of the way first, then copy the new one into place,
	// because Windows doesn't support directly overwriting a running binary.
	// Note, this should _only_ be done on Windows; on unix platforms, cross-device moves can fail (e.g.
	// binary is on another device than the system tmp dir); but on such platforms we don't need to do moves anyway

	newPath := filepath.Join(filepath.Dir(path), cliName)

	if c.OS == "windows" {
		// The old version will get deleted automatically eventually as we put it in the system's or user's temp dir
		previousVersionBinary := filepath.Join(downloadDir, cliName+".old")
		err = c.fs.Move(path, previousVersionBinary)
		if err != nil {
			return errors.Wrapf(err, errors.MoveFileErrorMsg, path, previousVersionBinary)
		}
		err = c.copyFile(newBin, newPath)
		if err != nil {
			// If we moved the old binary out of the way but couldn't put the new one in place,
			// attempt to restore the old binary to where it was before bailing
			restoreErr := c.fs.Move(previousVersionBinary, path)
			if restoreErr != nil {
				// Warning: this is a bad case where the user will need to re-download the CLI.  However,
				// we shouldn't reach here since if the Move succeeded in one direction it's likely to work
				// in the opposite direction as well
				return errors.Wrapf(restoreErr, errors.MoveRestoreErrorMsg, previousVersionBinary, path)
			}
			return errors.Wrapf(err, errors.CopyErrorMsg, newBin, newPath)
		}
	} else {
		err = c.copyFile(newBin, newPath)
		if err != nil {
			return errors.Wrapf(err, errors.CopyErrorMsg, newBin, newPath)
		}
	}

	if err := c.fs.Chmod(newPath, 0755); err != nil {
		return errors.Wrapf(err, errors.ChmodErrorMsg, newPath)
	}

	// After updating `ccloud` to `confluent`, remove `ccloud`.
	if newPath != path {
		if err := c.fs.Remove(path); err != nil {
			return errors.Wrapf(err, "unable to remove %s", path)
		}
	}

	return nil
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

// copyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and removes
// the dst if it exists.
func (c *client) copyFile(src, dst string) error {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return nil
	}
	sf, err := c.fs.Open(cleanSrc)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := c.fs.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return err
	}
	df, err := c.fs.Create(cleanDst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = c.fs.Copy(df, sf)
	return err
}
