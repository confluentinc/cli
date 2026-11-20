package flink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeApplicationFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		want   string
	}{
		{name: "empty", filter: "", want: ""},
		{name: "name is not folded", filter: "name=My-App*", want: "name=My-App*"},
		{name: "state lower-cased is folded", filter: "state=running", want: "state=RUNNING"},
		{name: "state already upper", filter: "state=RUNNING", want: "state=RUNNING"},
		{name: "state key case-insensitive", filter: "State=running", want: "state=RUNNING"},
		{name: "name and state", filter: "name=My-App*,state=reconciling", want: "name=My-App*,state=RECONCILING"},
		{name: "unknown key passed through", filter: "foo=Bar", want: "foo=Bar"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, normalizeApplicationFilter(test.filter))
		})
	}
}
