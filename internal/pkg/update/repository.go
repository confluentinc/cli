//go:generate mocker --prefix "" --dst mock/repository.go --pkg mock --selfpkg github.com/confluentinc/cli repository.go Repository
package update

import version "github.com/hashicorp/go-version"

// Repository is a collection of versioned packages
type Repository interface {
	// Returns a collection of versions for the named package, or an error if one occurred.
	GetAvailableVersions(name string) (version.Collection, error)

	// Downloads the versioned package to download dir to downloadDir.
	// Returns the full path to the downloaded package, the download size in bytes, or an error if one occurred.
	DownloadVersion(name, version, downloadDir string) (string, int64, error)
}

// Client lets you check for updated application binaries and install them if desired
type Client interface {
	CheckForUpdates(name string, currentVersion string, forceCheck bool) (updateAvailable bool, latestVersion string, err error)
	PromptToDownload(name, currVersion, latestVersion string, confirm bool) bool
	UpdateBinary(name, version, path string) error
}
