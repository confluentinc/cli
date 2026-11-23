package httpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	perrors "github.com/confluentinc/cli/v4/pkg/errors"
)

func TestRetryError(t *testing.T) {
	cause := errors.New("connection reset by peer")

	tests := []struct {
		name            string
		err             *RetryError
		wantMessage     string
		wantSuggestions string
	}{
		{
			name:            "retries ran out",
			err:             &RetryError{Attempts: 5, Err: cause},
			wantMessage:     "request failed after 5 attempts: connection reset by peer",
			wantSuggestions: "Check your network connection and proxy settings, then try again.",
		},
		{
			name:            "maybe applied, one attempt",
			err:             &RetryError{Attempts: 1, Err: cause, MaybeApplied: true},
			wantMessage:     "request failed but may have reached the server: connection reset by peer",
			wantSuggestions: "If this command makes a change, check whether it was applied, for example with the matching `list` or `describe` command, before trying again.",
		},
		{
			name:            "maybe applied, several attempts",
			err:             &RetryError{Attempts: 2, Err: cause, MaybeApplied: true},
			wantMessage:     "request failed after 2 attempts but may have reached the server: connection reset by peer",
			wantSuggestions: "If this command makes a change, check whether it was applied, for example with the matching `list` or `describe` command, before trying again.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.wantMessage, test.err.Error())
			require.ErrorIs(t, test.err, cause)

			var withSuggestions perrors.ErrorWithSuggestions
			require.ErrorAs(t, test.err, &withSuggestions)
			require.Equal(t, test.wantSuggestions, withSuggestions.GetSuggestionsMsg())
		})
	}
}
