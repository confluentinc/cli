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

	"github.com/confluentinc/cli/log"
)

type Params struct {
	Repository      Repository
	UpdateCheckFile string
	Logger          *log.Logger
}

type Client struct {
	*Params
}

// NewUpdateClient returns a client for updating CLI binaries
func NewUpdateClient(params *Params) (*Client, error) {
	return &Client{
		Params:  params,
	}, nil
}

// Check for new versions in s3 bucket
func (c *Client) CheckForUpdates(name string, currentVersion string) (updateAvailable bool, latestVersion string, err error) {
	availableVersions, err := c.Repository.GetAvailableVersions(name)
	if err != nil {
		return false, "", err
	}

	mostRecentVersion := availableVersions[len(availableVersions)-1]

	currVersion, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, "", fmt.Errorf("unable to parse %s version %s - %s", name, currentVersion, err)
	}

	if currVersion.LessThan(mostRecentVersion) {
		return true, mostRecentVersion.String(), nil
	}

	return false, currentVersion, nil
}


func (c *Client) PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool {
	if confirm && !isatty.IsTerminal(os.Stdout.Fd()) {
		c.Logger.Warn("disable confirm as stdout is not a tty")
		confirm = false
	}

	c.Logger.Printf("New version of %s is available", name)
	c.Logger.Printf("Current Version: %s", currVersion)
	c.Logger.Printf("Latest Version:  %s", latestVersion)

	if confirm {
		for {
			fmt.Print("Do you want to download an install this update? (y/n): ")

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

	return false
}

func (c *Client) UpdateBinary(name, version, path string) error {
	downloadDir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadDir)

	newBin, err := c.Repository.DownloadVersion(name, version, downloadDir)
	if err != nil {
		return err
	}

	fmt.Println(path)
	fmt.Println(newBin)
	//err = copyFile(newBin, path)
	//if err != nil {
	//	return err
	//}
	//
	//if err := os.Chmod(oldBin, 0755); err != nil {
	//	return err
	//}

	return nil
}


func (c *Client) TouchUpdateCheckFile() error {
	updateFile, err := homedir.Expand(LastCheckFile)
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
