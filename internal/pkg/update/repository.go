//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst mock/repository.go --pkg mock --selfpkg github.com/confluentinc/cli repository.go Repository
package update

import (
	"github.com/hashicorp/go-version" // This "version" alias is require for go:generate go run github.com/travisjeffery/mocker/cmd/mocker to work
)

// Repository is a collection of versioned packages
type Repository interface {
	GetLatestBinaryVersion(name string) (*version.Version, error)

	GetLatestReleaseNotesVersion() (*version.Version, error)

	GetAvailableBinaryVersions(name string) (version.Collection, error)

	GetAvailableReleaseNotesVersions() (version.Collection, error)

	// Downloads the versioned package to download dir to downloadDir.
	// Returns the full path to the downloaded package, the download size in bytes, or an error if one occurred.
	DownloadVersion(name, version, downloadDir string) (downloadPath string, downloadedBytes int64, err error)

	DownloadReleaseNotes(version string) (string, error)
}
