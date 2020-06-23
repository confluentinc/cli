package errors

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewErrorf(t *testing.T) {
	format := "an error message with `%s`"
	param := "param"
	errMsg := fmt.Sprintf(format, param)

	tests := []struct {
		name            string
		initializerFunc func(string, ...interface{}) CLIDefinedError
		wantPrefix      string
	}{
		{
			name:            "NewBackendErrorf",
			initializerFunc: NewBackendErrorf,
			wantPrefix:      backendErrorPrefix,
		},
		{
			name:            "NewResourceValidationErrorf",
			initializerFunc: NewResourceValidationErrorf,
			wantPrefix:      resourceValidationErrorPrefix,
		},
		{
			name:            "NewProhibitedActionErrorf",
			initializerFunc: NewProhibitedActionErrorf,
			wantPrefix:      prohibitedActionErrorPrefix,
		},
		{
			name:            "NewFlagUseErrorf",
			initializerFunc: NewFlagUseErrorf,
			wantPrefix:      flagUseErrorPrefix,
		},
		{
			name:            "NewCLIStateErrorf",
			initializerFunc: NewCLIStateErrorf,
			wantPrefix:      cliStateErrorPrefix,
		},
		{
			name:            "NewCorruptedCLIConfigErrorf",
			initializerFunc: NewCorruptedCLIConfigErrorf,
			wantPrefix:      corruptedCLIConfigErrorPrefix,
		},
		{
			name:            "NewResourceNotReadyErrorf",
			initializerFunc: NewResourceNotReadyErrorf,
			wantPrefix:      resourceNotReadyErrorPrefix,
		},
		{
			name:            "NewUnexpectedCLIBehaviorErrorf",
			initializerFunc: NewUnexpectedCLIBehaviorErrorf,
			wantPrefix:      unexpectedCLIBehaviorErrorPrefix,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErrorMsg := fmt.Sprintf(errorFormat, tt.wantPrefix, errMsg)
			err := tt.initializerFunc(format, param)
			require.Error(t, err)
			require.Equal(t, wantErrorMsg, err.Error())
		})
	}
}

func TestNewErrorWrapf(t *testing.T) {
	wrappedFormat := "%s: %s"
	wrappedErr := fmt.Errorf("wrapped error")

	format := "an error message with `%s`"
	param := "param"

	errMsg := fmt.Sprintf(format, param)
	wrappedErrorMsg := fmt.Sprintf(wrappedFormat, errMsg, wrappedErr.Error())

	tests := []struct {
		name            string
		initializerFunc func(error, string, ...interface{}) CLIDefinedError
		wantPrefix      string
	}{
		{
			name:            "NewBackendErrorWrapf",
			initializerFunc: NewBackendErrorWrapf,
			wantPrefix:      backendErrorPrefix,
		},
		{
			name:            "NewResourceValidationErrorWrapf",
			initializerFunc: NewResourceValidationErrorWrapf,
			wantPrefix:      resourceValidationErrorPrefix,
		},
		{
			name:            "NewProhibitedActionErrorWrapf",
			initializerFunc: NewProhibitedActionErrorWrapf,
			wantPrefix:      prohibitedActionErrorPrefix,
		},
		{
			name:            "NewFlagUseErrorWrapf",
			initializerFunc: NewFlagUseErrorWrapf,
			wantPrefix:      flagUseErrorPrefix,
		},
		{
			name:            "NewCLIStateErrorWrapf",
			initializerFunc: NewCLIStateErrorWrapf,
			wantPrefix:      cliStateErrorPrefix,
		},
		{
			name:            "NewCorruptedCLIConfigErrorWrapf",
			initializerFunc: NewCorruptedCLIConfigErrorWrapf,
			wantPrefix:      corruptedCLIConfigErrorPrefix,
		},
		{
			name:            "NewResourceNotReadyErrorWrapf",
			initializerFunc: NewResourceNotReadyErrorWrapf,
			wantPrefix:      resourceNotReadyErrorPrefix,
		},
		{
			name:            "NewUnexpectedCLIBehaviorErrorWrapf",
			initializerFunc: NewUnexpectedCLIBehaviorErrorWrapf,
			wantPrefix:      unexpectedCLIBehaviorErrorPrefix,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantErrorMsg := fmt.Sprintf(errorFormat, tt.wantPrefix, wrappedErrorMsg)
			err := tt.initializerFunc(wrappedErr, format, param)
			require.Error(t, err)
			require.Equal(t, wantErrorMsg, err.Error())
		})
	}
}

func TestDirectionsMessage(t *testing.T) {
	errorMsgFormat := "an error message"

	msgFormat := "A directions message with `%s`."
	msgParam := "param"

	err := NewBackendErrorf(errorMsgFormat)
	err.SetDirectionsMsg(msgFormat, msgParam)

	wantDirectionsMsg := fmt.Sprintf(msgFormat, msgParam)
	require.Equal(t, wantDirectionsMsg, err.GetDirectionsMsg())

	var b bytes.Buffer
	HandleSuggestionsMessageDisplay(err, &b)
	out := b.String()
	wantDirectionsOutput := fmt.Sprintf(suggestionsMessageFormat, wantDirectionsMsg)
	require.Equal(t, wantDirectionsOutput, out)
}

