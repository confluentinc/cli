package errors

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	"github.com/stretchr/testify/require"
)

var (
	wantSuggestionsMsgFormat = `
Suggestions:
    %s
`
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
			err:     &NotLoggedInError{},
			want:    NotLoggedInErrorMsg,
			wantErr: true,
		},
		{
			name:    "dynamic message",
			err:     &UnconfiguredAPISecretError{APIKey: "MYKEY", ClusterID: "lkc-mine"},
			want:    fmt.Sprintf(NoAPISecretStoredErrorMsg, "MYKEY", "lkc-mine"),
			wantErr: true,
		},
		{
			name:    "mds v1 backend error",
			err:     mdsv1.GenericOpenAPIError{},
			want:    fmt.Sprintf(GenericOpenAPIErrorMsg, "", ""),
			wantErr: true,
		},
		{
			name:    "mds v2 backend error",
			err:     mdsv2alpha1.GenericOpenAPIError{},
			want:    fmt.Sprintf(GenericOpenAPIErrorMsg, "", ""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = HandleCommon(tt.err); (err != nil) != tt.wantErr {
				t.Errorf("HandleCommon()\nerror: %v\nwantErr: %v", err, tt.wantErr)
			}
			if err.Error() != tt.want {
				t.Errorf("HandleCommon()\ngot: %s\nwant: %s", err, tt.want)
			}
		})
	}
}

func TestSuggestionsMessage(t *testing.T) {
	errorMessage := "im an error hi"
	suggestionsMessage := "This is a suggestion"
	err := NewErrorWithSuggestions(errorMessage, suggestionsMessage)
	var b bytes.Buffer
	DisplaySuggestionsMessage(err, &b)
	out := b.String()
	wantSuggestionsMsg := fmt.Sprintf(wantSuggestionsMsgFormat, suggestionsMessage)
	require.Equal(t, wantSuggestionsMsg, out)
}
