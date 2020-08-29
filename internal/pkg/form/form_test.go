package form

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
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
}

func TestShow(t *testing.T) {
	field := Field{Prompt: "Username"}
	testShow(t, field, nil, "Username: ")
}

func TestShowYesOrNo(t *testing.T) {
	field := Field{Prompt: "Ok?", IsYesOrNo: true}
	testShow(t, field, nil, "Ok? (y/n): ")
}

func TestShowDefault(t *testing.T) {
	field := Field{Prompt: "Username", DefaultValue: "user"}
	testShow(t, field, nil, "Username: (user) ")
}

func TestShowSavedValue(t *testing.T) {
	field := Field{Prompt: "Username"}
	testShow(t, field, "user", "Username: (user) ")
}

func testShow(t *testing.T, field Field, savedValue interface{}, output string) {
	command := new(cobra.Command)

	out := new(bytes.Buffer)
	command.SetOut(out)

	show(command, field, savedValue)
	require.Equal(t, output, out.String())
}

func TestRead(t *testing.T) {
	prompt := &mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return "user", nil
		},
	}

	username, _ := read(Field{}, prompt)
	require.Equal(t, "user", username)
}

func TestReadPassword(t *testing.T) {
	field := Field{IsHidden: true}

	prompt := &mock.Prompt{
		ReadLineMaskedFunc: func() (string, error) {
			return "pass", nil
		},
	}

	password, _ := read(field, prompt)
	require.Equal(t, "pass", password)
}

func TestSaveYesOrNo(t *testing.T) {
	field := Field{IsYesOrNo: true}

	for _, val := range []string{"y", "yes"} {
		res, err := save(field, val)
		require.NoError(t, err)
		require.True(t, res.(bool))
	}

	for _, val := range []string{"n", "no"} {
		res, err := save(field, val)
		require.NoError(t, err)
		require.False(t, res.(bool))
	}

	_, err := save(field, "maybe")
	require.Error(t, err)
}

func TestSaveDefaultVal(t *testing.T) {
	field := Field{DefaultValue: "default"}

	res, err := save(field, "")
	require.Equal(t, "default", res)
	require.NoError(t, err)
}

func TestSave(t *testing.T) {
	res, err := save(Field{}, "res")
	require.Equal(t, "res", res)
	require.NoError(t, err)
}
