package update

import (
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRelevantReleases(t *testing.T) {
	a := &github.RepositoryRelease{TagName: toPtr("v4.0.0")}
	b := &github.RepositoryRelease{TagName: toPtr("v3.6.0")}
	c := &github.RepositoryRelease{TagName: toPtr("v3.5.2")}

	releases := []*github.RepositoryRelease{a, b, c}

	minorReleases := getRelevantReleases(releases, semver.MustParse("3.5.2"), false)
	assert.Equal(t, []*github.RepositoryRelease{b}, minorReleases)

	majorReleases := getRelevantReleases(releases, semver.MustParse("3.5.2"), true)
	assert.Equal(t, []*github.RepositoryRelease{b, a}, majorReleases)
}

func TestGetReleaseVersion(t *testing.T) {
	version, err := getReleaseVersion(&github.RepositoryRelease{TagName: toPtr("v3.6.0")})
	require.NoError(t, err)
	require.Equal(t, semver.MustParse("3.6.0"), version)
}

func toPtr(s string) *string {
	return &s
}

func TestFindChecksum(t *testing.T) {
	checksums := strings.Join([]string{
		"111111  windows",
		"222222  darwin",
		"333333  linux",
	}, "\n")

	checksum, err := findChecksum(checksums, "darwin")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x22, 0x22, 0x22}, checksum)

	checksum, err = findChecksum(checksums, "alpine")
	assert.NoError(t, err)
	assert.Nil(t, checksum)
}

func TestConvertBytesToMegabytes(t *testing.T) {
	require.Equal(t, 1.0, convertBytesToMegabytes(1048576))
}
