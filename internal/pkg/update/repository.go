//go:generate mocker --prefix "" --dst mock/repository.go --pkg mock --selfpkg github.com/confluentinc/cli repository.go Repository
package update

import (
	version "github.com/hashicorp/go-version" // This "version" alias is require for go:generate mocker to work
)

// Repository is a collection of versioned packages
type Repository interface {
	// Returns a collection of versions for the named package, or an error if one occurred.
	GetAvailableVersions(name string) (version.Collection, error)

	// Downloads the versioned package to download dir to downloadDir.
	// Returns the full path to the downloaded package, the download size in bytes, or an error if one occurred.
	DownloadVersion(name, version, downloadDir string) (downloadPath string, downloadedBytes int64, err error)
}
