//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/client.go --pkg mock --selfpkg github.com/confluentinc/cli client.go Client
package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/atrox/homedir"
	"github.com/hashicorp/go-version"
	"github.com/jonboulle/clockwork"

	"github.com/confluentinc/cli/internal/pkg/errors"
	pio "github.com/confluentinc/cli/internal/pkg/io"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Client lets you check for updated application binaries and install them if desired
type Client interface {
	CheckForUpdates(name string, currentVersion string, forceCheck bool) (updateAvailable bool, latestVersion string, err error)
	PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool
	UpdateBinary(name, version, path string) error
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
	Logger     *log.Logger
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
func (c *client) CheckForUpdates(name string, currentVersion string, forceCheck bool) (updateAvailable bool, latestVersion string, err error) {
	if c.DisableCheck {
		return false, currentVersion, nil
	}
	shouldCheck, err := c.readCheckFile()
	if err != nil {
		return false, currentVersion, err
	}
	if !shouldCheck && !forceCheck {
		return false, currentVersion, nil
	}

	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		err = errors.Wrapf(err, "unable to parse %s version %s", name, currentVersion)
		return false, currentVersion, err
	}

	availableVersions, err := c.Repository.GetAvailableVersions(name)
	if err != nil {
		return false, currentVersion, errors.Wrapf(err, "unable to get available versions")
	}

	if err := c.touchCheckFile(); err != nil {
		return false, currentVersion, errors.Wrapf(err, "unable to touch last check file")
	}

	mostRecentVersion := availableVersions[len(availableVersions)-1]
	if currVersion.LessThan(mostRecentVersion) {
		return true, mostRecentVersion.Original(), nil
	}

	return false, currentVersion, nil
}

// PromptToDownload displays an interactive CLI prompt to download the latest version
func (c *client) PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool {
	if confirm && !c.fs.IsTerminal(c.Out.Fd()) {
		c.Logger.Warn("disable confirm as stdout is not a tty")
		confirm = false
	}

	fmt.Fprintf(c.Out, "New version of %s is available\n", name)
	fmt.Fprintf(c.Out, "Current Version: %s\n", currVersion)
	fmt.Fprintf(c.Out, "Latest Version:  %s\n", latestVersion)

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
func (c *client) UpdateBinary(name, version, path string) error {
	downloadDir, err := c.fs.TempDir("", name)
	if err != nil {
		return errors.Wrapf(err, "unable to get temp dir for %s", name)
	}
	defer func() {
		err = c.fs.RemoveAll(downloadDir)
		if err != nil {
			c.Logger.Warnf("unable to clean up temp download dir %s: %s", downloadDir, err)
		}
	}()

	fmt.Fprintf(c.Out, "Downloading %s version %s...\n", name, version)
	startTime := c.clock.Now()

	newBin, bytes, err := c.Repository.DownloadVersion(name, version, downloadDir)
	if err != nil {
		return errors.Wrapf(err, "unable to download %s version %s to %s", name, version, downloadDir)
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := c.clock.Now().Sub(startTime).Seconds()
	fmt.Fprintf(c.Out, "Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, timeSpent, mb/timeSpent)

	// On Windows, we have to move the old binary out of the way first, then copy the new one into place,
	// because Windows doesn't support directly overwriting a running binary.
	// Note, this should _only_ be done on Windows; on unix platforms, cross-devices moves can fail (e.g.
	// binary is on another device than the system tmp dir); but on such platforms we don't need to do moves anyway

	if c.OS == "windows" {
		// The old version will get deleted automatically eventually as we put it in the system's or user's temp dir
		previousVersionBinary := filepath.Join(downloadDir, name+".old")
		err = c.fs.Move(path, previousVersionBinary)
		if err != nil {
			return errors.Wrapf(err, "unable to move %s to %s", path, previousVersionBinary)
		}
		err = c.copyFile(newBin, path)
		if err != nil {
			// If we moved the old binary out of the way but couldn't put the new one in place,
			// attempt to restore the old binary to where it was before bailing
			restoreErr := c.fs.Move(previousVersionBinary, path)
			if restoreErr != nil {
				// Warning: this is a bad case where the user will need to re-download the CLI.  However,
				// we shouldn't reach here since if the Move succeeded in one direction it's likely to work
				// in the opposite direction as well
				return errors.Wrapf(restoreErr, "unable to move (restore) %s to %s", previousVersionBinary, path)
			}

			return errors.Wrapf(err, "unable to copy %s to %s", newBin, path)
		}
	} else {
		err = c.copyFile(newBin, path)
		if err != nil {
			return errors.Wrapf(err, "unable to copy %s to %s", newBin, path)
		}
	}

	if err := c.fs.Chmod(path, 0755); err != nil {
		return errors.Wrapf(err, "unable to chmod 0755 %s", path)
	}

	return nil
}

func (c *client) readCheckFile() (shouldCheck bool, err error) {
	// If CheckFile is not provided, then we'll always perform the check
	if c.CheckFile == "" {
		return true, nil
	}
	updateFile, err := homedir.Expand(c.CheckFile)
	if err != nil {
		return false, err
	}
	info, err := c.fs.Stat(updateFile)
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
	checkFile, err := homedir.Expand(c.CheckFile)
	if err != nil {
		return err
	}

	if _, err := c.fs.Stat(checkFile); os.IsNotExist(err) {
		if f, err := c.fs.Create(checkFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := c.fs.Chtimes(checkFile, c.clock.Now(), c.clock.Now()); err != nil {
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
