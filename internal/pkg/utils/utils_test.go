package utils

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/types"
)

const (
	resourceFlagName       = "resource"
	verboseFlagName        = "verbose"
	serviceAccountFlagName = "service-account"
	helpFlagName           = "help"
)

func TestContains(t *testing.T) {
	req := require.New(t)
	req.True(types.Contains([]string{"a"}, "a"))
}

func TestDoesNotContain(t *testing.T) {
	req := require.New(t)
	req.False(types.Contains([]string{}, "a"))
}

func TestDoesPathExist(t *testing.T) {
	t.Run("DoesPathExist: empty path returns false", func(t *testing.T) {
		req := require.New(t)
		valid := DoesPathExist("")
		req.False(valid)
	})
}

func TestLoadPropertiesFile(t *testing.T) {
	t.Run("LoadPropertiesFile: empty path yields error", func(t *testing.T) {
		req := require.New(t)
		_, err := LoadPropertiesFile("")
		req.Error(err)
	})
}

func TestUserInviteEmailRegex(t *testing.T) {
	type RegexTest struct {
		email   string
		matched bool
	}
	tests := []*RegexTest{
		{
			email:   "",
			matched: false,
		},
		{
			email:   "mtodzo@confluent.io",
			matched: true,
		},
		{
			email:   "m@t.t.com",
			matched: true,
		},
		{
			email:   "m@t",
			matched: true,
		},
		{
			email:   "google.com",
			matched: false,
		},
		{
			email:   "@images.google.com",
			matched: false,
		},
		{
			email:   "david.hyde+cli@confluent.io",
			matched: true,
		},
	}
	for _, test := range tests {
		require.Equal(t, test.matched, ValidateEmail(test.email))
	}
}

func TestIsFlagArg(t *testing.T) {
	type testCase struct {
		arg    string
		isFlag bool
	}

	tests := []*testCase{
		{
			arg:    "--resource",
			isFlag: true,
		},
		{
			arg:    "-o",
			isFlag: true,
		},
		{
			arg:    "-vvv",
			isFlag: true,
		},
		{
			arg:    "bob",
			isFlag: false,
		},
	}

	for _, test := range tests {
		got := IsFlagArg(test.arg)
		if test.isFlag != got {
			t.Errorf("Arg: %s, Expected: %t, Got: %t", test.arg, test.isFlag, got)
		}
	}
}

func TestIsFlagWithArg(t *testing.T) {
	flagMap := getFlagMap()
	type testCase struct {
		flag     *pflag.Flag
		expected bool
	}

	tests := []*testCase{
		{
			flag:     flagMap[resourceFlagName],
			expected: true,
		},
		{
			flag:     flagMap[serviceAccountFlagName],
			expected: true,
		},
		{
			flag:     flagMap[helpFlagName],
			expected: false,
		},
		{
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
	}

	for _, test := range tests {
		got := IsFlagWithArg(test.flag)
		if test.expected != got {
			t.Errorf("Flag: %s, Flag Type: %s, Expected: %t, Got: %t",
				test.flag.Name, test.flag.Value.Type(), test.expected, got)
		}
	}
}

func TestIsShorthandCountFlag(t *testing.T) {
	flagMap := getFlagMap()
	type testCase struct {
		arg      string
		flag     *pflag.Flag
		expected bool
	}

	tests := []*testCase{
		{
			arg:      "-v",
			flag:     flagMap[verboseFlagName],
			expected: true,
		},
		{
			arg:      "-vvv",
			flag:     flagMap[verboseFlagName],
			expected: true,
		},
		{
			arg:      "--verbose",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "--v",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "--vvv",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "v",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "vvv",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "verbose",
			flag:     flagMap[verboseFlagName],
			expected: false,
		},
		{
			arg:      "--verbose",
			flag:     flagMap[serviceAccountFlagName],
			expected: false,
		},
		{
			arg:      "--service-account",
			flag:     flagMap[serviceAccountFlagName],
			expected: false,
		},
	}

	for _, test := range tests {
		got := IsShorthandCountFlag(test.flag, test.arg)
		if test.expected != got {
			t.Errorf("Arg: %s, Flag: %s, Expected: %t, Got: %t", test.arg, test.flag.Name, test.expected, got)
		}
	}
}

func TestAbbreviate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{
			input:    "helloooooo",
			maxLen:   3,
			expected: "hel...",
		},
		{
			input:    "helloooooo",
			maxLen:   50,
			expected: "helloooooo",
		},
		{
			input:    "hi",
			maxLen:   2,
			expected: "hi",
		},
	}
	for _, tt := range tests {
		out := Abbreviate(tt.input, tt.maxLen)
		require.Equal(t, tt.expected, out)
	}
}

func getFlagMap() map[string]*pflag.Flag {
	cmd := &cobra.Command{
		Use: "cmd",
	}
	cmd.Flags().Int32(serviceAccountFlagName, 0, "The service account ID to filter by.")
	cmd.Flags().CountP(verboseFlagName, "v", "Increase verbosity")
	cmd.Flags().String(resourceFlagName, "", "Resource ID.")
	cmd.Flags().BoolP(helpFlagName, "h", false, "help")
	flagMap := make(map[string]*pflag.Flag)

	addToMap := func(flag *pflag.Flag) {
		flagMap[flag.Name] = flag
	}
	cmd.LocalFlags().VisitAll(addToMap)
	return flagMap
}

func TestCropString(t *testing.T) {
	for _, tt := range []struct {
		s       string
		n       int
		cropped string
	}{
		{"ABCDE", 4, "A..."},
		{"ABCDE", 5, "ABCDE"},
		{"ABCDE", 8, "ABCDE"},
	} {
		require.Equal(t, tt.cropped, CropString(tt.s, tt.n))
	}
}

func TestArrayToCommaDelimitedString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{},
			expected: "",
		},
		{
			input:    []string{"val1"},
			expected: `"val1"`,
		},
		{
			input:    []string{"val1", "val2"},
			expected: `"val1" or "val2"`,
		},
		{
			input:    []string{"val1", "val2", "val3"},
			expected: `"val1", "val2", or "val3"`,
		},
	}
	for _, tt := range tests {
		out := ArrayToCommaDelimitedString(tt.input)
		require.Equal(t, tt.expected, out)
	}
}
