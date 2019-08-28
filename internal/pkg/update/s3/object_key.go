//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/object_key.go --pkg mock --selfpkg github.com/confluentinc/cli object_key.go ObjectKey
package s3

import (
	"fmt"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version" // This "version" alias is require for go:generate go run github.com/travisjeffery/mocker/cmd/mocker to work
)

// ObjectKey represents an S3 Key for a versioned package
type ObjectKey interface {
	ParseVersion(key, name string) (match bool, foundVersion *version.Version, err error)
	URLFor(name, version string) string
}

// PrefixedKey is a prefixed S3 key
type PrefixedKey struct {
	Prefix string
	// Whether the S3 key has a VERSION prefix in the path before the package name
	PrefixVersion bool
	// Character used to separate sections of the package name
	Separator string
	// @VisibleForTesting, defaults to runtime.GOOS and runtime.GOARCH
	goos   string
	goarch string
}

// NewPrefixedKey returns a PrefixedKey for a given S3 path prefix and binary name.
//
// You must also specify whether the S3 key is prefixed with the version as well as
// the separator for parts of the package name (shown below with "_" separators).
//
// If prefixVersion, s3 key format is PREFIX/VERSION/PACKAGE_VERSION_OS_ARCH
//        otherwise, s3 key format is PREFIX/PACKAGE_VERSION_OS_ARCH
//
// Prefix may be an empty string. An error will be returned if sep is empty or a space.
func NewPrefixedKey(prefix, sep string, prefixVersion bool) (*PrefixedKey, error) {
	if sep == "" || sep == " " {
		return nil, fmt.Errorf("sep must be a non-empty string")
	}
	return &PrefixedKey{
		Prefix:        prefix,
		PrefixVersion: prefixVersion,
		Separator:     sep,
		goos:          runtime.GOOS,
		goarch:        runtime.GOARCH,
	}, nil
}

func (p *PrefixedKey) URLFor(name, version string) string {
	packageName := strings.Join([]string{name, version, p.goos, p.goarch}, p.Separator)
	if p.goos == "windows" {
		packageName += ".exe"
	}
	prefix := p.Prefix
	if p.Prefix != "" {
		prefix += "/"
	}
	if p.PrefixVersion {
		return fmt.Sprintf("%s%s/%s", prefix, version, packageName)
	} else {
		return fmt.Sprintf("%s%s", prefix, packageName)
	}
}

func (p *PrefixedKey) ParseVersion(key, name string) (match bool, foundVersion *version.Version, err error) {
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
	if !strings.HasSuffix(split[0], name) {
		return false, nil, nil
	}

	// Skip binaries without the right file extension
	if p.goos == "windows" {
		if !strings.HasSuffix(split[3], ".exe") {
			return false, nil, nil
		}
		split[3] = split[3][0 : len(split[3])-len(".exe")]
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

	// Skip if version is out of sync (which shouldn't happen, but, again, accidents happen)
	if p.PrefixVersion && !strings.Contains(split[0], "/"+split[1]+"/") {
		return false, nil, nil
	}

	ver, err := version.NewSemver(split[1])
	if err != nil {
		return false, nil, fmt.Errorf("unable to parse %s version - %s", name, err)
	}
	return true, ver, nil
}
