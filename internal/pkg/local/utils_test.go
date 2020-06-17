package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildTabbedList(t *testing.T) {
	req := require.New(t)

	arr := []string{"a", "b"}
	out := "  a\n  b\n"
	req.Equal(out, BuildTabbedList(arr))
}

func TestExtractConfig(t *testing.T) {
	req := require.New(t)

	in := []byte("key1=val1\nkey2=val2\n#commented=val\n")

	out := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}

	req.Equal(out, ExtractConfig(in))
}

func TestVersionCmpBasic(t *testing.T) {
	req := require.New(t)

	var cmp int
	var err error

	cmp, err = versionCmp("1.0.0", "2.0.0")
	req.NoError(err)
	isNegative(req, cmp)

	cmp, err = versionCmp("1.0.0", "1.0.0")
	req.NoError(err)
	req.Equal(0, cmp)

	cmp, err = versionCmp("2.0.0", "1.0.0")
	req.NoError(err)
	isPositive(req, cmp)
}

func TestVersionCmpSamePrefix(t *testing.T) {
	req := require.New(t)

	cmp, err := versionCmp("1.0.1", "1.0.2")
	req.NoError(err)
	isNegative(req, cmp)
}

func TestVersionCmpDifferentLengths(t *testing.T) {
	req := require.New(t)

	var cmp int
	var err error

	cmp, err = versionCmp("1.0", "1.0.1")
	req.NoError(err)
	isNegative(req, cmp)

	cmp, err = versionCmp("1.0", "1.0.0")
	req.NoError(err)
	req.Equal(0, cmp)
}

func isNegative(req *require.Assertions, n int) {
	req.True(n < 0)
}

func isPositive(req *require.Assertions, n int) {
	req.True(n > 0)
}
