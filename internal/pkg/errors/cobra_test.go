package errors

import (
	"fmt"
	"testing"

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
			name:    "static message",
			err:     ErrUnauthorized,
			want:    "You must login to access Confluent Cloud.",
			wantErr: true,
		},
		{
			name:    "dynamic message",
			err:     NotAuthenticatedError(fmt.Errorf("some dynamic message")),
			want:    "some dynamic message",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			var err error
			if err = HandleCommon(tt.err, cmd); (err != nil) != tt.wantErr {
				t.Errorf("HandleCommon() error: %v, wantErr: %v", err, tt.wantErr)
			}
			if err.Error() != tt.want {
				t.Errorf("HandleCommon() got = %s, want: %s", err, tt.want)
			}
		})
	}
}
