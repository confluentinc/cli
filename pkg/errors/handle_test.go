package errors

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
)

var wantSuggestionsMsgFormat = `
Suggestions:
    %s
`

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
			want:    fmt.Sprintf(GenericOpenApiErrorMsg, "", ""),
			wantErr: true,
		},
		{
			name:    "mds v2 backend error",
			err:     mdsv2alpha1.GenericOpenAPIError{},
			want:    fmt.Sprintf(GenericOpenApiErrorMsg, "", ""),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := HandleCommon(test.err)
			if (err != nil) != test.wantErr {
				t.Errorf("HandleCommon()\nerror: %v\nwantErr: %v", err, test.wantErr)
			}
			if err.Error() != test.want {
				t.Errorf("HandleCommon()\ngot: %s\nwant: %s", err, test.want)
			}
		})
	}
}

func TestDisplaySuggestionsMessage(t *testing.T) {
	suggestion := "This is a suggestion"
	err := NewErrorWithSuggestions("im an error hi", suggestion)
	require.Equal(t, fmt.Sprintf(wantSuggestionsMsgFormat, suggestion), DisplaySuggestionsMessage(err))
}

func TestCatchCCloudV2Error_RateLimited(t *testing.T) {
	for retryAfter, want := range map[string]string{
		"12":                            "Reduce the request rate and try again. The server asked to wait 12 seconds.",
		"Wed, 21 Oct 2026 07:28:00 GMT": "Reduce the request rate and try again.",
		"":                              "Reduce the request rate and try again.",
	} {
		r := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{retryAfter}},
			Body:       io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Too many requests"}]}`)),
		}
		err := CatchCCloudV2Error(fmt.Errorf("429 Too Many Requests"), r)
		require.EqualError(t, err, "Too many requests")
		withSuggestions, ok := err.(ErrorWithSuggestions)
		require.True(t, ok)
		require.True(t, strings.HasSuffix(withSuggestions.GetSuggestionsMsg(), want), retryAfter)
	}
}
