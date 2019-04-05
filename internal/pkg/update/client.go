package update

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atrox/homedir"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update/io"
	"github.com/hashicorp/go-version"
	"github.com/jonboulle/clockwork"
)

type client struct {
	*ClientParams
}

var _ Client = (*client)(nil)

// ClientParams are used to configure the update.Client
type ClientParams struct {
	Repository    Repository
	Logger        *log.Logger
	// Optional, if you wish to rate limit your update checks
	CheckFile     string
	// Optional, defaults to checking once every 24h
	CheckInterval time.Duration
	// Optional, defaults to the system clock
	Clock         clockwork.Clock
	// Optional, defaults to the OS filesystem
	FS            io.FileSystem
}

// NewClient returns a client for updating CLI binaries
func NewClient(params *ClientParams) *client {
	if params.CheckInterval == 0 {
		params.CheckInterval = 24 * time.Hour
	}
	if params.Clock == nil {
		params.Clock = clockwork.NewRealClock()
	}
	if params.FS == nil {
		params.FS = &io.RealFileSystem{}
	}
	return &client{
		ClientParams: params,
	}
}

// CheckForUpdates checks for new versions in the repo
func (c *client) CheckForUpdates(name string, currentVersion string, forceCheck bool) (updateAvailable bool, latestVersion string, err error) {
	shouldCheck, err := c.readCheckFile()
	if err != nil {
		return false, currentVersion, err
	}
	if !shouldCheck && !forceCheck{
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
	if confirm && !c.FS.IsTerminal(os.Stdout.Fd()) {
		c.Logger.Warn("disable confirm as stdout is not a tty")
		confirm = false
	}

	fmt.Printf("New version of %s is available\n", name)
	fmt.Printf("Current Version: %s\n", currVersion)
	fmt.Printf("Latest Version:  %s\n", latestVersion)

	if !confirm {
		return true
	}

	for {
		fmt.Print("Do you want to download and install this update? (y/n): ")

		reader := c.FS.NewBufferedReader(os.Stdin)
		choice, _ := reader.ReadString('\n')

		switch choice {
		case "yes", "y", "Y":
			return true
		case "no", "n", "N":
			return false
		default:
			fmt.Printf("%s is not a valid choice\n", choice)
			continue
		}
	}
}

// UpdateBinary replaces the named binary at path with the desired version
func (c *client) UpdateBinary(name, version, path string) error {
	downloadDir, err := c.FS.TempDir("", name)
	if err != nil {
		return errors.Wrapf(err, "unable to get temp dir for %s", name)
	}
	defer func() {
		err = c.FS.RemoveAll(downloadDir)
		c.Logger.Warnf("unable to clean up temp download dir %s: %s", downloadDir, err)
	}()

	fmt.Printf("Downloading %s version %s...\n", name, version)
	startTime := c.Clock.Now()

	newBin, bytes, err := c.Repository.DownloadVersion(name, version, downloadDir)
	if err != nil {
		return errors.Wrapf(err, "unable to download %s version %s to %s", name, version, downloadDir)
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := c.Clock.Now().Sub(startTime).Seconds()
	fmt.Printf("Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, timeSpent, mb/timeSpent)

	err = c.copyFile(newBin, path)
	if err != nil {
		return errors.Wrapf(err, "unable to copy %s to %s", newBin, path)
	}

	if err := c.FS.Chmod(path, 0755); err != nil {
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
	info, err := c.FS.Stat(updateFile)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	// if the file doesn't exist, check updates anyway -- indicates a new CLI install
	if os.IsNotExist(err) {
		return true, nil
	}
	// if the file was updated in the last (interval), don't check again
	if info.ModTime().After(c.Clock.Now().Add(-1 * c.CheckInterval)) {
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

	if _, err := c.FS.Stat(checkFile); os.IsNotExist(err) {
		if f, err := c.FS.Create(checkFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := c.FS.Chtimes(checkFile, c.Clock.Now(), c.Clock.Now()); err != nil {
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
	sf, err := c.FS.Open(cleanSrc)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := c.FS.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return err
	}
	df, err := c.FS.Create(cleanDst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = c.FS.Copy(df, sf)
	return err
}
