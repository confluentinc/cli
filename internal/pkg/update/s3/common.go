package s3

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
)

// Validates whether an S3 Key represents an installable package and parses the package version
type KeyParser interface {
	Validate(key string) (skip bool, foundVersion *version.Version, err error)
}

// Parses version-prefixed package S3 keys with the format PREFIX/VERSION/NAME_VERSION_OS_ARCH
type VersionPrefixedKeyParser struct {
	Prefix string
	Name   string
	// @VisibleForTesting, defaults to runtime.GOOS and runtime.GOARCH
	OS string
	Arch string
}

func (p *VersionPrefixedKeyParser) Validate(key string) (skip bool, foundVersion *version.Version, err error) {
	if p.OS == "" {
		p.OS = runtime.GOOS
	}
	if p.Arch == "" {
		p.Arch = runtime.GOARCH
	}

	split := strings.Split(key, "_")

	// Skip files that don't match our naming standards for binaries
	if len(split) != 4 {
		return false, nil, nil
	}

	// Skip objects from other directories
	if !strings.HasPrefix(split[0], p.Prefix) {
		return false, nil, nil
	}

	// Skip binaries other than the requested one
	if !strings.HasSuffix(split[0], p.Name) {
		return false, nil, nil
	}

	// Skip binaries not for this OS
	if split[2] != runtime.GOOS {
		return false, nil, nil
	}

	// Skip binaries not for this Arch
	if split[3] != runtime.GOARCH {
		return false, nil, nil
	}

	// Skip snapshot and dirty versions (which shouldn't be published, but accidents happen)
	if strings.Contains(split[1], "SNAPSHOT") {
		return false, nil, nil
	}
	if strings.Contains(split[1], "dirty") {
		return false, nil, nil
	}

	ver, err := version.NewSemver(split[1])
	if err != nil {
		return false, nil, fmt.Errorf("unable to parse %s version - %s", p.Name, err)
	}
	return true, ver, nil
}
