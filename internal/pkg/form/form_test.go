package form

import (
	"bytes"
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
}

func TestShow(t *testing.T) {
	field := Field{Prompt: "Username"}
	testShow(t, field, "Username: ")
}

func TestShowYesOrNo(t *testing.T) {
	field := Field{Prompt: "Ok?", IsYesOrNo: true}
	testShow(t, field, "Ok? (y/n): ")
}

func TestShowDefault(t *testing.T) {
	field := Field{Prompt: "Username", DefaultValue: "user"}
	testShow(t, field, "Username: (user) ")
}

func testShow(t *testing.T, field Field, output string) {
	command := new(cobra.Command)

	out := new(bytes.Buffer)
	command.SetOut(out)

	show(command, field)
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

func TestValidateYesOrNo(t *testing.T) {
	field := Field{IsYesOrNo: true}

	for _, val := range []string{"y", "yes"} {
		res, err := validate(field, val)
		require.NoError(t, err)
		require.True(t, res.(bool))
	}

	for _, val := range []string{"n", "no"} {
		res, err := validate(field, val)
		require.NoError(t, err)
		require.False(t, res.(bool))
	}

	_, err := validate(field, "maybe")
	require.Error(t, err)
}

func TestValidateDefaultVal(t *testing.T) {
	field := Field{DefaultValue: "default"}

	res, err := validate(field, "")
	require.Equal(t, "default", res)
	require.NoError(t, err)
}

func TestValidate(t *testing.T) {
	res, err := validate(Field{}, "res")
	require.Equal(t, "res", res)
	require.NoError(t, err)
}

func TestValidateRegexFail(t *testing.T) {
	_, err := validate(Field{Regex: `(?:[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`}, "milestodzo.com")
	require.Error(t, err)
}

func TestValidateRegexSuccess(t *testing.T) {
	res, err := validate(Field{Regex: `(?:[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`}, "mtodzo@confluent.io")
	require.Equal(t, "mtodzo@confluent.io", res)
	require.NoError(t, err)
}
