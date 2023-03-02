package form

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/mock"
)

func TestPrompt(t *testing.T) {
	f := New(
		Field{ID: "username", Prompt: "Username"},
		Field{ID: "password", Prompt: "Password", IsHidden: true},
	)

	command := &cobra.Command{}
	command.SetOut(new(bytes.Buffer))

	prompt := &mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return "user", nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			return "pass", nil
		},
	}

	err := f.Prompt(command, prompt)
	require.NoError(t, err)
	require.Equal(t, "user", f.Responses["username"].(string))
	require.Equal(t, "pass", f.Responses["password"].(string))

	// Format the test report correctly
	fmt.Println()
}
