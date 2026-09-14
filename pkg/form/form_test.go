package form

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/confluentinc/cli/v4/pkg/mock"
)

func TestPrompt(t *testing.T) {
	f := New(
		Field{ID: "username", Prompt: "Username"},
		Field{ID: "password", Prompt: "Password", IsHidden: true},
	)

	ctrl := gomock.NewController(t)
	prompt := mock.NewMockPrompt(ctrl)
	prompt.EXPECT().ReadLine().Return("user", nil).AnyTimes()
	prompt.EXPECT().ReadLineMasked().Return("pass", nil).AnyTimes()

	err := f.Prompt(prompt)
	require.NoError(t, err)
	require.Equal(t, "user", f.Responses["username"].(string))
	require.Equal(t, "pass", f.Responses["password"].(string))

	// Format the test report correctly
	fmt.Println()
}
