package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name       string
		argsString string
		want       []string
	}{
		{
			name:       "basic string",
			argsString: `ccloud api-key create --resource lkc-123 --description hello`,
			want:       []string{"ccloud", "api-key", "create", "--resource", "lkc-123", "--description", "hello"},
		},
		{
			name:       "string arg with space at the end",
			argsString: `ccloud api-key create --resource lkc-123 --description "hello world"`,
			want:       []string{"ccloud", "api-key", "create", "--resource", "lkc-123", "--description", "hello world"},
		},
		{
			name:       "space in between double quotes",
			argsString: `ccloud api-key create --description "hello world" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", "hello world", "--resource", "lkc-123"},
		},
		{
			name:       "space in between single quotes",
			argsString: `ccloud api-key create --description 'hello world' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", "hello world", "--resource", "lkc-123"},
		},
		{
			name:       "escape double quotes pair inside of double quotes",
			argsString: `ccloud api-key create --description "escape \"quotes\" string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \"quotes\" string`, "--resource", "lkc-123"},
		},
		{
			name:       "escape double quotes pair inside of single quotes",
			argsString: `ccloud api-key create --description 'escape \"quotes\" string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \"quotes\" string`, "--resource", "lkc-123"},
		},
		{
			name:       "escape single quotes pair inside of double quotes",
			argsString: `ccloud api-key create --description "escape \'quotes\' string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \'quotes\' string`, "--resource", "lkc-123"},
		},
		{
			name:       "escape single quotes pair inside of single quotes",
			argsString: `ccloud api-key create --description 'escape \'quotes\' string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \'quotes\' string`, "--resource", "lkc-123"},
		},
		{
			name:       "one escape double quotes inside of double quotes",
			argsString: `ccloud api-key create --description "escape \"quotes string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \"quotes string`, "--resource", "lkc-123"},
		},
		{
			name:       "one escape double quotes inside of single quotes",
			argsString: `ccloud api-key create --description 'escape \"quotes string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \"quotes string`, "--resource", "lkc-123"},
		},
		{
			name:       "one escape single quotes inside of double quotes",
			argsString: `ccloud api-key create --description "escape \'quotes string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \'quotes string`, "--resource", "lkc-123"},
		},
		{
			name:       "one escape single quotes inside of single quotes",
			argsString: `ccloud api-key create --description 'escape \'quotes string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape \'quotes string`, "--resource", "lkc-123"},
		},
		{
			name:       "single quotes pair inside of double quotes",
			argsString: `ccloud api-key create --description "escape 'quotes' string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape 'quotes' string`, "--resource", "lkc-123"},
		},
		{
			name:       "one single quotes inside of double quotes",
			argsString: `ccloud api-key create --description "escape 'quotes string" --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", `escape 'quotes string`, "--resource", "lkc-123"},
		},
		{
			name:       "double quotes pair inside of single quotes",
			argsString: `ccloud api-key create --description 'escape "quotes" string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", "escape \"quotes\" string", "--resource", "lkc-123"},
		},
		{
			name:       "one double quotes inside of single quotes",
			argsString: `ccloud api-key create --description 'escape "quotes string' --resource lkc-123`,
			want:       []string{"ccloud", "api-key", "create", "--description", "escape \"quotes string", "--resource", "lkc-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitArgs(tt.argsString)
			if !TestEq(tt.want, got) {
				t.Errorf("SplitArgs got = %s, want %s", got, tt.want)
				for _, s := range got {
					fmt.Println(s)
				}
			}
		})
	}
}
