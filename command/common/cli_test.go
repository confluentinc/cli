package common

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/confluentinc/cli/shared"
	"github.com/spf13/cobra"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    string
		wantErr bool
	}{
		{
			name: "static message",
			err:  shared.ErrUnauthorized,
			want: "You must login to access Confluent Cloud.\n",
		},
		{
			name: "dynamic message",
			err:  shared.NotAuthenticatedError(fmt.Errorf("some dynamic message")),
			want: "some dynamic message\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOutput(buf)
			if err := HandleError(tt.err, cmd); (err != nil) != tt.wantErr {
				t.Errorf("HandleError() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buf.String() != tt.want {
				t.Errorf("HandleError() got = %v, want %v", buf.String(), tt.want)
			}
		})
	}
}
