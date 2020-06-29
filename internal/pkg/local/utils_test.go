package local

import (
	"testing"

	"github.com/spf13/pflag"
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

func TestContains(t *testing.T) {
	req := require.New(t)
	req.True(Contains([]string{"a"}, "a"))
}

func TestDoesNotContain(t *testing.T) {
	req := require.New(t)
	req.False(Contains([]string{}, "a"))
}

func TestCollectFlags(t *testing.T) {
	req := require.New(t)

	flags := pflag.NewFlagSet("", pflag.ExitOnError)
	flags.Bool("bool-skip", false, "")
	flags.Bool("bool-use", true, "")
	flags.Int("int-skip", 0, "")
	flags.Int("int-use", 1, "")
	flags.String("string-skip", "", "")
	flags.String("string-use", "example", "")

	defaults := map[string]interface{}{
		"bool-skip":   false,
		"bool-use":    false,
		"int-skip":    0,
		"int-use":     0,
		"string-skip": "",
		"string-use":  "",
	}

	args, err := CollectFlags(flags, defaults)
	req.NoError(err)
	req.ElementsMatch([]string{"--bool-use", "--int-use", "1", "--string-use", "example"}, args)
}
