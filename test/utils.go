package test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

// NormalizeNewLines replaces \r\n and \r newline sequences with \n
func NormalizeNewLines(raw string) string {
	normalized := bytes.Replace([]byte(raw), []byte{13, 10}, []byte{10}, -1)
	normalized = bytes.Replace(normalized, []byte{13}, []byte{10}, -1)
	return string(normalized)
}

func LoadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(FixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func FixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), "fixtures", "output", fixture)
}
