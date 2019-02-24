package s3

import "github.com/hashicorp/go-version"

type PublicRepo struct {}

func (r *PublicRepo) GetAvailableVersions(name string) (version.Collection, error) {
	return nil, nil
}

func (r *PublicRepo) DownloadVersion(name, version, downloadDir string) (string, error) {
	return "", nil
}
