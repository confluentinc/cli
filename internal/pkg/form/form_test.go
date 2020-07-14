package form

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrompt(t *testing.T) {
	f := New(
		Field{ID: "username", Prompt: "Username"},
		Field{ID: "password", Prompt: "Password", IsHidden: true},
	)

	r := bufio.NewReader(strings.NewReader("user\npass\n"))
	w := bufio.NewWriter(new(bytes.Buffer))

	err := f.Prompt(r, w)
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
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	err := show(w, field, savedValue)
	require.NoError(t, err)
	require.Equal(t, output, buf.String())
}

func TestRead(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("user\n"))
	username, _ := read(in, Field{})
	require.Equal(t, "user", username)
}

func TestReadToEOF(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("user"))
	username, _ := read(in, Field{})
	require.Equal(t, "user", username)
}

func TestSave(t *testing.T) {
	res, err := save(Field{}, "res")
	require.NoError(t, err)
	require.Equal(t, "res", res)
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

func TestSaveDefaultValue(t *testing.T) {
	field := Field{DefaultValue: "default"}

	res, err := save(field, "")
	require.NoError(t, err)
	require.Equal(t, "default", res)
}
