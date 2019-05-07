package errors

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    string
		wantErr bool
		// Need to check the type is preserved or the type switch won't actually work.
		// This happens with "type Foo error" definitions. They just all hit the first switch case.
		wantErrType string // reflect.TypeOf().String()
	}{
		{
			name:    "static message",
			err:     ErrUnauthorized,
			want:    "You must login to access Confluent Cloud.",
			wantErr: true,
		},
		{
			name:        "dynamic message",
			err:         NewNotAuthenticatedError("some dynamic message"),
			want:        "some dynamic message",
			wantErr:     true,
			wantErrType: "errors.NotAuthenticatedError",
		},
		{
			name:        "dynamic message - not first case",
			err:         &UnconfiguredAPIKeyContextError{ClusterID: "cid", APIKey: "apikeyid"},
			want:        "please add API secret with 'api-key store apikeyid --cluster cid'",
			wantErr:     true,
			wantErrType: "*errors.UnconfiguredAPIKeyContextError",
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
				t.Errorf("HandleCommon() got: %s, want: %s", err, tt.want)
			}
			errType := reflect.TypeOf(err).String()
			if tt.wantErrType != "" && tt.wantErrType != errType {
				t.Errorf("HandleCommon() got type: %s, wantErrType: %s", errType, tt.wantErrType)
			}
		})
	}
}
