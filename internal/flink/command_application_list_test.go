package flink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildApplicationFilter(t *testing.T) {
	tests := []struct {
		name   string
		appn   string
		status string
		want   string
	}{
		{name: "empty", appn: "", status: "", want: ""},
		{name: "name only", appn: "my-app", status: "", want: "name=my-app"},
		{name: "name wildcard", appn: "my-app*", status: "", want: "name=my-app*"},
		{name: "status only", appn: "", status: "RUNNING", want: "state=RUNNING"},
		{name: "name and status", appn: "a*", status: "RUNNING", want: "name=a*,state=RUNNING"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, buildApplicationFilter(test.appn, test.status))
		})
	}
}
