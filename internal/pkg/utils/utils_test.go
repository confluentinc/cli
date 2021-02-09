package utils

import (
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
