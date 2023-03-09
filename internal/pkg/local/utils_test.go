package local

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestBuildTabbedList(t *testing.T) {
	t.Parallel()

	req := require.New(t)

	arr := []string{"a", "b"}
	out := "  a\n  b\n"
	req.Equal(out, BuildTabbedList(arr))
}

func TestExtractConfig(t *testing.T) {
	t.Parallel()

	req := require.New(t)

	in := []byte("key1=val1\nkey2=val2\n#commented=val\n")

	out := map[string]any{
		"key1": "val1",
		"key2": "val2",
	}

	req.Equal(out, ExtractConfig(in))
}

func TestCollectFlags(t *testing.T) {
	t.Parallel()

	req := require.New(t)

	flags := pflag.NewFlagSet("", pflag.ExitOnError)
	flags.Bool("bool-skip", false, "")
	flags.Bool("bool-use", true, "")
	flags.Int("int-skip", 0, "")
	flags.Int("int-use", 1, "")
	flags.String("string-skip", "", "")
	flags.String("string-use", "example", "")
	flags.StringSlice("string-array-skip", []string{}, "")
	flags.StringSlice("string-array-use", []string{"A", "B"}, "")

	defaults := map[string]any{
		"bool-skip":         false,
		"bool-use":          false,
		"int-skip":          0,
		"int-use":           0,
		"string-skip":       "",
		"string-use":        "",
		"string-array-skip": []string{},
		"string-array-use":  []string{},
	}

	args, err := CollectFlags(flags, defaults)
	req.NoError(err)
	req.ElementsMatch(
		[]string{
			"--bool-use",
			"--int-use", "1",
			"--string-use", "example",
			"--string-array-use", "A", "--string-array-use", "B",
		},
		args,
	)
}
