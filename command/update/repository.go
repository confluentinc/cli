package update

import "github.com/hashicorp/go-version"

type Repository interface {
	GetAvailableVersions(name string) (version.Collection, error)
	DownloadVersion(name, version, downloadDir string) (string, int64, error)
}
