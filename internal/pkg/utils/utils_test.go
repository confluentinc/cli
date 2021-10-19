package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

const (
	resourceFlagName       = "resource"
	verboseFlagName        = "verbose"
	serviceAccountFlagName = "service-account"
	helpFlagName           = "help"
)

func TestContains(t *testing.T) {
	req := require.New(t)
	req.True(Contains([]string{"a"}, "a"))
}

func TestDoesNotContain(t *testing.T) {
	req := require.New(t)
	req.False(Contains([]string{}, "a"))
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

func TestReadConfigsFromFile(t *testing.T) {
	req := require.New(t)
	type testCase struct {
		config   string
		expected map[string]string
		wantErr  bool
	}
	tests := []testCase{
		{
			config: "key=val\n key2=val2 \n key3=val password=pass",
			expected: map[string]string{
				"key":  "val",
				"key2": "val2",
				"key3": "val password=pass",
			},
			wantErr: false,
		},
		{
			config:  "keyval\nkey2 = val2\n key3=val password=pass",
			wantErr: true,
		},
	}
	file1, err := ioutil.TempFile(os.TempDir(), "test")
	req.NoError(err)
	_, err = file1.Write([]byte(tests[0].config))
	req.NoError(err)
	defer os.Remove(file1.Name())
	out1, err := ReadConfigsFromFile(file1.Name())
	if tests[0].wantErr {
		req.Error(err)
	} else {
		req.NoError(err)
		validateConfigMap(req, tests[0].expected, out1)
	}

	file2, err := ioutil.TempFile(os.TempDir(), "test")
	req.NoError(err)
	_, err = file2.Write([]byte(tests[1].config))
	req.NoError(err)
	defer os.Remove(file2.Name())
	out2, err := ReadConfigsFromFile(file2.Name())
	if tests[1].wantErr {
		req.Error(err)
	} else {
		req.NoError(err)
		validateConfigMap(req, tests[1].expected, out2)
	}
}

func validateConfigMap(req *require.Assertions, expected map[string]string, out map[string]string) {
	req.Equal(len(expected), len(out))
	for k, v := range out {
		req.Equal(expected[k], v)
	}
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
