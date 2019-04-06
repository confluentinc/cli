package s3

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
)

// ObjectKey represents an S3 Key for a versioned package
type ObjectKey interface {
	ParseVersion(key string) (match bool, foundVersion *version.Version, err error)
	URLFor(name, version string) string
}

// VersionPrefixedKey is a version-prefixed S3 key with the format PREFIX/VERSION/NAME_VERSION_OS_ARCH
type VersionPrefixedKey struct {
	Prefix string
	Name   string
	// Optional char used to separate sections of the package name, defaults to "_"
	Separator string
	// @VisibleForTesting, defaults to runtime.GOOS and runtime.GOARCH
	goos   string
	goarch string
}

// NewVersionPrefixedKey returns a VersionPrefixedKey for a given S3 path prefix and binary name
func NewVersionPrefixedKey(prefix, name, sep string) *VersionPrefixedKey {
	if sep == "" {
		sep = "_"
	}
	return &VersionPrefixedKey{
		Prefix:    prefix,
		Name:      name,
		Separator: sep,
		goos:      runtime.GOOS,
		goarch:    runtime.GOARCH,
	}
}

func (p *VersionPrefixedKey) URLFor(name, version string) string {
	packageName := strings.Join([]string{name, version, p.goos, p.goarch}, p.Separator)
	return fmt.Sprintf("%s/%s/%s", p.Prefix, version, packageName)
}

func (p *VersionPrefixedKey) ParseVersion(key string) (bool, *version.Version, error) {
	split := strings.Split(key, p.Separator)

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
	if split[2] != p.goos {
		return false, nil, nil
	}

	// Skip binaries not for this Arch
	if split[3] != p.goarch {
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
