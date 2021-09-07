package utils

import (
	"fmt"
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
		&RegexTest{
			email:   "",
			matched: false,
		},
		&RegexTest{
			email:   "mtodzo@confluent.io",
			matched: true,
		},
		&RegexTest{
			email:   "m@t.t.com",
			matched: true,
		},
		&RegexTest{
			email:   "m@t",
			matched: true,
		},
		&RegexTest{
			email:   "google.com",
			matched: false,
		},
		&RegexTest{
			email:   "@images.google.com",
			matched: false,
		},
		&RegexTest{
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
	for _, t := range tests {
		file, err := ioutil.TempFile(os.TempDir(), "test")
		req.NoError(err)
		_, err = file.Write([]byte(t.config))
		req.NoError(err)
		out, err := ReadConfigsFromFile(file.Name())
		if t.wantErr {
			req.Error(err)
		} else {
			req.NoError(err)
			validateConfigMap(req, t.expected, out)
		}
		err = os.Remove(file.Name())
		req.NoError(err)
	}
}

func validateConfigMap(req *require.Assertions, expected map[string]string, out map[string]string) {
	req.Equal(len(expected), len(out))
	for k, v := range out {
		fmt.Println(k + " " + v)
		req.Equal(expected[k], v)
	}
}
