package controller

import (
	"net/http"
	"testing"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestInputController_renderError(t *testing.T) {
	type fields struct {
		History       history.History
		appController *ApplicationController
	}
	type args struct {
		err *types.StatementError
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
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	inputController := &InputController{appController: mockAppController}
	err := &types.StatementError{HttpResponseCode: http.StatusUnauthorized}

	// Test unauthorized error - should exit application
	mockAppController.EXPECT().ExitApplication().Times(1)
	result := inputController.isSessionValid(err)
	require.False(t, result)

	// Test other error
	err = &types.StatementError{Msg: "Something went wrong."}
	result = inputController.isSessionValid(err)
	require.True(t, result)
}

func TestShouldUseTView(t *testing.T) {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      bool
	}{
		{
			name:      "local statement should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "local statement should not use TView even if unbounded",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "non-local unbounded statement should always use TView",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      true,
		},
		{
			name:      "statement with no results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      false,
		},
		{
			name: "statement with one column and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}, {Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: false,
		},
		{
			name: "statement with two columns and one row should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: false,
		},
		{
			name: "statement with two columns and two rows should use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}, {Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: true,
		},
		{
			name: "local statement with two columns and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}, {Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldUseTView(tt.statement))
		})
	}
}
