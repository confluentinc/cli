package controller

import (
	"net/http"
	"testing"

	"github.com/confluentinc/go-prompt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestInputController_renderError(t *testing.T) {
	type fields struct {
		History         History
		appController   *ApplicationController
		smartCompletion bool
		table           *TableController
		p               *prompt.Prompt
		store           StoreInterface
	}
	type args struct {
		err *StatementError
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &InputController{
				appController: tt.fields.appController,
			}
			require.Equal(t, tt.want, c.isSessionValid(tt.args.err))
		})
	}
}

func TestRenderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAppController := NewMockApplicationControllerInterface(ctrl)

	inputController := &InputController{appController: mockAppController}
	err := &StatementError{HttpResponseCode: http.StatusUnauthorized}

	// Test unauthorized error - should exit application
	mockAppController.EXPECT().ExitApplication().Times(1)
	result := inputController.isSessionValid(err)
	require.False(t, result)

	// Test other error
	err = &StatementError{Msg: "Something went wrong."}
	result = inputController.isSessionValid(err)
	require.True(t, result)
}
