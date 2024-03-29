//go:generate mocker --prefix "" --dst mock/repository.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 repository.go Repository
package update

import (
	"github.com/hashicorp/go-version" // This "version" alias is require for go:generate go run github.com/travisjeffery/mocker/cmd/mocker to work
)

// Repository is a collection of versioned packages
type Repository interface {
	GetLatestMajorAndMinorVersion(name string, current *version.Version) (*version.Version, *version.Version, error)
	GetLatestReleaseNotesVersions(name, currentVersion string) (version.Collection, error)
	GetAvailableBinaryVersions(name string) (version.Collection, error)
	GetAvailableReleaseNotesVersions(name string) (version.Collection, error)
	DownloadVersion(name, version string) ([]byte, error)
	DownloadReleaseNotes(name, version string) (string, error)
	DownloadChecksums(name, version string) (string, error)
}
