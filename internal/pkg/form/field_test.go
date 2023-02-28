package form

import (
	"testing"

	"github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	var field Field

	prompt := &mock.Prompt{
		ReadLineFunc: func() (string, error) {
			return "user", nil
		},
	}

	username, _ := field.read(prompt)
	require.Equal(t, "user", username)
}

func TestRead_Password(t *testing.T) {
	field := Field{IsHidden: true}

	prompt := &mock.Prompt{
		ReadLineMaskedFunc: func() (string, error) {
			return "pass", nil
		},
	}

	password, _ := field.read(prompt)
	require.Equal(t, "pass", password)
}

func TestValidate_YesOrNo(t *testing.T) {
	field := Field{IsYesOrNo: true}

	for _, val := range []string{"y", "yes"} {
		res, err := field.validate(val)
		assert.NoError(t, err)
		assert.True(t, res.(bool))
	}

	for _, val := range []string{"n", "no"} {
		res, err := field.validate(val)
		assert.NoError(t, err)
		assert.False(t, res.(bool))
	}

	_, err := field.validate("maybe")
	require.Error(t, err)
}

func TestValidate_DefaultVal(t *testing.T) {
	field := Field{DefaultValue: "default"}

	res, err := field.validate("")
	require.Equal(t, "default", res)
	require.NoError(t, err)
}

func TestValidate(t *testing.T) {
	var field Field

	res, err := field.validate("res")
	require.Equal(t, "res", res)
	require.NoError(t, err)
}

func TestValidate_RegexFail(t *testing.T) {
	field := Field{Regex: `(?:[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`}
	_, err := field.validate("milestodzo.com")
	require.Error(t, err)
}

func TestValidate_RegexSuccess(t *testing.T) {
	field := Field{Regex: `(?:[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+(?:\.[a-z0-9!#$%&'*+\/=?^_\x60{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`}
	res, err := field.validate("mtodzo@confluent.io")
	require.Equal(t, "mtodzo@confluent.io", res)
	require.NoError(t, err)
}

func TestString(t *testing.T) {
	field := Field{Prompt: "Username"}
	require.Equal(t, "Username: ", field.String())
}

func TestString_YesOrNo(t *testing.T) {
	field := Field{Prompt: "Ok?", IsYesOrNo: true}
	require.Equal(t, "Ok? (y/n): ", field.String())
}

func TestString_DefaultValue(t *testing.T) {
	field := Field{Prompt: "Username", DefaultValue: "user"}
	require.Equal(t, "Username: (user) ", field.String())
}
