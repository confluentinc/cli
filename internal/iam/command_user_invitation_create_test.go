package iam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateEmail(t *testing.T) {
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
		require.Equal(t, test.matched, validateEmail(test.email))
	}
}
