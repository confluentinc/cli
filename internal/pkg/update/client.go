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
	"github.com/mattn/go-isatty"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type client struct {
	repository    Repository
	lastCheckFile string
	logger        *log.Logger
}

// NewClient returns a client for updating CLI binaries
func NewClient(repo Repository, lastCheckFile string, logger *log.Logger) Client {
	return &client{
		repository:    repo,
		lastCheckFile: lastCheckFile,
		logger:        logger,
	}
}

// CheckForUpdates checks for new versions in the repo
func (c *client) CheckForUpdates(name string, currentVersion string) (updateAvailable bool, latestVersion string, err error) {
	availableVersions, err := c.repository.GetAvailableVersions(name)
	if err != nil {
		return false, "", err
	}

	mostRecentVersion := availableVersions[len(availableVersions)-1]

	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		err = fmt.Errorf("unable to parse %s version %s - %s", name, currentVersion, err)
		return false, "", err
	}

	if currVersion.LessThan(mostRecentVersion) {
		return true, mostRecentVersion.String(), nil
	}

	return false, currentVersion, nil
}

// NotifyIfUpdateAvailable prints a message if an update is available
func (c *client) NotifyIfUpdateAvailable(name string, currentVersion string) error {
	updateAvailable, _, err := c.CheckForUpdates(name, currentVersion)
	if err != nil {
		return err
	}
	if updateAvailable {
		fmt.Fprintf(os.Stderr, "Updates are available for %s. To install them, please run:\n$ %s update\n\n", name, name)
	}
	return nil
}

// PromptToDownload displays an interactive CLI prompt to download the latest version
func (c *client) PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool {
	if confirm && !isatty.IsTerminal(os.Stdout.Fd()) {
		c.logger.Warn("disable confirm as stdout is not a tty")
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
	startTime := time.Now()

	newBin, bytes, err := c.repository.DownloadVersion(name, version, downloadDir)
	if err != nil {
		return err
	}

	mb := float64(bytes) / 1024.0 / 1024.0
	timeSpent := time.Since(startTime).Seconds()
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

func (c *client) TouchUpdateCheckFile() error {
	updateFile, err := homedir.Expand(c.lastCheckFile)
	if err != nil {
		return err
	}

	if _, err := os.Stat(updateFile); os.IsNotExist(err) {
		if f, err := os.Create(updateFile); err != nil {
			return err
		} else {
			f.Close()
		}
	} else if err := os.Chtimes(updateFile, time.Now(), time.Now()); err != nil {
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
