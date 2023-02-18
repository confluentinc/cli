package releasenotes

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	r := New()
	require.Empty(t, r.major)
	require.Empty(t, r.minor)
	require.Empty(t, r.patch)
}

func TestNewFromBody(t *testing.T) {
	r := NewFromBody(strings.Join([]string{
		"Release Notes",
		"-------------",
		"<!--",
		"If this PR introduces any user-facing changes, please document them below. You may delete or ignore any unused sections.",
		"Please match the style of previous release notes: https://docs.confluent.io/confluent-cli/current/release-notes.html",
		"-->",
		"",
		"Breaking Changes",
		"- <PLACEHOLDER>",
		"",
		"New Features",
		"- Feature 1",
		"- Feature 2",
		"",
		"Bug Fixes",
		"- <PLACEHOLDER>",
		"- Bug Fix 1",
		"",
		"Section",
		"-------",
		"- Ignore this bullet point",
	}, "\n"))

	require.Empty(t, r.major)
	require.ElementsMatch(t, r.minor, []string{"Feature 1", "Feature 2"})
	require.ElementsMatch(t, r.patch, []string{"Bug Fix 1"})
}

func TestMerge(t *testing.T) {
	a := &ReleaseNotes{
		major: []string{"A"},
		minor: []string{"A"},
		patch: []string{"A"},
	}
	b := &ReleaseNotes{
		major: []string{"B"},
		minor: []string{"B"},
		patch: []string{"B"},
	}

	a.Merge(b)

	require.ElementsMatch(t, a.major, []string{"A", "B"})
	require.ElementsMatch(t, a.minor, []string{"A", "B"})
	require.ElementsMatch(t, a.patch, []string{"A", "B"})
}

func TestGetBump_Major(t *testing.T) {
	r := &ReleaseNotes{
		major: []string{""},
		minor: []string{""},
		patch: []string{""},
	}
	bump, err := r.GetBump()
	require.NoError(t, err)
	require.Equal(t, "major", bump)
}

func TestGetBump_Minor(t *testing.T) {
	r := &ReleaseNotes{
		major: []string{},
		minor: []string{""},
		patch: []string{""},
	}
	bump, err := r.GetBump()
	require.NoError(t, err)
	require.Equal(t, "minor", bump)
}

func TestGetBump_Patch(t *testing.T) {
	r := &ReleaseNotes{
		major: []string{},
		minor: []string{},
		patch: []string{""},
	}
	bump, err := r.GetBump()
	require.NoError(t, err)
	require.Equal(t, "patch", bump)
}

func TestGetBump_ErrorNoUpdates(t *testing.T) {
	r := &ReleaseNotes{
		major: []string{},
		minor: []string{},
		patch: []string{},
	}
	_, err := r.GetBump()
	require.Error(t, err)
}

func TestBumpVersion(t *testing.T) {
	v := version.Must(version.NewSemver("v1.1.1"))
	assert.Equal(t, bumpVersion(v, "major"), "v2.0.0")
	assert.Equal(t, bumpVersion(v, "minor"), "v1.2.0")
	assert.Equal(t, bumpVersion(v, "patch"), "v1.1.2")
}
