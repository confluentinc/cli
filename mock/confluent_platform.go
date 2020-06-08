package mock

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var count = 0

type ConfluentPlatform struct {
	ConfluentHome       string
	ConfluentCurrent    string
	ConfluentCurrentDir string
}

func NewConfluentPlatform() *ConfluentPlatform {
	return new(ConfluentPlatform)
}

func (cp *ConfluentPlatform) NewConfluentHome() error {
	dir, err := newTestDir()
	cp.ConfluentHome = dir
	if err != nil {
		return err
	}

	return os.Setenv("CONFLUENT_HOME", dir)
}

func (cp *ConfluentPlatform) NewConfluentCurrent() error {
	dir, err := newTestDir()
	cp.ConfluentCurrent = dir
	if err != nil {
		return err
	}

	return os.Setenv("CONFLUENT_CURRENT", dir)
}

func (cp *ConfluentPlatform) NewConfluentCurrentDir() error {
	if err := cp.NewConfluentCurrent(); err != nil {
		return err
	}

	dir, err := newTestDir()
	cp.ConfluentCurrentDir = dir
	if err != nil {
		return err
	}

	trackingFile := filepath.Join(cp.ConfluentCurrent, "confluent.current")
	return ioutil.WriteFile(trackingFile, []byte(dir), 0644)
}

func newTestDir() (string, error) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("confluent.test.%06d", count))
	count++

	return path, os.Mkdir(path, 0777)
}

func (cp *ConfluentPlatform) AddScriptToConfluentHome(file string, contents string) error {
	return cp.AddFileToConfluentHome(file, contents, 0755)
}

func (cp *ConfluentPlatform) AddEmptyFileToConfluentHome(file string) error {
	return cp.AddFileToConfluentHome(file, "", 0644)
}

func (cp *ConfluentPlatform) AddFileToConfluentHome(file string, contents string, perm os.FileMode) error {
	path := filepath.Join(cp.ConfluentHome, file)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(contents), perm)
}

func (cp *ConfluentPlatform) TearDown() {
	os.RemoveAll(cp.ConfluentHome)
	os.RemoveAll(cp.ConfluentCurrent)
	os.RemoveAll(cp.ConfluentCurrentDir)
}
