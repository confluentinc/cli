package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
			email:   "m12asdf/@t.com",
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
	}
	for _, test := range tests {
		require.Equal(t, test.matched, validateEmail(test.email))
	}
}
