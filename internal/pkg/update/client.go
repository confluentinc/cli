package update

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/atrox/homedir"
	"github.com/hashicorp/go-version"
	"github.com/jonboulle/clockwork"
	"github.com/mattn/go-isatty"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
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
}

// NewClient returns a client for updating CLI binaries
func NewClient(params *ClientParams) *client {
	if params.CheckInterval == 0 {
		params.CheckInterval = 24 * time.Hour
	}
	if params.Clock == nil {
		params.Clock = clockwork.NewRealClock()
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
	if confirm && !isatty.IsTerminal(os.Stdout.Fd()) {
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

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		choice := string([]byte(input)[0])
		switch choice {
		case "y":
			return true
		case "n":
			return false
		default:
			fmt.Printf("%s is not a valid choice", choice)
			continue
		}
	}
}

// UpdateBinary replaces the named binary at path with the desired version
func (c *client) UpdateBinary(name, version, path string) error {
	downloadDir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadDir)

	fmt.Printf("Downloading %s version %s...\n", name, version)
	startTime := c.Clock.Now()

	newBin, bytes, err := c.Repository.DownloadVersion(name, version, downloadDir)
	if err != nil {
		return err
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := c.Clock.Now().Sub(startTime).Seconds()
	fmt.Printf("Done. Downloaded %.2f MB in %.0f seconds. (%.2f MB/s)\n", mb, timeSpent, mb/timeSpent)

	err = copyFile(newBin, path)
	if err != nil {
		return err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
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
	info, err := os.Stat(updateFile)
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

	if _, err := os.Stat(checkFile); os.IsNotExist(err) {
		if f, err := os.Create(checkFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := os.Chtimes(checkFile, c.Clock.Now(), c.Clock.Now()); err != nil {
		return err
	}
	return nil
}

// copyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and removes
// the dst if it exists.
func copyFile(src, dst string) error {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	return err
}
