package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func LoadFixture(t *testing.T, fixture string) string {
	content, err := os.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), "fixtures", "output", fixture)
}

func getInputFixturePath(t *testing.T, directoryName string, file string) string {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}
	return filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", directoryName, file)
}
